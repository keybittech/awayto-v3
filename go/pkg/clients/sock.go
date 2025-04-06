package clients

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

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
	GetSubscriberSocketCommand
	SendSocketMessageSocketCommand
	AddSubscribedTopicSocketCommand
	GetSubscribedTargetsSocketCommand
	DeleteSubscribedTopicSocketCommand
	HasSubscribedTopicSocketCommand
)

type Subscribers map[string]*types.Subscriber
type Connections map[string]net.Conn

var (
	CID_LENGTH                    = 36
	SOCK_WORKERS                  = 10
	SOCK_COMMAND_CHAN_BUFFER_SIZE = 10
)

type Socket struct {
	interfaces.ISocket
	pool *WorkerPool[interfaces.SocketCommand, interfaces.SocketRequest, interfaces.SocketResponse]
}

type ConnectedSocket struct {
	Conn net.Conn
	Sock *Socket
}

func InitSocket() *Socket {
	subscribers := make(Subscribers)
	connections := make(Connections)

	processFunc := func(cmd interfaces.SocketCommand) {
		// Process the socket command based on its type
		switch cmd.Ty {
		case CreateSocketTicketSocketCommand:
			auth := uuid.NewString()
			connectionId := uuid.NewString()
			ticket := auth + ":" + connectionId

			subscriber, ok := subscribers[cmd.Request.UserSub]

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
			subscribers[cmd.Request.UserSub] = subscriber
			cmd.ReplyChan <- interfaces.SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{Ticket: ticket},
			}

		case CreateSocketConnectionSocketCommand:
			if subscriber, ok := subscribers[cmd.Request.UserSub]; ok {
				auth, connId, err := util.SplitSocketId(cmd.Request.Ticket)
				if err != nil {
					cmd.ReplyChan <- interfaces.SocketResponse{Error: err}
					break
				}

				if _, ok := subscriber.Tickets[auth]; !ok {
					cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("invalid ticket")}
					break
				}

				delete(subscriber.Tickets, auth)
				subscriber.ConnectionIds += connId

				connections[connId] = cmd.Request.Conn
				subscribers[cmd.Request.UserSub] = subscriber
				cmd.ReplyChan <- interfaces.SocketResponse{
					SocketResponseParams: &types.SocketResponseParams{Subscriber: subscriber},
				}
			} else {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("no sub found in sock")}
			}

			// when testing this use a test socket to do whatever operation then get subscriber by id to test changes
		case DeleteSocketConnectionSocketCommand:
			if subscriber, ok := subscribers[cmd.Request.UserSub]; ok {
				_, connId, _ := util.SplitSocketId(cmd.Request.Ticket)
				delete(connections, connId)
				connIdStartIdx := strings.Index(subscriber.ConnectionIds, connId)
				connIdEndIdx := connIdStartIdx + CID_LENGTH // uuid length
				if connIdStartIdx == -1 || len(subscriber.ConnectionIds) < connIdEndIdx {
					cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("connection id not found to delete")}
					break
				}
				subscriber.ConnectionIds = subscriber.ConnectionIds[:connIdStartIdx] + subscriber.ConnectionIds[connIdEndIdx:]
				if len(subscriber.ConnectionIds) == 0 {
					delete(subscribers, cmd.Request.UserSub)
				}
			}
			cmd.ReplyChan <- interfaces.SocketResponse{}
		case GetSubscriberSocketCommand:
			auth, _, err := util.SplitSocketId(cmd.Request.Ticket)
			if err != nil {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: err}
				break
			}

			var foundSub *types.Subscriber

			for _, subscriber := range subscribers {
				if _, ok := subscriber.Tickets[auth]; ok {
					foundSub = subscriber
					break
				}
			}

			if foundSub != nil {
				cmd.ReplyChan <- interfaces.SocketResponse{
					SocketResponseParams: &types.SocketResponseParams{Subscriber: foundSub},
				}
			} else {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("no subscriber found for ticket")}
			}

		case SendSocketMessageSocketCommand:
			var sentAtLeastOne bool
			var sendErr error
			var attemptedTargets string
			for i := 0; i+CID_LENGTH <= len(cmd.Request.Targets); i += CID_LENGTH {
				connId := cmd.Request.Targets[i : i+CID_LENGTH]
				if strings.Index(attemptedTargets, connId) == -1 {
					if conn, ok := connections[connId]; ok {
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
				sendErr = errors.New(fmt.Sprintf("no targets to send to %v %v", cmd, connections))
			}
			cmd.ReplyChan <- interfaces.SocketResponse{Error: sendErr}

		case AddSubscribedTopicSocketCommand:
			// Do not remove the _ here as we want to directly modify the original subscribers object
			if _, ok := subscribers[cmd.Request.UserSub]; !ok {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("no subscriber found to add topic")}
				break
			}
			for i := 0; i+CID_LENGTH <= len(cmd.Request.Targets); i += CID_LENGTH {
				connId := cmd.Request.Targets[i : i+CID_LENGTH]
				if strings.Index(subscribers[cmd.Request.UserSub].SubscribedTopics[cmd.Request.Topic], connId) == -1 {
					subscribers[cmd.Request.UserSub].SubscribedTopics[cmd.Request.Topic] += connId
				}
			}
			cmd.ReplyChan <- interfaces.SocketResponse{}

		case GetSubscribedTargetsSocketCommand:
			if subscriber, ok := subscribers[cmd.Request.UserSub]; ok {
				cmd.ReplyChan <- interfaces.SocketResponse{
					SocketResponseParams: &types.SocketResponseParams{Targets: subscriber.ConnectionIds},
				}
			} else {
				cmd.ReplyChan <- interfaces.SocketResponse{Error: errors.New("subscriber not found")}
			}

		case DeleteSubscribedTopicSocketCommand:
			if _, ok := subscribers[cmd.Request.UserSub]; ok {
				delete(subscribers[cmd.Request.UserSub].SubscribedTopics, cmd.Request.Topic)
			}
			cmd.ReplyChan <- interfaces.SocketResponse{}

		case HasSubscribedTopicSocketCommand:
			hasSub := false
			if subscriber, okSub := subscribers[cmd.Request.UserSub]; okSub {
				if _, okTopic := subscriber.SubscribedTopics[cmd.Request.Topic]; okTopic {
					hasSub = true
				}
			}
			cmd.ReplyChan <- interfaces.SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{HasSub: hasSub},
			}

		default:
			log.Fatal("unknown command type", cmd.Ty)
		}
	}

	pool := NewWorkerPool[
		interfaces.SocketCommand,
		interfaces.SocketRequest,
		interfaces.SocketResponse,
	](
		SOCK_WORKERS,
		SOCK_COMMAND_CHAN_BUFFER_SIZE,
		processFunc,
	)
	pool.Start()

	return &Socket{pool: pool}
}

// GetCommandChannel returns the channel for sending commands
func (s *Socket) GetCommandChannel() chan<- interfaces.SocketCommand {
	return s.pool.GetCommandChannel()
}

// Close gracefully shuts down the socket worker pool
func (s *Socket) Close() {
	s.pool.Stop()
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

// func (s *Socket) GetSubscriberByTicket(ticket string) (*types.Subscriber, error) {
// 	if ticket == "" {
// 		return nil, util.ErrCheck(errors.New("ticket required"))
// 	}
//
// 	response, err := s.SendCommand(GetSubscriberSocketCommand, interfaces.SocketRequest{
// 		SocketRequestParams: &types.SocketRequestParams{
// 			Ticket: ticket,
// 		},
// 	})
//
// 	if err = ChannelError(err, response.Error); err != nil {
// 		return nil, util.ErrCheck(err)
// 	}
//
// 	return response.Subscriber, nil
// }

// func (s *Socket) AddSubscribedTopic(userSub, topic string, targets string) error {
// 	response, err := s.SendCommand(AddSubscribedTopicSocketCommand, interfaces.SocketRequest{
// 		SocketRequestParams: &types.SocketRequestParams{
// 			UserSub: userSub,
// 			Topic:   topic,
// 			Targets: targets,
// 		},
// 	})
//
// 	if err = ChannelError(err, response.Error); err != nil {
// 		return util.ErrCheck(err)
// 	}
//
// 	return nil
// }

// func (s *Socket) DeleteSubscribedTopic(userSub, topic string) error {
// 	response, err := s.SendCommand(DeleteSubscribedTopicSocketCommand, interfaces.SocketRequest{
// 		SocketRequestParams: &types.SocketRequestParams{
// 			UserSub: userSub,
// 			Topic:   topic,
// 		},
// 	})
//
// 	if err = ChannelError(err, response.Error); err != nil {
// 		return util.ErrCheck(err)
// 	}
//
// 	return nil
// }

// func (s *Socket) HasTopicSubscription(userSub, topic string) (bool, error) {
// 	response, err := s.SendCommand(HasSubscribedTopicSocketCommand, interfaces.SocketRequest{
// 		SocketRequestParams: &types.SocketRequestParams{
// 			UserSub: userSub,
// 			Topic:   topic,
// 		},
// 	})
//
// 	if err = ChannelError(err, response.Error); err != nil {
// 		return false, util.ErrCheck(err)
// 	}
//
// 	return response.HasSub, nil
// }

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
