package clients

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	GetGroupTargetsSocketCommand
	DeleteSubscribedTopicSocketCommand
	HasSubscribedTopicSocketCommand
)

const (
	CID_LENGTH    = 36
	sockHandlerId = "sock"
)

var (
	emptySocketResponse       = SocketResponse{}
	authSubscriberNotFound    = errors.New("auth subscriber not found")
	invalidTicket             = errors.New("invalid ticket")
	invalidTicketConnectionId = errors.New("invalid ticket conn id")
	messageRequired           = errors.New("message required")
	noDeletionConnectionId    = errors.New("connection id not found to delete")
	noSubFoundInSock          = errors.New("no sub found in sock")
	noSubscriberForAuth       = errors.New("no subscriber for auth")
	noSubscriberToAddTopic    = errors.New("no subscriber found to add topic")
	noSubscriberTargets       = errors.New("no subscriber targets")
	noTargetsToSend           = errors.New("no targets to send to")
	noTargetsProvided         = errors.New("no targets provided")
	socketCommandMustHaveSub  = errors.New("socket command request must contain a user sub")
	subRequiredToSend         = errors.New("user sub required to send a message")
)

type Subscribers map[string]*types.Subscriber
type Connections map[string]*websocket.Conn
type GroupTargets map[string]string
type AuthSubscribers map[string]string

type SocketMaps struct {
	subscribers     Subscribers
	connections     Connections
	groupTargets    GroupTargets
	authSubscribers AuthSubscribers
	mu              sync.RWMutex
}

func NewSocketMaps() *SocketMaps {
	return &SocketMaps{
		subscribers:     make(Subscribers),
		connections:     make(Connections),
		groupTargets:    make(GroupTargets),
		authSubscribers: make(AuthSubscribers),
	}
}

type Socket struct {
	handlerId  string
	SocketMaps *SocketMaps
}

func InitSocket() *Socket {

	InitGlobalWorkerPool(4, 8)

	socketMaps := NewSocketMaps()

	GetGlobalWorkerPool().RegisterProcessFunction(sockHandlerId, func(sockCmd CombinedCommand) bool {

		cmd, ok := sockCmd.(SocketCommand)
		if !ok {
			return false
		}

		defer func(replyChan chan SocketResponse) {
			if r := recover(); r != nil {
				err := errors.New("socket channel recovery: " + fmt.Sprint(r))
				replyChan <- SocketResponse{Error: err}
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
				subscriber.ConnectionId = connectionId
				subscriber.Tickets[auth] = connectionId
			} else {
				subscriber = &types.Subscriber{
					UserSub:          cmd.Request.UserSub,
					GroupId:          cmd.Request.GroupId,
					RoleBits:         cmd.Request.RoleBits,
					ConnectionId:     connectionId,
					Tickets:          map[string]string{auth: connectionId},
					SubscribedTopics: make(map[string]string),
				}
			}
			socketMaps.subscribers[cmd.Request.UserSub] = subscriber

			cmd.ReplyChan <- SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					Ticket: ticket,
				},
			}

		case CreateSocketConnectionSocketCommand:
			auth, _, err := util.SplitColonJoined(cmd.Request.Ticket)
			if err != nil {
				cmd.ReplyChan <- SocketResponse{
					Error: err,
				}
				break
			}

			userSub, ok := socketMaps.authSubscribers[auth]
			if !ok {
				cmd.ReplyChan <- SocketResponse{
					Error: authSubscriberNotFound,
				}
				break
			}

			subscriber, ok := socketMaps.subscribers[userSub]
			if ok {
				auth, connId, err := util.SplitColonJoined(cmd.Request.Ticket)
				if err != nil {
					cmd.ReplyChan <- SocketResponse{
						Error: err,
					}
					break
				}

				if _, ok := subscriber.Tickets[auth]; !ok {
					cmd.ReplyChan <- SocketResponse{
						Error: invalidTicket,
					}
					break
				}

				socketMaps.groupTargets[subscriber.GroupId] += connId

				delete(subscriber.Tickets, auth)
				delete(socketMaps.authSubscribers, auth)

				if subscriber.ConnectionId != connId {
					cmd.ReplyChan <- SocketResponse{
						Error: invalidTicketConnectionId,
					}
					break
				}

				subscriber.ConnectionIds += connId

				socketMaps.connections[connId] = cmd.Request.Conn

				cmd.ReplyChan <- SocketResponse{
					SocketResponseParams: &types.SocketResponseParams{
						UserSub:  subscriber.UserSub,
						GroupId:  subscriber.GroupId,
						RoleBits: subscriber.RoleBits,
					},
				}
			} else {
				cmd.ReplyChan <- SocketResponse{
					Error: noSubFoundInSock,
				}
			}

		case DeleteSocketConnectionSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if ok {
				// println("deleting socket connection for", cmd.Request.ConnId)
				delete(socketMaps.connections, cmd.Request.ConnId)

				subscriber.ConnectionIds = strings.Replace(subscriber.ConnectionIds, cmd.Request.ConnId, "", 1)
				socketMaps.groupTargets[cmd.Request.ConnId] = strings.Replace(socketMaps.groupTargets[cmd.Request.GroupId], cmd.Request.ConnId, "", 1)

				// println("subscriber", cmd.Request.UserSub, "got new connids", subscriber.ConnectionIds)
				if len(subscriber.ConnectionIds) == 0 {
					// println("deleted subscriber", cmd.Request.UserSub, "with 0 connid")
					delete(socketMaps.subscribers, cmd.Request.UserSub)
				}
			}
			cmd.ReplyChan <- emptySocketResponse

		case SendSocketMessageSocketCommand:
			// println("\n\n============== Send event ================")
			var sentAtLeastOne bool
			var sendErr error
			var attemptedTargets string

			// _, ok := socketMaps.subscribers[cmd.Request.UserSub]
			// if !ok {
			// 	cmd.ReplyChan <- SocketResponse{
			// 		Error: errors.New("bad subscriber during logging send message"),
			// 	}
			// 	break
			// }

			// println("user sub:", subscriber.UserSub)
			// println("conn id:", subscriber.ConnectionId, "is trying to send to", cmd.Request.Targets, string(cmd.Request.MessageBytes))
			// var connectionIds string
			// for k := range socketMaps.connections {
			// 	connectionIds += k + " "
			// }
			// println("tar len", len(cmd.Request.Targets))
			for i := 0; i+CID_LENGTH <= len(cmd.Request.Targets); i += CID_LENGTH {
				connId := cmd.Request.Targets[i : i+CID_LENGTH]
				// println("checking connid", connId)
				// println("current attempted", attemptedTargets)
				if strings.Index(attemptedTargets, connId) == -1 {
					// println("attempting to check", connId, "with current ids being", connectionIds)
					if conn, ok := socketMaps.connections[connId]; ok {
						// println("setting deadline")
						// if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
						// 	sendErr = err
						// 	continue
						// }

						// println("did match connection")
						sendErr = conn.WriteMessage(websocket.TextMessage, cmd.Request.MessageBytes)
						if sendErr != nil {
							// println("FAILED WITH WRITE ", sendErr.Error())
							continue
						}
						// println("did send success")
						sentAtLeastOne = true
					}

					attemptedTargets += connId
				}
			}
			if !sentAtLeastOne && sendErr == nil {
				// println("FAILED WITH SEND ONE ")
				sendErr = noTargetsToSend
			}
			cmd.ReplyChan <- SocketResponse{
				Error: sendErr,
			}
			// println("============== End Send event ================\n")

		case AddSubscribedTopicSocketCommand:
			subscriber, ok := socketMaps.subscribers[cmd.Request.UserSub]
			if !ok {
				cmd.ReplyChan <- SocketResponse{
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
				cmd.ReplyChan <- SocketResponse{
					Error: noSubscriberTargets,
				}
				break
			}

			cmd.ReplyChan <- SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					Targets: subscriber.ConnectionIds,
				},
			}

		case GetGroupTargetsSocketCommand:
			targets, ok := socketMaps.groupTargets[cmd.Request.GroupId]
			if !ok {
				cmd.ReplyChan <- SocketResponse{
					Error: noSubscriberTargets,
				}
				break
			}

			cmd.ReplyChan <- SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					Targets: targets,
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

			cmd.ReplyChan <- SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					HasSub: hasSub,
				},
			}

		default:
			log.Fatal("unknown command type", cmd.Ty)
		}

		return true
	})

	util.DebugLog.Println("Sock Init")

	return &Socket{
		handlerId:  sockHandlerId,
		SocketMaps: socketMaps,
	}
}

type SocketRequest struct {
	*websocket.Conn
	*types.SocketRequestParams
}

type SocketResponse struct {
	Error error
	*types.SocketResponseParams
}

type SocketCommand struct {
	Request   SocketRequest
	ReplyChan chan SocketResponse
	*types.WorkerCommandParams
}

func (cmd SocketCommand) GetClientId() string {
	return cmd.ClientId
}

func (cmd SocketCommand) GetReplyChannel() any {
	return cmd.ReplyChan
}

func (s *Socket) RouteCommand(ctx context.Context, cmd SocketCommand) error {
	return GetGlobalWorkerPool().RouteCommand(ctx, cmd)
}

func (s *Socket) Close() {
	GetGlobalWorkerPool().UnregisterProcessFunction(s.handlerId)
}

func (s *Socket) Connected(userSub string) bool {
	s.SocketMaps.mu.RLock()
	defer s.SocketMaps.mu.RUnlock()
	if subscriber, ok := s.SocketMaps.subscribers[userSub]; ok {
		_, ok := s.SocketMaps.connections[subscriber.ConnectionId]
		return ok
	}
	return false
}

func (s *Socket) SendCommand(ctx context.Context, cmdType int32, request *types.SocketRequestParams) (*types.SocketResponseParams, error) {
	finish := util.RunTimer(cmdType)
	defer finish()

	if request.UserSub == "" {
		return nil, socketCommandMustHaveSub
	}
	createCmd := func(replyChan chan SocketResponse) SocketCommand {
		return SocketCommand{
			WorkerCommandParams: &types.WorkerCommandParams{
				Ty:       cmdType,
				ClientId: request.UserSub,
			},
			Request: SocketRequest{
				SocketRequestParams: request,
			},
			ReplyChan: replyChan,
		}
	}

	res, err := SendCommand(ctx, s, createCmd)
	err = ChannelError(err, res.Error)
	if err != nil {
		return nil, err
	}

	return res.SocketResponseParams, nil
}

func (s *Socket) StoreConn(ctx context.Context, ticket string, conn *websocket.Conn) (*types.ConcurrentUserSession, error) {
	createCmd := func(replyChan chan SocketResponse) SocketCommand {
		return SocketCommand{
			WorkerCommandParams: &types.WorkerCommandParams{
				Ty:       CreateSocketConnectionSocketCommand,
				ClientId: "worker",
			},
			Request: SocketRequest{
				SocketRequestParams: &types.SocketRequestParams{
					Ticket: ticket,
				},
				Conn: conn,
			},
			ReplyChan: replyChan,
		}
	}

	res, err := SendCommand(ctx, s, createCmd)
	err = ChannelError(err, res.Error)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return types.NewConcurrentUserSession(&types.UserSession{
		UserSub:  res.SocketResponseParams.UserSub,
		GroupId:  res.SocketResponseParams.GroupId,
		RoleBits: res.SocketResponseParams.RoleBits,
	}), nil
}

func (s *Socket) GetSocketTicket(ctx context.Context, session *types.ConcurrentUserSession) (string, error) {
	response, err := s.SendCommand(ctx, CreateSocketTicketSocketCommand, &types.SocketRequestParams{
		UserSub:  session.GetUserSub(),
		GroupId:  session.GetGroupId(),
		RoleBits: session.GetRoleBits(),
	})

	if err != nil {
		return "", util.ErrCheck(err)
	}

	return response.Ticket, nil
}

func (s *Socket) SendMessageBytes(ctx context.Context, userSub, targets string, messageBytes []byte) error {
	if targets == "" {
		return util.ErrCheck(noTargetsProvided)
	}
	if len(userSub) == 0 {
		return util.ErrCheck(subRequiredToSend)
	}

	_, err := s.SendCommand(ctx, SendSocketMessageSocketCommand, &types.SocketRequestParams{
		UserSub:      userSub,
		Targets:      targets,
		MessageBytes: messageBytes,
	})
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (s *Socket) SendMessage(ctx context.Context, userSub, targets string, message *types.SocketMessage) error {
	finish := util.RunTimer()
	defer finish()
	if message == nil {
		return util.ErrCheck(messageRequired)
	}
	return s.SendMessageBytes(ctx, userSub, targets, util.GenerateMessage(util.DefaultPadding, message))
}

func (s *Socket) RoleCall(userSub string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if connected := s.Connected(userSub); !connected {
		return nil
	}

	response, err := s.SendCommand(ctx, GetSubscribedTargetsSocketCommand, &types.SocketRequestParams{
		UserSub: userSub,
	})
	if err != nil {
		return util.ErrCheck(err)
	}

	if len(response.Targets) > 0 {
		err := s.SendMessage(ctx, userSub, response.Targets, &types.SocketMessage{
			Action: types.SocketActions_ROLE_CALL,
		})
		if err != nil {
			return util.ErrCheck(err)
		}
	}

	return nil
}

func (s *Socket) GroupRoleCall(userSub, groupId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := s.SendCommand(ctx, GetGroupTargetsSocketCommand, &types.SocketRequestParams{
		UserSub: userSub,
		GroupId: groupId,
	})
	if err != nil && !errors.Is(err, noSubscriberTargets) {
		return util.ErrCheck(err)
	}

	if response != nil && len(response.Targets) > 0 {
		err := s.SendMessage(ctx, userSub, response.Targets, &types.SocketMessage{
			Action: types.SocketActions_ROLE_CALL,
		})
		if err != nil {
			return util.ErrCheck(err)
		}
	}

	return nil
}
