package clients

import (
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

var testSocket *Socket
var testSocketUserSession *types.UserSession
var testSocketMessage *types.SocketMessage
var testTicket, auth, testConnId string

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
	_, testConnId, err = util.SplitColonJoined(testTicket)
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

func getClientData(numClients int, commandType int32, requestParams *types.SocketRequestParams, conn ...bool) ([]string, map[string]func(replyChan chan SocketResponse) SocketCommand) {
	clientIds := make([]string, numClients)
	clientCommands := make(map[string]func(replyChan chan SocketResponse) SocketCommand)
	for i := 0; i < numClients; i++ {
		clientId := "client-" + strconv.Itoa(numClients) + "-" + strconv.Itoa(i)
		clientIds[i] = clientId
		clientCommands[clientId] = func(replyChan chan SocketResponse) SocketCommand {
			requestParams.UserSub = clientId
			request := SocketRequest{SocketRequestParams: requestParams}
			if conn != nil {
				request.Conn = &util.NullConn{}
			}
			return SocketCommand{
				WorkerCommandParams: &types.WorkerCommandParams{
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

// Benchmark the worker pool
func doSendCommandBench(clientCount, commandsPerClient int, b *testing.B) {
	socket := InitSocket()
	defer socket.Close()
	clientIds, createCommands := getClientData(
		clientCount,
		CreateSocketTicketSocketCommand,
		&types.SocketRequestParams{
			GroupId: "group-id",
			Roles:   "APP_GROUP_ADMIN",
		},
	)

	totalIterations := b.N / (clientCount * commandsPerClient)
	if totalIterations < 1 {
		totalIterations = 1
	}

	b.SetParallelism(clientCount)
	reset(b)

	b.RunParallel(func(pb *testing.PB) {
		clientIndex := int(atomic.AddInt32(new(int32), 1)-1) % clientCount

		iteration := 0
		for pb.Next() {
			if iteration >= totalIterations {
				break
			}

			for j := 0; j < commandsPerClient; j++ {
				SendCommand(socket, createCommands[clientIds[clientIndex]])
			}

			iteration++
		}
	})
}

func BenchmarkSendCommandSingle1Client(b *testing.B) {
	doSendCommandBench(1, 1, b)
}

func BenchmarkSendCommandSingle10Clients(b *testing.B) {
	doSendCommandBench(10, 1, b)
}

func BenchmarkSendCommandSingle100Clients(b *testing.B) {
	doSendCommandBench(100, 1, b)
}

func BenchmarkSendCommandSingle1000Clients(b *testing.B) {
	doSendCommandBench(1000, 1, b)
}

func BenchmarkSendCommand1Client1Message(b *testing.B) {
	doSendCommandBench(1, 1, b)
}

func BenchmarkSendCommand1Client10Messages(b *testing.B) {
	doSendCommandBench(1, 10, b)
}

func BenchmarkSendCommand1Client100Messages(b *testing.B) {
	doSendCommandBench(1, 100, b)
}

func BenchmarkSendCommand10Clients1Message(b *testing.B) {
	doSendCommandBench(10, 1, b)
}

func BenchmarkSendCommand10Clients10Messages(b *testing.B) {
	doSendCommandBench(10, 10, b)
}

func BenchmarkSendCommand10Clients100Messages(b *testing.B) {
	doSendCommandBench(10, 100, b)
}

func BenchmarkSendCommand100Clients1Message(b *testing.B) {
	doSendCommandBench(100, 1, b)
}

func BenchmarkSendCommand100Clients10Messages(b *testing.B) {
	doSendCommandBench(100, 10, b)
}

func BenchmarkSendCommand100Clients100Messages(b *testing.B) {
	doSendCommandBench(100, 100, b)
}

func BenchmarkSendCommand1000Clients1Message(b *testing.B) {
	doSendCommandBench(1000, 1, b)
}

func BenchmarkSendCommand1000Clients10Messages(b *testing.B) {
	doSendCommandBench(1000, 10, b)
}

func BenchmarkSendCommand1000Clients100Messages(b *testing.B) {
	doSendCommandBench(1000, 100, b)
}

func TestSocket_GetSocketTicket(t *testing.T) {
	t.Parallel()
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
	initSession := testSocketUserSession
	initSession.UserSub = "user-init-conn"
	ticket, err := testSocket.GetSocketTicket(initSession)
	if err != nil {
		t.Fatal(err)
	}

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
		{name: "use ticket to make connection", args: args{util.NewNullConn(), initSession.UserSub, ticket}, wantErr: false},
		{name: "requires a connection object", args: args{nil, "test-user", "a:b"}, wantErr: true},
		{name: "prevents connection with no ticket", args: args{util.NewNullConn(), initSession.UserSub, ""}, wantErr: true},
		{name: "prevents connection with malformed ticket", args: args{util.NewNullConn(), initSession.UserSub, "a:"}, wantErr: true},
		{name: "prevents connection with wrong ticket", args: args{util.NewNullConn(), initSession.UserSub, "a:b"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			response, err := testSocket.SendCommand(CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
				UserSub: tt.args.userSub,
				Ticket:  tt.args.ticket,
			}, tt.args.conn)

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

		response, err := testSocket.SendCommand(CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
			UserSub: testSocketUserSession.UserSub,
			Ticket:  ticket,
		}, util.NullConn{})

		err = ChannelError(err, response.Error)
		if err != nil {
			t.Errorf("Failed with first ticket: %v", err)
		}

		response, err = testSocket.SendCommand(CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
			UserSub: testSocketUserSession.UserSub,
			Ticket:  ticket,
		}, util.NullConn{})

		err = ChannelError(err, response.Error)
		if err == nil {
			t.Error("Able to reuse ticket")
		}
	})
}

func getNewConnection(t *testing.T, userSub string) (*types.UserSession, string, string) {
	newSession := testSocketUserSession
	newSession.UserSub = userSub
	ticket, err := testSocket.GetSocketTicket(newSession)
	if err != nil {
		t.Fatal(err)
	}
	response, err := testSocket.SendCommand(CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
		UserSub: testSocketUserSession.UserSub,
		Ticket:  ticket,
	}, util.NullConn{})

	err = ChannelError(err, response.Error)
	if err != nil {
		t.Errorf("Failed with first ticket: %v", err)
	}

	_, connId, err := util.SplitColonJoined(ticket)
	if err != nil {
		t.Fatal(err)
	}

	return newSession, ticket, connId
}

func TestSocket_SendMessageBytes(t *testing.T) {
	session, _, connId := getNewConnection(t, "user-send-message-bytes")
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
		{name: "send a basic message", args: args{session.UserSub, connId, []byte("PING")}, wantErr: false},
		{name: "errors when no targets provided", args: args{session.UserSub, "", []byte("PING")}, wantErr: true},
		{name: "errors when no user sub provided", args: args{"", connId, []byte("PING")}, wantErr: true},
		{name: "errors when no targets found", args: args{session.UserSub, "bad-target", []byte("PING")}, wantErr: true},
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
	session, _, connId := getNewConnection(t, "user-send-message")
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
		{name: "sends normal message", args: args{session.UserSub, connId, testSocketMessage}, wantErr: false},
		{name: "sends partial message", args: args{session.UserSub, connId, &types.SocketMessage{}}, wantErr: false},
		{name: "errors when no message", args: args{session.UserSub, connId, nil}, wantErr: true},
		{name: "errors when no targets", args: args{session.UserSub, "", testSocketMessage}, wantErr: true},
		{name: "errors when no user sub", args: args{"", testConnId, testSocketMessage}, wantErr: true},
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
		cmd     SocketCommand
		request *types.SocketRequestParams
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
		want     SocketResponse
		wantErr  bool
		validate func(t *testing.T, got SocketResponse) bool
	}{
		{
			name: "Create ticket success",
			args: args{
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: CreateSocketTicketSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub: state.subscriberId,
					GroupId: "group-id",
					Roles:   "APP_GROUP_ADMIN",
				},
			},
			want:    emptySocketResponse,
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
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
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: CreateSocketConnectionSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub: state.subscriberId,
				},
			},
			want:    emptySocketResponse,
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
				if got.Error != nil {
					t.Errorf("Unexpected error: %v", got.Error)
					return false
				}

				if got.SocketResponseParams == nil || got.SocketResponseParams.UserSub == "" {
					t.Errorf("Expected subscriber userSub in response, got none")
					return false
				}

				if got.SocketResponseParams == nil || got.SocketResponseParams.GroupId == "" {
					t.Errorf("Expected subscriber userSub in response, got none")
					return false
				}

				if got.SocketResponseParams == nil || got.SocketResponseParams.Roles == "" {
					t.Errorf("Expected subscriber userSub in response, got none")
					return false
				}

				if got.SocketResponseParams.UserSub != state.subscriberId {
					t.Errorf("Expected subscriber ID %s, got %s", state.subscriberId, got.SocketResponseParams.UserSub)
					return false
				}

				return true
			},
		},
		{
			name: "Add subscribed topic",
			args: args{
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: AddSubscribedTopicSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub: state.subscriberId,
					Topic:   "notifications",
				},
			},
			want:    emptySocketResponse,
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
				return got.Error == nil
			},
		},
		{
			name: "Has subscribed topic success",
			args: args{
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: HasSubscribedTopicSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub: state.subscriberId,
					Topic:   "notifications",
				},
			},
			want: SocketResponse{
				SocketResponseParams: &types.SocketResponseParams{
					HasSub: true,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
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
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: SendSocketMessageSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub:      state.subscriberId,
					MessageBytes: []byte("test message"),
				},
			},
			want:    emptySocketResponse,
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
				return got.Error == nil
			},
		},
		{
			name: "Get subscribed targets success",
			args: args{
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: GetSubscribedTargetsSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub: state.subscriberId,
				},
			},
			want:    emptySocketResponse,
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
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
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: DeleteSubscribedTopicSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub: state.subscriberId,
					Topic:   "notifications",
				},
			},
			want:    emptySocketResponse,
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
				return got.Error == nil
			},
		},
		{
			name: "Delete connection success",
			args: args{
				cmd: SocketCommand{
					WorkerCommandParams: &types.WorkerCommandParams{
						Ty: DeleteSocketConnectionSocketCommand,
					},
				},
				request: &types.SocketRequestParams{
					UserSub: state.subscriberId,
				},
			},
			want:    emptySocketResponse,
			wantErr: false,
			validate: func(t *testing.T, got SocketResponse) bool {
				return got.Error == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.args.cmd.Ty {
			case CreateSocketConnectionSocketCommand:
				tt.args.request.Ticket = state.ticket
			case AddSubscribedTopicSocketCommand:
				tt.args.request.Targets = state.connId
			case SendSocketMessageSocketCommand:
				tt.args.request.Targets = state.connId
			case DeleteSocketConnectionSocketCommand:
				tt.args.request.Ticket = state.auth + ":" + state.connId
			default:

			}

			got, err := testSocket.SendCommand(tt.args.cmd.Ty, tt.args.request, &util.NullConn{})

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

func TestNewSocketMaps(t *testing.T) {
	tests := []struct {
		name string
		want *SocketMaps
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSocketMaps(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSocketMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSocket_RouteCommand(t *testing.T) {
	type args struct {
		cmd SocketCommand
	}
	tests := []struct {
		name    string
		s       *Socket
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.RouteCommand(tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("Socket.RouteCommand(%v) error = %v, wantErr %v", tt.args.cmd, err, tt.wantErr)
			}
		})
	}
}

func TestSocket_Close(t *testing.T) {
	tests := []struct {
		name string
		s    *Socket
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Close()
		})
	}
}

func TestSocketCommand_GetClientId(t *testing.T) {
	tests := []struct {
		name string
		cmd  SocketCommand
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.GetClientId(); got != tt.want {
				t.Errorf("SocketCommand.GetClientId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSocketCommand_GetReplyChannel(t *testing.T) {
	tests := []struct {
		name string
		cmd  SocketCommand
		want interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.GetReplyChannel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SocketCommand.GetReplyChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSocket_Connected(t *testing.T) {
	type args struct {
		userSub string
	}
	tests := []struct {
		name string
		s    *Socket
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Connected(tt.args.userSub); got != tt.want {
				t.Errorf("Socket.Connected(%v) = %v, want %v", tt.args.userSub, got, tt.want)
			}
		})
	}
}
