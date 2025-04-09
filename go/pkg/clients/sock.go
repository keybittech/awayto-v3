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

	"github.com/google/uuid"
)

const (
	CreateSocketTicketSocketCommand = iota
	DeleteSocketTicketSocketCommand
	CreateSocketConnectionSocketCommand
	DeleteSocketConnectionSocketCommand
	SendSocketMessageSocketCommand
	AddSubscribedTopicSocketCommand
	GetSubscribedTargetsSocketCommand
	DeleteSubscribedTopicSocketCommand
	HasSubscribedTopicSocketCommand
)

type Subscribers map[string]*types.Subscriber
type Connections map[string]net.Conn
type AuthSubscribers map[string]string

type SocketMaps struct {
	subscribers     Subscribers
	connections     Connections
	authSubscribers AuthSubscribers
	mu              sync.RWMutex
}

func NewSocketMaps() *SocketMaps {
	return &SocketMaps{
		subscribers:     make(Subscribers),
		connections:     make(Connections),
		authSubscribers: make(AuthSubscribers),
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

var emptySocketResponse = interfaces.SocketResponse{}

var (
	authSubscriberNotFound    = errors.New("auth subscriber not found")
	noSubscriberForAuth       = errors.New("no subscriber for auth")
	invalidTicket             = errors.New("invalid ticket")
	invalidTicketConnectionId = errors.New("invalid ticket conn id")
	noDeletionConnectionId    = errors.New("connection id not found to delete")
	noSubFoundInSock          = errors.New("no sub found in sock")
	noTargetsToSend           = errors.New("no targets to send to")
	noSubscriberToAddTopic    = errors.New("no subscriber found to add topic")
	noSubscriberTargets       = errors.New("no subscriber targets")
)

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
				err := errors.New("socket channel recovery: " + fmt.Sprint(r))
				replyChan <- interfaces.SocketResponse{Error: err}
			}
			close(replyChan)
		}(cmd.ReplyChan)

		socketMaps.mu.Lock()
		defer socketMaps.mu.Unlock()

		switch cmd.Ty {
		case CreateSocketTicketSocketCommand:
			auth := uuid.NewString()
			socketMaps.authSubscribers[auth] = cmd.Request.UserSub

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
					ConnectionId:     connectionId,
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
			auth, _, err := util.SplitColonJoined(cmd.Request.Ticket)
			if err != nil {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: err}
				break
			}

			userSub, ok := socketMaps.authSubscribers[auth]
			if !ok {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: authSubscriberNotFound}
				break
			}

			subscriber, ok := socketMaps.subscribers[userSub]
			if ok {
				auth, connId, err := util.SplitColonJoined(cmd.Request.Ticket)
				if err != nil {
					cmd.ReplyChan <- interfaces.SocketResponse{
						Error: err,
					}
					break
				}

				if _, ok := subscriber.Tickets[auth]; !ok {
					cmd.ReplyChan <- interfaces.SocketResponse{
						Error: invalidTicket,
					}
					break
				}

				delete(subscriber.Tickets, auth)
				delete(socketMaps.authSubscribers, auth)

				if subscriber.ConnectionId != connId {
					cmd.ReplyChan <- interfaces.SocketResponse{
						Error: invalidTicketConnectionId,
					}
					break
				}

				subscriber.ConnectionIds += connId

				socketMaps.connections[connId] = cmd.Request.Conn

				cmd.ReplyChan <- interfaces.SocketResponse{
					SocketResponseParams: &types.SocketResponseParams{
						UserSub: subscriber.UserSub,
						GroupId: subscriber.GroupId,
						Roles:   subscriber.Roles,
					},
				}
			} else {
				cmd.ReplyChan <- interfaces.SocketResponse{
					Error: noSubFoundInSock,
				}
			}

		case DeleteSocketConnectionSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				println("deleting socket connection for", cmd.Request.ConnId)
				delete(socketMaps.connections, cmd.Request.ConnId)

				connIdStartIdx := strings.Index(subscriber.ConnectionIds, cmd.Request.ConnId)
				connIdEndIdx := connIdStartIdx + CID_LENGTH // uuid length
				if connIdStartIdx == -1 || len(subscriber.ConnectionIds) < connIdEndIdx {
					cmd.ReplyChan <- interfaces.SocketResponse{
						Error: noDeletionConnectionId,
					}
					break
				}
				subscriber.ConnectionIds = subscriber.ConnectionIds[:connIdStartIdx] + subscriber.ConnectionIds[connIdEndIdx:]
				println("subscriber", cmd.Request.UserSub, "got new connids", subscriber.ConnectionIds)
				if len(subscriber.ConnectionIds) == 0 {
					println("deleted subscriber", cmd.Request.UserSub, "with 0 connid")
					delete(socketMaps.subscribers, cmd.Request.UserSub)
				}
			}
			cmd.ReplyChan <- emptySocketResponse

		case SendSocketMessageSocketCommand:
			println("\n\n============== Send event ================")
			var sentAtLeastOne bool
			var sendErr error
			var attemptedTargets string

			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if !ok {
				cmd.ReplyChan <- interfaces.SocketResponse{
					Error: errors.New("bad subscriber during logging send message"),
				}
				break
			}

			println("user sub:", subscriber.UserSub)
			println("conn id:", subscriber.ConnectionId, "is trying to send to", cmd.Request.Targets, string(cmd.Request.MessageBytes))
			var connectionIds string
			for k := range socketMaps.connections {
				connectionIds += k + " "
			}
			println("tar len", len(cmd.Request.Targets))
			for i := 0; i+CID_LENGTH <= len(cmd.Request.Targets); i += CID_LENGTH {
				connId := cmd.Request.Targets[i : i+CID_LENGTH]
				println("checking connid", connId)
				println("current attempted", attemptedTargets)
				if strings.Index(attemptedTargets, connId) == -1 {
					println("attempting to check", connId, "with current ids being", connectionIds)
					if conn, ok := socketMaps.connections[connId]; ok {
						println("did match connection")
						sendErr = util.WriteSocketConnectionMessage(cmd.Request.MessageBytes, conn)
						if sendErr != nil {
							println("FAILED WITH WRITE ERROR")
							continue
						}
						println("did send success")
						sentAtLeastOne = true
					}

					attemptedTargets += connId
				}
			}
			if !sentAtLeastOne && sendErr == nil {
				println("FAILED WITH SEND ONE ERROR")
				sendErr = noTargetsToSend
			}
			cmd.ReplyChan <- interfaces.SocketResponse{
				Error: sendErr,
			}
			println("============== End Send event ================\n")

		case AddSubscribedTopicSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if !ok {
				cmd.ReplyChan <- interfaces.SocketResponse{
					Error: noSubscriberToAddTopic,
				}
				break
			}

			for i := 0; i+CID_LENGTH <= len(cmd.Request.Targets); i += CID_LENGTH {
				connId := cmd.Request.Targets[i : i+CID_LENGTH]
				if strings.Index(subscriber.SubscribedTopics[cmd.Request.Topic], connId) == -1 {
					subscriber.SubscribedTopics[cmd.Request.Topic] += connId
				}
			}
			cmd.ReplyChan <- emptySocketResponse

		case GetSubscribedTargetsSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if !ok {
				cmd.ReplyChan <- interfaces.SocketResponse{
					Error: noSubscriberTargets,
				}
			}

			cmd.ReplyChan <- interfaces.SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					Targets: subscriber.ConnectionIds,
				},
			}

		case DeleteSubscribedTopicSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				delete(subscriber.SubscribedTopics, cmd.Request.Topic)
			}
			cmd.ReplyChan <- emptySocketResponse

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

var socketCommandMustHaveSub = errors.New("socket command request must contain a user sub")

func (s *Socket) SendCommand(cmdType int32, request *types.SocketRequestParams, conn ...net.Conn) (interfaces.SocketResponse, error) {
	if request.UserSub == "" {
		return emptySocketResponse, socketCommandMustHaveSub
	}
	var userConn net.Conn
	if len(conn) > 0 {
		userConn = conn[0]
	}
	createCmd := func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand {
		return interfaces.SocketCommand{
			SocketCommandParams: &types.SocketCommandParams{
				Ty:       cmdType,
				ClientId: request.UserSub,
			},
			Request: interfaces.SocketRequest{
				SocketRequestParams: request,
				Conn:                userConn,
			},
			ReplyChan: replyChan,
		}
	}

	return SendCommand(s, createCmd)
}

func (s *Socket) GetSocketTicket(session *types.UserSession) (string, error) {
	response, err := s.SendCommand(CreateSocketTicketSocketCommand, &types.SocketRequestParams{
		UserSub: session.UserSub,
		GroupId: session.GroupId,
		Roles:   strings.Join(session.AvailableUserGroupRoles, " "),
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

	response, err := s.SendCommand(SendSocketMessageSocketCommand, &types.SocketRequestParams{
		UserSub:      userSub,
		Targets:      targets,
		MessageBytes: messageBytes,
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
	response, err := s.SendCommand(GetSubscribedTargetsSocketCommand, &types.SocketRequestParams{
		UserSub: userSub,
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
