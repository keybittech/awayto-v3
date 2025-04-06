package clients

import (
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/google/uuid"
)

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
	Subscriber *types.Subscriber
	Total      int
	Targets    string
	HasSub     bool
}

type SocketCommand struct {
	Ty        SocketCommandType
	Params    SocketParams
	ReplyChan chan SocketResponse
	ClientId  string
}

func (cmd SocketCommand) GetClientId() string {
	return cmd.ClientId
}

type Subscribers map[string]*types.Subscriber
type Connections map[string]net.Conn

var (
	CID_LENGTH                    = 36
	SOCK_WORKERS                  = 10
	SOCK_COMMAND_CHAN_BUFFER_SIZE = 10
)

type Socket struct {
	pool *WorkerPool[SocketCommand, SocketParams, SocketResponse]
	// Ch   chan<- SocketCommand
}

func InitSocket() interfaces.ISocket {
	subscribers := make(Subscribers)
	connections := make(Connections)

	processFunc := func(cmd SocketCommand) {
		// Process the socket command based on its type
		switch cmd.Ty {
		case CreateSocketTicketSocketCommand:
			auth := uuid.NewString()
			connectionId := uuid.NewString()
			ticket := auth + ":" + connectionId

			subscriber, ok := subscribers[cmd.Params.UserSub]

			if ok {
				subscriber.Tickets[auth] = connectionId
			} else {
				subscriber = &types.Subscriber{
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
				auth, connId, err := util.SplitSocketId(cmd.Params.Ticket)
				if err != nil {
					cmd.ReplyChan <- SocketResponse{Error: err}
					break
				}

				if _, ok := subscriber.Tickets[auth]; !ok {
					cmd.ReplyChan <- SocketResponse{Error: errors.New("invalid ticket")}
					break
				}

				delete(subscriber.Tickets, auth)
				subscriber.ConnectionIds += connId

				connections[connId] = cmd.Params.Conn
				subscribers[cmd.Params.UserSub] = subscriber
				cmd.ReplyChan <- SocketResponse{Subscriber: subscriber}
			} else {
				cmd.ReplyChan <- SocketResponse{Error: errors.New("no sub found in sock")}
			}

			// when testing this use a test socket to do whatever operation then get subscriber by id to test changes
		case DeleteSocketConnectionSocketCommand:
			if subscriber, ok := subscribers[cmd.Params.UserSub]; ok {
				_, connId, _ := util.SplitSocketId(cmd.Params.Ticket)
				delete(connections, connId)
				connIdStartIdx := strings.Index(subscriber.ConnectionIds, connId)
				connIdEndIdx := connIdStartIdx + CID_LENGTH // uuid length
				if connIdStartIdx == -1 || len(subscriber.ConnectionIds) < connIdEndIdx {
					cmd.ReplyChan <- SocketResponse{Error: errors.New("connection id not found to delete")}
					break
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
				cmd.ReplyChan <- SocketResponse{Subscriber: foundSub}
			} else {
				cmd.ReplyChan <- SocketResponse{Error: errors.New("no subscriber found for ticket")}
			}

		case SendSocketMessageSocketCommand:
			var sentAtLeastOne bool
			var sendErr error
			var attemptedTargets string
			for i := 0; i+CID_LENGTH <= len(cmd.Params.Targets); i += CID_LENGTH {
				connId := cmd.Params.Targets[i : i+CID_LENGTH]
				if strings.Index(attemptedTargets, connId) == -1 {
					if conn, ok := connections[connId]; ok {
						sendErr = util.WriteSocketConnectionMessage(cmd.Params.MessageBytes, conn)
						if sendErr != nil {
							continue
						}
						sentAtLeastOne = true
					}

					attemptedTargets += connId
				}
			}
			if !sentAtLeastOne {
				sendErr = errors.New("no targets to send to")
			}
			cmd.ReplyChan <- SocketResponse{Error: sendErr}

		case AddSubscribedTopicSocketCommand:
			// Do not remove the _ here as we want to directly modify the original subscribers object
			if _, ok := subscribers[cmd.Params.UserSub]; !ok {
				cmd.ReplyChan <- SocketResponse{Error: errors.New("no subscriber found to add topic")}
				break
			}
			for i := 0; i+CID_LENGTH <= len(cmd.Params.Targets); i += CID_LENGTH {
				connId := cmd.Params.Targets[i : i+CID_LENGTH]
				if strings.Index(subscribers[cmd.Params.UserSub].SubscribedTopics[cmd.Params.Topic], connId) == -1 {
					subscribers[cmd.Params.UserSub].SubscribedTopics[cmd.Params.Topic] += connId
				}
			}
			cmd.ReplyChan <- SocketResponse{}

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
			cmd.ReplyChan <- SocketResponse{}

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

	pool := NewWorkerPool[SocketCommand, SocketParams, SocketResponse](SOCK_WORKERS, SOCK_COMMAND_CHAN_BUFFER_SIZE, processFunc)
	pool.Start()

	return &Socket{pool: pool}
}

// NewSocket creates a new Socket with a worker pool
func NewSocket(numWorkers, bufferSize int) *Socket {
	processFunc := func(cmd SocketCommand) {
		// Process the socket command based on its type
		switch cmd.Ty {
		case CreateSocketTicketSocketCommand:
			// Simulate processing time (e.g., network operations)
			time.Sleep(time.Millisecond * 150)
			response := SocketResponse{
				Ticket:  "ticket-" + cmd.ClientId + "-" + cmd.Params.UserSub,
				Targets: cmd.Params.Targets,
				HasSub:  false,
			}
			cmd.ReplyChan <- response
		case HasSubscribedTopicSocketCommand:
			// Simulate processing time
			time.Sleep(time.Millisecond * 100)
			response := SocketResponse{
				Ticket:  "",
				Targets: cmd.Params.Targets,
				HasSub:  true,
			}
			cmd.ReplyChan <- response
		}
	}

	pool := NewWorkerPool[SocketCommand, SocketParams, SocketResponse](numWorkers, bufferSize, processFunc)
	pool.Start()

	return &Socket{pool: pool}
}

// GetCommandChannel returns the channel for sending commands
func (s *Socket) GetCommandChannel() chan<- SocketCommand {
	return s.pool.GetCommandChannel()
}

// Close gracefully shuts down the socket worker pool
func (s *Socket) Close() {
	s.pool.Stop()
}

func (s *Socket) SendCommand(cmdType SocketCommandType, params SocketParams) (SocketResponse, error) {
	createCmd := func(p SocketParams, replyChan chan SocketResponse) SocketCommand {
		return SocketCommand{
			Ty:        cmdType,
			Params:    p,
			ReplyChan: replyChan,
		}
	}

	return SendCommand(s, createCmd, params)
}

func (s *Socket) InitConnection(conn net.Conn, userSub string, ticket string) (func(), error) {
	response, err := s.SendCommand(CreateSocketConnectionSocketCommand, SocketParams{
		UserSub: userSub,
		Ticket:  ticket,
		Conn:    conn,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return func() {
		s.SendCommand(DeleteSocketConnectionSocketCommand, SocketParams{
			UserSub: userSub,
			Ticket:  ticket,
		})
	}, nil
}

func (s *Socket) GetSocketTicket(session *types.UserSession) (string, error) {
	response, err := s.SendCommand(CreateSocketTicketSocketCommand, SocketParams{
		UserSub: session.UserSub,
		GroupId: session.GroupId,
		Roles:   strings.Join(session.AvailableUserGroupRoles, " "),
	})

	if err = ChannelError(err, response.Error); err != nil {
		return "", util.ErrCheck(err)
	}

	return response.Ticket, nil
}

func (s *Socket) SendMessageBytes(messageBytes []byte, targets string) error {
	if targets == "" {
		return util.ErrCheck(errors.New("no targets provided"))
	}

	response, err := s.SendCommand(SendSocketMessageSocketCommand, SocketParams{
		Targets:      targets,
		MessageBytes: messageBytes,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (s *Socket) SendMessage(message *types.SocketMessage, targets string) error {
	if message == nil {
		return util.ErrCheck(errors.New("message object required"))
	}
	if len(targets) == 0 {
		return util.ErrCheck(errors.New("no targets to send message to"))
	}

	return s.SendMessageBytes(util.GenerateMessage(util.DefaultPadding, message), targets)
}

func (s *Socket) GetSubscriberByTicket(ticket string) (*types.Subscriber, error) {
	if ticket == "" {
		return nil, util.ErrCheck(errors.New("ticket required"))
	}

	response, err := s.SendCommand(GetSubscriberSocketCommand, SocketParams{
		Ticket: ticket,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Subscriber, nil
}

func (s *Socket) AddSubscribedTopic(userSub, topic string, targets string) error {
	response, err := s.SendCommand(AddSubscribedTopicSocketCommand, SocketParams{
		UserSub: userSub,
		Topic:   topic,
		Targets: targets,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (s *Socket) DeleteSubscribedTopic(userSub, topic string) error {
	response, err := s.SendCommand(DeleteSubscribedTopicSocketCommand, SocketParams{
		UserSub: userSub,
		Topic:   topic,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (s *Socket) HasTopicSubscription(topic string) (bool, error) {
	response, err := s.SendCommand(HasSubscribedTopicSocketCommand, SocketParams{
		Topic: topic,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return false, util.ErrCheck(err)
	}

	return response.HasSub, nil
}

func (s *Socket) RoleCall(userSub string) error {
	response, _ := s.SendCommand(GetSubscribedTargetsSocketCommand, SocketParams{
		UserSub: userSub,
	})

	if len(response.Targets) > 0 {
		err := s.SendMessage(&types.SocketMessage{Action: types.SocketActions_ROLE_CALL}, response.Targets)
		if err != nil {
			return util.ErrCheck(response.Error)
		}
	}

	return nil
}
