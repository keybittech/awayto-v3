package clients

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

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
	GetSubscribedTopicTargetsSocketCommand
	DeleteSubscribedTopicSocketCommand
	HasSubscribedTopicSocketCommand
)

type SocketParams struct {
	UserSub      string
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
	Timestamp  string              `json:"timestamp,omitempty"`
	Sender     string              `json:"sender,omitempty"`
	Target     string              `json:"target,omitempty"`
	Action     types.SocketActions `json:"action,omitempty"`
	Topic      string              `json:"topic,omitempty"`
	Store      bool                `json:"store,omitempty"`
	Payload    interface{}         `json:"payload,omitempty"`
	Historical bool                `json:"historical,omitempty"`
}

type Subscriber struct {
	UserSub          string
	ConnectionIds    []string
	Tickets          map[string]string
	SubscribedTopics map[string][]string
}

type Subscribers map[string]Subscriber
type Connections map[string]net.Conn

func InitSocket() ISocket {

	// TODO only setup sockets for payers
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

				for _, subscriber := range subscribers {
					for ticketAuth := range subscriber.Tickets {
						if auth == ticketAuth {
							cmd.ReplyChan <- SocketResponse{Subscriber: &subscriber}
						}
					}
				}

			case SendSocketMessageSocketCommand:
				sent := []string{}
				failed := []string{}
				for _, target := range cmd.Params.Targets {
					if conn, ok := connections[target]; ok {
						err := util.WriteSocketConnectionMessage(cmd.Params.MessageBytes, conn)
						if err != nil {
							failed = append(failed, target)
						} else {
							sent = append(sent, target)
						}
					}
				}
				cmd.ReplyChan <- SocketResponse{Total: len(cmd.Params.Targets), Sent: sent, Failed: failed}

			case AddSubscribedTopicSocketCommand:
				// Do not remove the _ here as we want to directly modify the original subscribers object
				if _, ok := subscribers[cmd.Params.UserSub]; ok {
					subscribers[cmd.Params.UserSub].SubscribedTopics[cmd.Params.Topic] = cmd.Params.Targets
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

	sc := &Socket{}
	sc.SetChan(cmds)
	return sc
}

func (s *Socket) Chan() chan<- SocketCommand {
	return s.ch
}

func (s *Socket) SetChan(c chan<- SocketCommand) {
	s.ch = c
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

func (s *Socket) GetSocketTicket(sub string) (string, error) {
	replyChan := make(chan SocketResponse)
	s.Chan() <- SocketCommand{
		Ty:        CreateSocketTicketSocketCommand,
		Params:    SocketParams{UserSub: sub},
		ReplyChan: replyChan,
	}
	reply := <-replyChan

	if reply.Error != nil {
		return "", util.ErrCheck(reply.Error)
	}

	return reply.Ticket, nil
}

func (s *Socket) SendMessageBytes(targets []string, messageBytes []byte) {
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

	fmt.Println(fmt.Sprintf("Sent message bytes messages. Sent %d Failed %d", len(reply.Sent), len(reply.Failed)))
}

func (s *Socket) SendMessage(targets []string, message SocketMessage) {
	if len(targets) == 0 {
		return
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		util.ErrCheck(err)
		return
	}

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
	fmt.Println(fmt.Sprintf("Sent normal messages. Sent %d Failed %d", len(reply.Sent), len(reply.Failed)))
}

func (s *Socket) SendMessageWithReply(targets []string, message SocketMessage, replyChan chan SocketResponse) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return util.ErrCheck(err)
	}

	s.Chan() <- SocketCommand{
		Ty: SendSocketMessageSocketCommand,
		Params: SocketParams{
			Targets:      targets,
			MessageBytes: messageBytes,
		},
		ReplyChan: replyChan,
	}

	return nil
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
