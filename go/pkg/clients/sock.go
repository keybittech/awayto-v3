package clients

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/proto"

	"github.com/google/uuid"
)

const (
	CreateSocketTicketSocketCommand = iota
	DeleteSocketTicketSocketCommand
	CreateSocketConnectionSocketCommand
	DeleteSocketConnectionSocketCommand
	GetSubscriberSocketCommand
	SendSocketMessageSocketCommand
	AddSubscribedTopicSocketCommand
	GetSubscribedTargetsSocketCommand
	DeleteSubscribedTopicSocketCommand
	HasSubscribedTopicSocketCommand
)

type Subscribers map[string]*types.Subscriber
type Connections map[string]net.Conn

type SocketMaps struct {
	subscribers Subscribers
	connections Connections
	mu          sync.RWMutex
}

func NewSocketMaps() *SocketMaps {
	return &SocketMaps{
		subscribers: make(Subscribers),
		connections: make(Connections),
	}
}

var (
	CID_LENGTH = 36
)

type Socket struct {
	interfaces.ISocket
	handlerId string
}

type ConnectedSocket struct {
	Conn net.Conn
	Sock *Socket
}

const sockHandlerId = "sock"

func InitSocket() *Socket {

	InitGlobalWorkerPool(4, 8)

	socketMaps := NewSocketMaps()

	GetGlobalWorkerPool().RegisterProcessFunction(sockHandlerId, func(sockCmd CombinedCommand) bool {

		cmd, ok := sockCmd.(interfaces.SocketCommand)
		if !ok {
			return false
		}

		defer func(replyChan chan interfaces.SocketResponse) {
			if r := recover(); r != nil {
				err := errors.New(fmt.Sprintf("Did recover from %+v", r))
				replyChan <- interfaces.SocketResponse{Error: err}
			}
			close(replyChan)
		}(cmd.ReplyChan)

		socketMaps.mu.Lock()
		defer socketMaps.mu.Unlock()

		switch cmd.Ty {
		case CreateSocketTicketSocketCommand:
			auth := uuid.NewString()
			connectionId := uuid.NewString()
			ticket := auth + ":" + connectionId

			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				subscriber.Tickets[auth] = connectionId
			} else {
				subscriber = &types.Subscriber{
					UserSub:          cmd.Request.UserSub,
					GroupId:          cmd.Request.GroupId,
					Roles:            cmd.Request.Roles,
					Tickets:          map[string]string{auth: connectionId},
					SubscribedTopics: make(map[string]string),
				}
			}
			socketMaps.subscribers[cmd.Request.UserSub] = subscriber

			cmd.ReplyChan <- interfaces.SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					Ticket: ticket,
				},
			}

		case CreateSocketConnectionSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				auth, connId, err := util.SplitSocketId(cmd.Request.Ticket)
				if err != nil {
					cmd.ReplyChan <- interfaces.SocketResponse{
						Error: err,
					}
					break
				}

				if _, ok := subscriber.Tickets[auth]; !ok {
					cmd.ReplyChan <- interfaces.SocketResponse{
						Error: errors.New("invalid ticket"),
					}
					break
				}

				delete(subscriber.Tickets, auth)
				subscriber.ConnectionIds += connId

				socketMaps.connections[connId] = cmd.Request.Conn

				cmd.ReplyChan <- interfaces.SocketResponse{
					SocketResponseParams: &types.SocketResponseParams{
						Subscriber: proto.Clone(subscriber).(*types.Subscriber),
					},
				}
			} else {
				cmd.ReplyChan <- interfaces.SocketResponse{
					Error: errors.New("no sub found in sock"),
				}
			}

		case DeleteSocketConnectionSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				_, connId, _ := util.SplitSocketId(cmd.Request.Ticket)

				delete(socketMaps.connections, connId)

				connIdStartIdx := strings.Index(subscriber.ConnectionIds, connId)
				connIdEndIdx := connIdStartIdx + CID_LENGTH // uuid length
				if connIdStartIdx == -1 || len(subscriber.ConnectionIds) < connIdEndIdx {
					cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("connection id not found to delete")}
					break
				}
				subscriber.ConnectionIds = subscriber.ConnectionIds[:connIdStartIdx] + subscriber.ConnectionIds[connIdEndIdx:]
				if len(subscriber.ConnectionIds) == 0 {
					delete(socketMaps.subscribers, cmd.Request.UserSub)
				}
			}
			cmd.ReplyChan <- interfaces.SocketResponse{}
		case GetSubscriberSocketCommand:
			auth, connId, err := util.SplitSocketId(cmd.Request.Ticket)
			if err != nil {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: err}
				break
			}

			var found bool

			for _, subscriber := range socketMaps.subscribers {
				if _, ok := subscriber.Tickets[auth]; ok || strings.Index(subscriber.ConnectionIds, connId) != -1 {
					found = true
					cmd.ReplyChan <- interfaces.SocketResponse{
						SocketResponseParams: &types.SocketResponseParams{
							Subscriber: proto.Clone(subscriber).(*types.Subscriber),
						},
						Error: nil,
					}
					break
				}
			}

			if !found {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("no subscriber found for ticket")}
			}

		case SendSocketMessageSocketCommand:
			var sentAtLeastOne bool
			var sendErr error
			var attemptedTargets string
			for i := 0; i+CID_LENGTH <= len(cmd.Request.Targets); i += CID_LENGTH {
				connId := cmd.Request.Targets[i : i+CID_LENGTH]
				if strings.Index(attemptedTargets, connId) == -1 {
					if conn, ok := socketMaps.connections[connId]; ok {
						sendErr = util.WriteSocketConnectionMessage(cmd.Request.MessageBytes, conn)
						if sendErr != nil {
							continue
						}
						sentAtLeastOne = true
					}

					attemptedTargets += connId
				}
			}
			if !sentAtLeastOne {
				sendErr = errors.New(fmt.Sprintf("no targets to send to"))
			}
			cmd.ReplyChan <- interfaces.SocketResponse{
				Error: sendErr,
			}

		case AddSubscribedTopicSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if !ok {
				cmd.ReplyChan <- interfaces.SocketResponse{
					Error: errors.New("no subscriber found to add topic"),
				}
				break
			}

			for i := 0; i+CID_LENGTH <= len(cmd.Request.Targets); i += CID_LENGTH {
				connId := cmd.Request.Targets[i : i+CID_LENGTH]
				if strings.Index(subscriber.SubscribedTopics[cmd.Request.Topic], connId) == -1 {
					subscriber.SubscribedTopics[cmd.Request.Topic] += connId
				}
			}
			cmd.ReplyChan <- interfaces.SocketResponse{}

		case GetSubscribedTargetsSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				cmd.ReplyChan <- interfaces.SocketResponse{
					SocketResponseParams: &types.SocketResponseParams{
						Targets: subscriber.ConnectionIds,
					},
				}
			} else {
				cmd.ReplyChan <- interfaces.SocketResponse{
					Error: errors.New("subscriber not found"),
				}
			}

		case DeleteSubscribedTopicSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				delete(subscriber.SubscribedTopics, cmd.Request.Topic)
			}
			cmd.ReplyChan <- interfaces.SocketResponse{}

		case HasSubscribedTopicSocketCommand:
			hasSub := false
			subscriber, okSub := socketMaps.subscribers[cmd.Request.UserSub]
			if okSub {
				if _, okTopic := subscriber.SubscribedTopics[cmd.Request.Topic]; okTopic {
					hasSub = true
				}
			}
			cmd.ReplyChan <- interfaces.SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					HasSub: hasSub,
				},
			}

		default:
			log.Fatal("unknown command type", cmd.Ty)
		}

		return true
	})

	return &Socket{handlerId: sockHandlerId}
}

func (s *Socket) RouteCommand(cmd interfaces.SocketCommand) error {
	return GetGlobalWorkerPool().RouteCommand(cmd)
}

func (s *Socket) Close() {
	GetGlobalWorkerPool().UnregisterProcessFunction(s.handlerId)
}

func (s *Socket) SendCommand(cmdType int32, request interfaces.SocketRequest) (interfaces.SocketResponse, error) {
	createCmd := func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand {
		return interfaces.SocketCommand{
			SocketCommandParams: &types.SocketCommandParams{
				Ty:       cmdType,
				ClientId: request.UserSub,
			},
			Request:   request,
			ReplyChan: replyChan,
		}
	}

	return SendCommand(s, createCmd)
}

func (s *Socket) GetSocketTicket(session *types.UserSession) (string, error) {
	response, err := s.SendCommand(CreateSocketTicketSocketCommand, interfaces.SocketRequest{
		SocketRequestParams: &types.SocketRequestParams{
			UserSub: session.UserSub,
			GroupId: session.GroupId,
			Roles:   strings.Join(session.AvailableUserGroupRoles, " "),
		},
	})

	if err = ChannelError(err, response.Error); err != nil {
		return "", util.ErrCheck(err)
	}

	return response.Ticket, nil
}

func (s *Socket) SendMessageBytes(userSub, targets string, messageBytes []byte) error {
	if targets == "" {
		return util.ErrCheck(errors.New("no targets provided"))
	}
	if len(userSub) == 0 {
		return util.ErrCheck(errors.New("user sub required to send a message"))
	}

	response, err := s.SendCommand(SendSocketMessageSocketCommand, interfaces.SocketRequest{
		SocketRequestParams: &types.SocketRequestParams{
			UserSub:      userSub,
			Targets:      targets,
			MessageBytes: messageBytes,
		},
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (s *Socket) SendMessage(userSub, targets string, message *types.SocketMessage) error {
	if message == nil {
		return util.ErrCheck(errors.New("message required"))
	}
	return s.SendMessageBytes(userSub, targets, util.GenerateMessage(util.DefaultPadding, message))
}

func (s *Socket) RoleCall(userSub string) error {
	response, err := s.SendCommand(GetSubscribedTargetsSocketCommand, interfaces.SocketRequest{
		SocketRequestParams: &types.SocketRequestParams{
			UserSub: userSub,
		},
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	if len(response.Targets) > 0 {
		err := s.SendMessage(userSub, response.Targets, &types.SocketMessage{Action: types.SocketActions_ROLE_CALL})
		if err != nil {
			return util.ErrCheck(response.Error)
		}
	}

	return nil
}
