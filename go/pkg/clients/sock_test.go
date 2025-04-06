package clients

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

var testSocket *Socket
var testSocketUserSession *types.UserSession
var testSocketMessage *types.SocketMessage
var testTicket, auth, connId string

func init() {
	var err error
	testSocket = InitSocket()
	testSocketUserSession = &types.UserSession{
		UserSub:                 "user-sub",
		GroupId:                 "group-id",
		AvailableUserGroupRoles: []string{"APP_GROUP_ADMIN"},
	}
	testTicket, err = testSocket.GetSocketTicket(testSocketUserSession)
	if err != nil || testTicket == "" {
		log.Fatal(err)
	}
	_, connId, err = util.SplitSocketId(testTicket)
	if err != nil {
		log.Fatal(err)
	}
	testSocketMessage = &types.SocketMessage{
		Action:     44,
		Topic:      "topic",
		Payload:    "payload",
		Store:      false,
		Historical: false,
		Sender:     "sender",
		Timestamp:  "timestamp",
	}
}

func TestInitSocket(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "does init the websocket client"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitSocket(); got == nil {
				t.Error("InitSocket() returned nil")
			}
		})
	}
}

func getClientData(numClients int, commandType int32, requestParams *types.SocketRequestParams, conn ...bool) ([]string, map[string]func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand) {
	clientIds := make([]string, numClients)
	clientCommands := make(map[string]func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand)
	for i := 0; i < numClients; i++ {
		clientId := "client-" + strconv.Itoa(numClients) + "-" + strconv.Itoa(i)
		clientIds[i] = clientId
		clientCommands[clientId] = func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand {
			requestParams.UserSub = clientId
			request := interfaces.SocketRequest{SocketRequestParams: requestParams}
			if conn != nil {
				request.Conn = &interfaces.NullConn{}
			}
			return interfaces.SocketCommand{
				SocketCommandParams: &types.SocketCommandParams{
					Ty:       commandType,
					ClientId: clientId,
				},
				Request:   request,
				ReplyChan: replyChan,
			}
		}
	}
	return clientIds, clientCommands
}

func BenchmarkSendCommandSingle1Client(b *testing.B) {
	doSendCommandSingleBench(1, b)
}

func BenchmarkSendCommandSingle10Clients(b *testing.B) {
	doSendCommandSingleBench(10, b)
}

func BenchmarkSendCommandSingle100Clients(b *testing.B) {
	doSendCommandSingleBench(100, b)
}

func BenchmarkSendCommandSingle1000Clients(b *testing.B) {
	doSendCommandSingleBench(1000, b)
}

func doSendCommandSingleBench(clientCount int, b *testing.B) {
	socket := InitSocket()
	defer socket.Close()

	clientIds, clientCommands := getClientData(
		clientCount,
		CreateSocketTicketSocketCommand,
		&types.SocketRequestParams{
			GroupId: "group-id",
			Roles:   "APP_GROUP_ADMIN",
		},
	)

	reset(b)

	for i := 0; i < b.N/clientCount; i++ {
		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)

			currentC := c
			go func() {
				defer wg.Done()
				SendCommand(socket, clientCommands[clientIds[currentC]])
			}()
		}

		wg.Wait()
	}
}

func BenchmarkSendCommandMulti1Client1Message(b *testing.B) {
	doSendCommandMultiBench(1, 1, b)
}

func BenchmarkSendCommandMulti1Client10Messages(b *testing.B) {
	doSendCommandMultiBench(1, 10, b)
}

func BenchmarkSendCommandMulti1Client100Messages(b *testing.B) {
	doSendCommandMultiBench(1, 100, b)
}

func BenchmarkSendCommandMulti10Clients1Message(b *testing.B) {
	doSendCommandMultiBench(10, 1, b)
}

func BenchmarkSendCommandMulti10Clients10Messages(b *testing.B) {
	doSendCommandMultiBench(10, 10, b)
}

func BenchmarkSendCommandMulti10Clients100Messages(b *testing.B) {
	doSendCommandMultiBench(10, 100, b)
}

func BenchmarkSendCommandMulti100Clients1Message(b *testing.B) {
	doSendCommandMultiBench(100, 1, b)
}

func BenchmarkSendCommandMulti100Clients10Messages(b *testing.B) {
	doSendCommandMultiBench(100, 10, b)
}

func BenchmarkSendCommandMulti100Clients100Messages(b *testing.B) {
	doSendCommandMultiBench(100, 100, b)
}

func BenchmarkSendCommandMulti1000Clients1Message(b *testing.B) {
	doSendCommandMultiBench(1000, 1, b)
}

func BenchmarkSendCommandMulti1000Clients10Messages(b *testing.B) {
	doSendCommandMultiBench(1000, 10, b)
}

func BenchmarkSendCommandMulti1000Clients100Messages(b *testing.B) {
	doSendCommandMultiBench(1000, 100, b)
}

// Benchmark the worker pool
func doSendCommandMultiBench(clientCount, commandsPerClient int, b *testing.B) {
	clientIds, createCommands := getClientData(
		clientCount,
		CreateSocketTicketSocketCommand,
		&types.SocketRequestParams{
			GroupId: "group-id",
			Roles:   "APP_GROUP_ADMIN",
		},
	)

	socket := InitSocket()
	reset(b)
	for i := 0; i < b.N/(clientCount*commandsPerClient); i++ {

		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)
			currentC := c
			go func() {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					SendCommand(socket, createCommands[clientIds[currentC]])
				}
			}()
		}

		wg.Wait()
	}

	socket.Close()
}

func TestSocket_GetSocketTicket(t *testing.T) {
	type args struct {
		session *types.UserSession
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "get a ticket", args: args{testSocketUserSession}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testSocket.GetSocketTicket(tt.args.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.GetSocketTicket(%v) error = %v, wantErr %v", tt.args.session, err, tt.wantErr)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.GetSocketTicket(%v) error = %v, wantErr %v", tt.args.session, err, tt.wantErr)
				return
			}

			parts := strings.Split(got, ":")
			if len(parts) != 2 {
				t.Errorf("Socket.GetSocketTicket(%v) = %v, ticket format should be UUID:UUID", tt.args.session, got)
				return
			}

			if len(parts[0]) != 36 || len(parts[1]) != 36 {
				t.Errorf("Socket.GetSocketTicket(%v) = %v, each part should be a 36-character UUID", tt.args.session, got)
			}
		})
	}
}

func TestSocket_InitConnection(t *testing.T) {
	type args struct {
		conn    net.Conn
		userSub string
		ticket  string
	}
	tests := []struct {
		name    string
		args    args
		want    func()
		wantErr bool
	}{
		{name: "use ticket to make connection", args: args{interfaces.NewNullConn(), testSocketUserSession.UserSub, testTicket}, wantErr: false},
		{name: "prevents connection with no ticket", args: args{interfaces.NewNullConn(), testSocketUserSession.UserSub, ""}, wantErr: true},
		{name: "prevents connection with malformed ticket", args: args{interfaces.NewNullConn(), testSocketUserSession.UserSub, "a:"}, wantErr: true},
		{name: "prevents connection with wrong ticket", args: args{interfaces.NewNullConn(), testSocketUserSession.UserSub, "a:b"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			response, err := testSocket.SendCommand(CreateSocketConnectionSocketCommand, interfaces.SocketRequest{
				SocketRequestParams: &types.SocketRequestParams{
					UserSub: tt.args.userSub,
					Ticket:  tt.args.ticket,
				},
				Conn: tt.args.conn,
			})

			err = ChannelError(err, response.Error)

			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.InitConnection(%v, %v, %v) error = %v, wantErr %v", tt.args.conn, tt.args.userSub, tt.args.ticket, err, tt.wantErr)
				return
			}

		})
	}
}

func TestSocket_TicketRemovalBehavior(t *testing.T) {
	t.Run("test ticket after renewal", func(t *testing.T) {
		ticket, _ := testSocket.GetSocketTicket(testSocketUserSession)

		response, err := testSocket.SendCommand(CreateSocketConnectionSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: testSocketUserSession.UserSub,
				Ticket:  ticket,
			},
			Conn: interfaces.NewNullConn(),
		})

		err = ChannelError(err, response.Error)
		if err != nil {
			t.Errorf("Failed with first ticket: %v", err)
		}

		response, err = testSocket.SendCommand(CreateSocketConnectionSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: testSocketUserSession.UserSub,
				Ticket:  ticket,
			},
			Conn: interfaces.NewNullConn(),
		})

		err = ChannelError(err, response.Error)
		if err == nil {
			t.Error("Able to reuse ticket")
		}
	})
}

func TestSocket_SendMessageBytes(t *testing.T) {
	userSub := "user-sub"
	type args struct {
		userSub      string
		targets      string
		messageBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "send a basic message", args: args{userSub, connId, []byte("PING")}, wantErr: false},
		{name: "errors when no targets provided", args: args{userSub, "", []byte("PING")}, wantErr: true},
		{name: "errors when no user sub provided", args: args{"", connId, []byte("PING")}, wantErr: true},
		{name: "errors when no targets found", args: args{userSub, "bad-target", []byte("PING")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testSocket.SendMessageBytes(tt.args.userSub, tt.args.targets, tt.args.messageBytes); (err != nil) != tt.wantErr {
				t.Errorf("Socket.SendMessageBytes(%v, %v) error = %v, wantErr %v", tt.args.messageBytes, tt.args.targets, err, tt.wantErr)
			}
		})
	}
}

func TestSocket_SendMessage(t *testing.T) {
	userSub := "user-sub"
	type args struct {
		userSub string
		targets string
		message *types.SocketMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "sends normal message", args: args{userSub, connId, testSocketMessage}, wantErr: false},
		{name: "sends partial message", args: args{userSub, connId, &types.SocketMessage{}}, wantErr: false},
		{name: "errors when no message", args: args{userSub, connId, nil}, wantErr: true},
		{name: "errors when no targets", args: args{userSub, "", testSocketMessage}, wantErr: true},
		{name: "errors when no user sub", args: args{"", connId, testSocketMessage}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testSocket.SendMessage(tt.args.userSub, tt.args.targets, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Socket.SendMessage(%v, %v) error = %v, wantErr %v", tt.args.message, tt.args.targets, err, tt.wantErr)
			}
		})
	}
}

func TestSocket_RoleCall(t *testing.T) {
	type args struct {
		userSub string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testSocket.RoleCall(tt.args.userSub); (err != nil) != tt.wantErr {
				t.Errorf("Socket.RoleCall(%v) error = %v, wantErr %v", tt.args.userSub, err, tt.wantErr)
			}
		})
	}
}

//	func TestSocket_GetCommandChannel(t *testing.T) {
//		type fields struct {
//			Ch chan<- SocketCommand
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   chan<- SocketCommand
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				s := &Socket{
//					Ch: tt.fields.Ch,
//				}
//				if got := s.GetCommandChannel(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("Socket.GetCommandChannel() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
func TestSocket_SendCommand(t *testing.T) {
	type args struct {
		cmd     interfaces.SocketCommand
		request interfaces.SocketRequest
	}

	// Create a struct to hold dynamic values that will be shared between tests
	var state struct {
		ticket       string
		connId       string
		auth         string
		subscriberId string
	}

	state.subscriberId = "user123"

	tests := []struct {
		name     string
		args     args
		want     interfaces.SocketResponse
		wantErr  bool
		validate func(t *testing.T, got interfaces.SocketResponse) bool
	}{
		{
			name: "Create ticket success",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: CreateSocketTicketSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
						GroupId: "group-id",
						Roles:   "APP_GROUP_ADMIN",
					},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				if got.SocketResponseParams == nil || got.SocketResponseParams.Ticket == "" {
					t.Errorf("Expected ticket in response, got none")
					return false
				}

				// Save the ticket for later tests
				state.ticket = got.SocketResponseParams.Ticket
				parts := strings.Split(state.ticket, ":")
				if len(parts) != 2 {
					t.Errorf("Invalid ticket format: %s", state.ticket)
					return false
				}
				state.auth = parts[0]
				state.connId = parts[1]

				return true
			},
		},
		{
			name: "Create connection success",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: CreateSocketConnectionSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
					},
					Conn: &interfaces.NullConn{},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				if got.Error != nil {
					t.Errorf("Unexpected error: %v", got.Error)
					return false
				}

				if got.SocketResponseParams == nil || got.SocketResponseParams.Subscriber == nil {
					t.Errorf("Expected subscriber in response, got none")
					return false
				}

				sub := got.SocketResponseParams.Subscriber
				if sub.UserSub != state.subscriberId {
					t.Errorf("Expected subscriber ID %s, got %s", state.subscriberId, sub.UserSub)
					return false
				}

				if !strings.Contains(sub.ConnectionIds, state.connId) {
					t.Errorf("Expected connection ID %s in subscriber, got %s", state.connId, sub.ConnectionIds)
					return false
				}

				return true
			},
		},
		{
			name: "Add subscribed topic",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: AddSubscribedTopicSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
						Topic:   "notifications",
					},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				return got.Error == nil
			},
		},
		{
			name: "Has subscribed topic success",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: HasSubscribedTopicSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
						Topic:   "notifications",
					},
				},
			},
			want: interfaces.SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					HasSub: true,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				if got.Error != nil {
					t.Errorf("Unexpected error: %v", got.Error)
					return false
				}

				if got.SocketResponseParams == nil || !got.SocketResponseParams.HasSub {
					t.Errorf("Expected HasSub to be true, got false or nil")
					return false
				}

				return true
			},
		},
		{
			name: "Send socket message success",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: SendSocketMessageSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub:      state.subscriberId,
						MessageBytes: []byte("test message"),
					},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				return got.Error == nil
			},
		},
		{
			name: "Get subscribed targets success",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: GetSubscribedTargetsSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
					},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				if got.Error != nil {
					t.Errorf("Unexpected error: %v", got.Error)
					return false
				}

				if got.SocketResponseParams == nil || got.SocketResponseParams.Targets == "" {
					t.Errorf("Expected targets in response, got none, %+v", got)
					return false
				}

				if !strings.Contains(got.SocketResponseParams.Targets, state.connId) {
					t.Errorf("Expected connection ID %s in targets, got %s", state.connId, got.SocketResponseParams.Targets)
					return false
				}

				return true
			},
		},
		{
			name: "Delete subscribed topic",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: DeleteSubscribedTopicSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
						Topic:   "notifications",
					},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				return got.Error == nil
			},
		},
		{
			name: "Get subscriber success",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: GetSubscriberSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
						// Ticket added during test
					},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				if got.Error != nil {
					t.Errorf("Unexpected error: %v", got.Error)
					return false
				}

				if got.SocketResponseParams == nil || got.SocketResponseParams.Subscriber == nil {
					t.Errorf("Expected subscriber in response, got none")
					return false
				}

				sub := got.SocketResponseParams.Subscriber
				if sub.UserSub != state.subscriberId {
					t.Errorf("Expected subscriber ID %s, got %s", state.subscriberId, sub.UserSub)
					return false
				}

				return true
			},
		},
		{
			name: "Delete connection success",
			args: args{
				cmd: interfaces.SocketCommand{
					SocketCommandParams: &types.SocketCommandParams{
						Ty: DeleteSocketConnectionSocketCommand,
					},
				},
				request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: state.subscriberId,
					},
				},
			},
			want:    interfaces.SocketResponse{},
			wantErr: false,
			validate: func(t *testing.T, got interfaces.SocketResponse) bool {
				return got.Error == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.args.cmd.Ty {
			case CreateSocketConnectionSocketCommand:
				tt.args.request.SocketRequestParams.Ticket = state.ticket
			case AddSubscribedTopicSocketCommand:
				tt.args.request.SocketRequestParams.Targets = state.connId
			case SendSocketMessageSocketCommand:
				tt.args.request.SocketRequestParams.Targets = state.connId
			case GetSubscriberSocketCommand:
				tt.args.request.SocketRequestParams.Ticket = state.auth + ":" + state.connId
			case DeleteSocketConnectionSocketCommand:
				tt.args.request.SocketRequestParams.Ticket = state.auth + ":" + state.connId
			}

			got, err := testSocket.SendCommand(tt.args.cmd.Ty, tt.args.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.SendCommand(%v, %v) error = %v, wantErr %v", tt.args.cmd, tt.args.request, err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				if !tt.validate(t, got) {
					t.Errorf("Socket.SendCommand(%v, %v) validation failed", tt.args.cmd, tt.args.request)
				}
			}
		})
	}
}
