package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
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
	GetSubscribedTopicTargetsSocketCommand
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
	Targets      []string
	MessageBytes []byte
}

type SocketResponse struct {
	Ticket     string
	Error      error
	Subscriber *Subscriber
	Total      int
	Targets    []string
	Sent       []string
	Failed     []string
	HasSub     bool
}

type SocketCommand struct {
	Ty        SocketCommandType
	Params    SocketParams
	ReplyChan chan SocketResponse
}

type SocketMessage struct {
	Action     types.SocketActions
	Store      bool
	Historical bool
	Topic      string
	Sender     string
	Payload    interface{}
	Timestamp  string
}

type Subscriber struct {
	UserSub          string
	GroupId          string
	Roles            string
	ConnectionIds    []string
	Tickets          map[string]string
	SubscribedTopics map[string][]string
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
						SubscribedTopics: map[string][]string{},
					}
				}
				subscribers[cmd.Params.UserSub] = subscriber
				cmd.ReplyChan <- SocketResponse{Ticket: ticket}

			case CreateSocketConnectionSocketCommand:
				if subscriber, ok := subscribers[cmd.Params.UserSub]; ok {
					auth, connId, _ := util.SplitSocketId(cmd.Params.Ticket)
					delete(subscriber.Tickets, auth)
					subscriber.ConnectionIds = append(subscriber.ConnectionIds, connId)
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
					subscriber.ConnectionIds = util.StringOut(connId, subscriber.ConnectionIds)
					if len(subscriber.ConnectionIds) == 0 {
						delete(subscribers, cmd.Params.UserSub)
					} else {
						subscribers[cmd.Params.UserSub] = subscriber
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
					for ticketAuth := range subscriber.Tickets {
						if auth == ticketAuth {
							foundSub = &subscriber
							break
						}
					}
					if foundSub != nil {
						break
					}
				}

				if foundSub != nil {
					cmd.ReplyChan <- SocketResponse{Subscriber: foundSub}
				} else {
					cmd.ReplyChan <- SocketResponse{Error: errors.New("no subscriber found for ticket")}
				}

			case SendSocketMessageSocketCommand:
				sent := []string{}
				failed := []string{}
				var sendErr error
				for _, target := range cmd.Params.Targets {
					if conn, ok := connections[target]; ok {
						sendErr = util.WriteSocketConnectionMessage(cmd.Params.MessageBytes, conn)
						if sendErr != nil {
							failed = append(failed, target)
						} else {
							sent = append(sent, target)
						}
					}
				}
				cmd.ReplyChan <- SocketResponse{Total: len(cmd.Params.Targets), Sent: sent, Failed: failed, Error: sendErr}

			case AddSubscribedTopicSocketCommand:
				// Do not remove the _ here as we want to directly modify the original subscribers object
				if _, ok := subscribers[cmd.Params.UserSub]; ok {
					subscribers[cmd.Params.UserSub].SubscribedTopics[cmd.Params.Topic] = cmd.Params.Targets
				}

			case GetSubscribedTargetsSocketCommand:
				if subscriber, ok := subscribers[cmd.Params.UserSub]; ok {
					cmd.ReplyChan <- SocketResponse{Targets: subscriber.ConnectionIds}
				} else {
					cmd.ReplyChan <- SocketResponse{Error: errors.New("subscriber not found")}
				}

			case GetSubscribedTopicTargetsSocketCommand:
				targets := []string{}
				if sub, okSub := subscribers[cmd.Params.UserSub]; okSub {
					if topicTargets, okTopic := sub.SubscribedTopics[cmd.Params.Topic]; okTopic {
						targets = append(targets, topicTargets...)
					}
				}
				cmd.ReplyChan <- SocketResponse{Targets: targets}

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

func GenerateMessage(padTo int, message SocketMessage) []byte {
	storeStr := "f"
	if message.Store {
		storeStr = "t"
	}

	historicalStr := "f"
	if message.Historical {
		historicalStr = "t"
	}

	payloadStr := ""
	switch v := message.Payload.(type) {
	case string:
		payloadStr = v
	case nil:
		payloadStr = ""
	default:
		pl, err := json.Marshal(v)
		if err == nil {
			payloadStr = string(pl)
		}
	}

	return []byte(fmt.Sprintf("%s%d%s%s%s%s%s%s%s%s%s%s%s%s",
		util.PaddedLen(padTo, len(strconv.Itoa(int(message.Action.Number())))), message.Action.Number(),
		util.PaddedLen(padTo, 1), storeStr,
		util.PaddedLen(padTo, 1), historicalStr,
		util.PaddedLen(padTo, len(message.Timestamp)), message.Timestamp,
		util.PaddedLen(padTo, len(message.Topic)), message.Topic,
		util.PaddedLen(padTo, len(message.Sender)), message.Sender,
		util.PaddedLen(padTo, len(payloadStr)), payloadStr,
	))
}

func (s *Socket) SendMessageBytes(targets []string, messageBytes []byte) error {
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

	// fmt.Println(fmt.Sprintf("Sent message bytes messages. Sent %d Failed %d", len(reply.Sent), len(reply.Failed)))

	if reply.Error != nil {
		return util.ErrCheck(reply.Error)
	}

	return nil
}

func (s *Socket) SendMessage(targets []string, message SocketMessage) error {
	if len(targets) == 0 {
		return util.ErrCheck(errors.New("no targets to send message to"))
	}

	return s.SendMessageBytes(targets, GenerateMessage(util.DefaultPadding, message))
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

func (s *Socket) AddSubscribedTopic(userSub, topic string, existingCids []string) {
	s.Chan() <- SocketCommand{
		Ty: AddSubscribedTopicSocketCommand,
		Params: SocketParams{
			UserSub: userSub,
			Topic:   topic,
			Targets: existingCids,
		},
	}
}

func (s *Socket) GetSubscribedTopicTargets(userSub, topic string) []string {
	replyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty: GetSubscribedTopicTargetsSocketCommand,
		Params: SocketParams{
			UserSub: userSub,
			Topic:   topic,
		},
		ReplyChan: replyChan,
	}
	reply := <-replyChan
	close(replyChan)
	return reply.Targets
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

func (s *Socket) NotifyTopicUnsub(topic, socketId string, targets []string) {
	s.SendMessage(targets, SocketMessage{
		Action:  types.SocketActions_UNSUBSCRIBE_TOPIC,
		Topic:   topic,
		Payload: socketId,
	})
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
		if err := s.SendMessage(reply.Targets, SocketMessage{Action: types.SocketActions_ROLE_CALL}); err != nil {
			return util.ErrCheck(reply.Error)
		}
	}

	return nil
}
