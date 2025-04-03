package clients

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/google/uuid"
)

type Socket struct {
	ch chan<- SocketCommand
}

type SocketCommandType int

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

type SocketParams struct {
	UserSub      string
	GroupId      string
	Roles        string
	Ticket       string
	Topic        string
	Conn         net.Conn
	Targets      string
	MessageBytes []byte
}

type SocketResponse struct {
	Ticket     string
	Error      error
	Subscriber *Subscriber
	Total      int
	Targets    string
	HasSub     bool
}

type SocketCommand struct {
	Ty        SocketCommandType
	Params    SocketParams
	ReplyChan chan SocketResponse
}

type Subscriber struct {
	UserSub          string
	GroupId          string
	Roles            string
	ConnectionIds    string
	Tickets          map[string]string
	SubscribedTopics map[string]string
}

type Subscribers map[string]Subscriber
type Connections map[string]net.Conn

func InitSocket() ISocket {

	cmds := make(chan SocketCommand)
	subscribers := make(Subscribers)
	connections := make(Connections)

	go func() {
		for cmd := range cmds {
			switch cmd.Ty {
			case CreateSocketTicketSocketCommand:
				auth := uuid.NewString()
				connectionId := uuid.NewString()
				ticket := auth + ":" + connectionId

				subscriber, ok := subscribers[cmd.Params.UserSub]

				if ok {
					subscriber.Tickets[auth] = connectionId
				} else {
					subscriber = Subscriber{
						UserSub:          cmd.Params.UserSub,
						GroupId:          cmd.Params.GroupId,
						Roles:            cmd.Params.Roles,
						Tickets:          map[string]string{auth: connectionId},
						SubscribedTopics: make(map[string]string),
					}
				}
				subscribers[cmd.Params.UserSub] = subscriber
				cmd.ReplyChan <- SocketResponse{Ticket: ticket}

			case CreateSocketConnectionSocketCommand:
				if subscriber, ok := subscribers[cmd.Params.UserSub]; ok {
					auth, connId, _ := util.SplitSocketId(cmd.Params.Ticket)
					delete(subscriber.Tickets, auth)
					subscriber.ConnectionIds += connId + " "
					connections[connId] = cmd.Params.Conn
					subscribers[cmd.Params.UserSub] = subscriber
					cmd.ReplyChan <- SocketResponse{Subscriber: &subscriber}
				} else {
					cmd.ReplyChan <- SocketResponse{Error: errors.New("no sub found in sock")}
				}

			case DeleteSocketConnectionSocketCommand:
				if subscriber, ok := subscribers[cmd.Params.UserSub]; ok {
					_, connId, _ := util.SplitSocketId(cmd.Params.Ticket)
					delete(connections, connId)
					connIdStartIdx := strings.Index(subscriber.ConnectionIds, connId)
					connIdEndIdx := connIdStartIdx + 37 // uuid + space length
					if connIdStartIdx == -1 || len(subscriber.ConnectionIds) < connIdEndIdx {
						cmd.ReplyChan <- SocketResponse{}
						continue
					}
					subscriber.ConnectionIds = subscriber.ConnectionIds[:connIdStartIdx] + subscriber.ConnectionIds[connIdEndIdx:]
					if len(subscriber.ConnectionIds) == 0 {
						delete(subscribers, cmd.Params.UserSub)
					}
				}
				cmd.ReplyChan <- SocketResponse{}
			case GetSubscriberSocketCommand:
				auth, _, err := util.SplitSocketId(cmd.Params.Ticket)
				if err != nil {
					cmd.ReplyChan <- SocketResponse{Error: err}
				}

				var foundSub *Subscriber

				for _, subscriber := range subscribers {
					if _, ok := subscriber.Tickets[auth]; ok {
						foundSub = &subscriber
						break
					}
				}

				if foundSub != nil {
					cmd.ReplyChan <- SocketResponse{Subscriber: foundSub}
				} else {
					cmd.ReplyChan <- SocketResponse{Error: errors.New("no subscriber found for ticket")}
				}

			case SendSocketMessageSocketCommand:
				var sendErr error
				var lastIdx int
				var attemptedTargets string
				singleTargetLen := len(cmd.Params.Targets) - 1
				for currentIdx, targetRune := range cmd.Params.Targets {
					if targetRune == ' ' || currentIdx == singleTargetLen {
						if currentIdx == singleTargetLen {
							currentIdx = currentIdx + 1
						}

						currentTarget := cmd.Params.Targets[lastIdx:currentIdx]
						if strings.Index(attemptedTargets, currentTarget) == -1 {
							if conn, ok := connections[currentTarget]; ok {
								sendErr = util.WriteSocketConnectionMessage(cmd.Params.MessageBytes, conn)
								if sendErr != nil {
									continue
								}
							}

							attemptedTargets += currentTarget
						}
						lastIdx = currentIdx + 1
					}
				}
				cmd.ReplyChan <- SocketResponse{Error: sendErr}

			case AddSubscribedTopicSocketCommand:
				// Do not remove the _ here as we want to directly modify the original subscribers object
				if _, ok := subscribers[cmd.Params.UserSub]; !ok {
					continue
				}
				var lastIdx int
				for currentIdx, targetRune := range cmd.Params.Targets {
					if targetRune == ' ' {
						currentTarget := cmd.Params.Targets[lastIdx:currentIdx]
						if strings.Index(subscribers[cmd.Params.UserSub].SubscribedTopics[cmd.Params.Topic], currentTarget) == -1 {
							subscribers[cmd.Params.UserSub].SubscribedTopics[cmd.Params.Topic] += currentTarget + " "
						}

						lastIdx = currentIdx + 1
					}
				}

			case GetSubscribedTargetsSocketCommand:
				if subscriber, ok := subscribers[cmd.Params.UserSub]; ok {
					cmd.ReplyChan <- SocketResponse{Targets: subscriber.ConnectionIds}
				} else {
					cmd.ReplyChan <- SocketResponse{Error: errors.New("subscriber not found")}
				}

			case DeleteSubscribedTopicSocketCommand:
				if _, ok := subscribers[cmd.Params.UserSub]; ok {
					delete(subscribers[cmd.Params.UserSub].SubscribedTopics, cmd.Params.Topic)
				}

			case HasSubscribedTopicSocketCommand:
				hasSub := false
				if subscriber, okSub := subscribers[cmd.Params.UserSub]; okSub {
					if _, okTopic := subscriber.SubscribedTopics[cmd.Params.Topic]; okTopic {
						hasSub = true
					}
				}
				cmd.ReplyChan <- SocketResponse{HasSub: hasSub}

			default:
				log.Fatal("unknown command type", cmd.Ty)
			}
		}
	}()

	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println(fmt.Sprintf("\nSock Report:\nConnections: %+v\nSubscribers: %+v\n", connections, subscribers))
			}
		}
	}()

	return &Socket{cmds}
}

func (s *Socket) Chan() chan<- SocketCommand {
	return s.ch
}

func (s *Socket) InitConnection(conn net.Conn, userSub string, ticket string) (func(), error) {
	replyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty: CreateSocketConnectionSocketCommand,
		Params: SocketParams{
			UserSub: userSub,
			Ticket:  ticket,
			Conn:    conn,
		},
		ReplyChan: replyChan,
	}

	reply := <-replyChan
	if reply.Error != nil {
		return nil, util.ErrCheck(reply.Error)
	}

	return func() {
		replyChan := make(chan SocketResponse)
		s.Chan() <- SocketCommand{
			Ty: DeleteSocketConnectionSocketCommand,
			Params: SocketParams{
				UserSub: userSub,
				Ticket:  ticket,
			},
			ReplyChan: replyChan,
		}
		<-replyChan
	}, nil
}

func (s *Socket) GetSocketTicket(session *UserSession) (string, error) {
	replyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty:        CreateSocketTicketSocketCommand,
		Params:    SocketParams{UserSub: session.UserSub, GroupId: session.GroupId, Roles: strings.Join(session.AvailableUserGroupRoles, " ")},
		ReplyChan: replyChan,
	}
	reply := <-replyChan

	if reply.Error != nil {
		return "", util.ErrCheck(reply.Error)
	}

	return reply.Ticket, nil
}

func (s *Socket) SendMessageBytes(messageBytes []byte, targets string) error {
	replyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty: SendSocketMessageSocketCommand,
		Params: SocketParams{
			Targets:      targets,
			MessageBytes: messageBytes,
		},
		ReplyChan: replyChan,
	}
	reply := <-replyChan
	close(replyChan)

	if reply.Error != nil {
		return util.ErrCheck(reply.Error)
	}

	return nil
}

func (s *Socket) SendMessage(message *util.SocketMessage, targets string) error {
	if len(targets) == 0 {
		return util.ErrCheck(errors.New("no targets to send message to"))
	}

	return s.SendMessageBytes(util.GenerateMessage(util.DefaultPadding, message), targets)
}

func (s *Socket) GetSubscriberByTicket(ticket string) (*Subscriber, error) {
	sockGetSubReplyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty:        GetSubscriberSocketCommand,
		Params:    SocketParams{Ticket: ticket},
		ReplyChan: sockGetSubReplyChan,
	}
	sockGetSubReply := <-sockGetSubReplyChan
	close(sockGetSubReplyChan)

	if sockGetSubReply.Error != nil {
		return nil, util.ErrCheck(sockGetSubReply.Error)
	}

	return sockGetSubReply.Subscriber, nil
}

func (s *Socket) AddSubscribedTopic(userSub, topic string, targets string) {
	s.Chan() <- SocketCommand{
		Ty: AddSubscribedTopicSocketCommand,
		Params: SocketParams{
			UserSub: userSub,
			Topic:   topic,
			Targets: targets,
		},
	}
}

func (s *Socket) DeleteSubscribedTopic(userSub, topic string) {
	s.Chan() <- SocketCommand{
		Ty: DeleteSubscribedTopicSocketCommand,
		Params: SocketParams{
			UserSub: userSub,
			Topic:   topic,
		},
	}
}

func (s *Socket) HasTopicSubscription(userSub, topic string) bool {
	replyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty: HasSubscribedTopicSocketCommand,
		Params: SocketParams{
			UserSub: userSub,
			Topic:   topic,
		},
		ReplyChan: replyChan,
	}
	reply := <-replyChan
	close(replyChan)
	return reply.HasSub
}

func (s *Socket) RoleCall(userSub string) error {
	replyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty: GetSubscribedTargetsSocketCommand,
		Params: SocketParams{
			UserSub: userSub,
		},
		ReplyChan: replyChan,
	}
	reply := <-replyChan
	close(replyChan)

	if len(reply.Targets) > 0 {
		if err := s.SendMessage(&util.SocketMessage{Action: types.SocketActions_ROLE_CALL}, reply.Targets); err != nil {
			return util.ErrCheck(reply.Error)
		}
	}

	return nil
}
