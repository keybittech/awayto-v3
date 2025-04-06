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

func getClientCommands(numClients int) map[int]func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand {
	createCommands := make(map[int]func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand)

	for c := 0; c < numClients; c++ {
		clientId := "client-" + strconv.Itoa(c)
		createCommands[c] = func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand {
			return interfaces.SocketCommand{
				SocketCommandParams: &types.SocketCommandParams{
					Ty: CreateSocketTicketSocketCommand,
				},
				Request: interfaces.SocketRequest{
					SocketRequestParams: &types.SocketRequestParams{
						UserSub: clientId,
						Targets: "target",
					},
				},
				ReplyChan: replyChan,
			}
		}
	}
	return createCommands
}

// Benchmark the worker pool
func BenchmarkInitSocket1Client1Messages(b *testing.B) {
	clientCount := 1
	commandsPerClient := 1

	createCommands := getClientCommands(clientCount)

	socket := InitSocket()
	reset(b)
	for i := 0; i < b.N; i++ {

		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					SendCommand(socket, createCommands[c])
				}
			}()
		}

		wg.Wait()
	}

	socket.Close()
}

// Benchmark the worker pool
func BenchmarkInitSocket1Client100Messages(b *testing.B) {
	clientCount := 1
	commandsPerClient := 100

	createCommands := getClientCommands(clientCount)

	socket := InitSocket()
	reset(b)
	for i := 0; i < b.N; i++ {

		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					SendCommand(socket, createCommands[c])
				}
			}()
		}

		wg.Wait()
	}

	socket.Close()
}

// Benchmark the worker pool
func BenchmarkInitSocket10Clients1Message(b *testing.B) {
	clientCount := 10
	commandsPerClient := 1

	createCommands := getClientCommands(clientCount)

	socket := InitSocket()
	reset(b)
	for i := 0; i < b.N; i++ {

		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					SendCommand(socket, createCommands[c])
				}
			}()
		}

		wg.Wait()
	}

	socket.Close()
}

// Benchmark the worker pool
func BenchmarkInitSocket10Clients100Messages(b *testing.B) {
	clientCount := 10
	commandsPerClient := 100

	createCommands := getClientCommands(clientCount)

	socket := InitSocket()
	reset(b)
	for i := 0; i < b.N; i++ {

		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					SendCommand(socket, createCommands[c])
				}
			}()
		}

		wg.Wait()
	}

	socket.Close()
}

// Benchmark the worker pool
func BenchmarkInitSocket100Clients1Message(b *testing.B) {
	clientCount := 100
	commandsPerClient := 1

	createCommands := getClientCommands(clientCount)

	socket := InitSocket()
	reset(b)
	for i := 0; i < b.N; i++ {

		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					SendCommand(socket, createCommands[c])
				}
			}()
		}

		wg.Wait()
	}

	socket.Close()
}

// Benchmark the worker pool
func BenchmarkInitSocket100Clients100Messages(b *testing.B) {
	clientCount := 100
	commandsPerClient := 100

	createCommands := getClientCommands(clientCount)

	socket := InitSocket()
	reset(b)
	for i := 0; i < b.N; i++ {

		var wg sync.WaitGroup

		for c := 0; c < clientCount; c++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					SendCommand(socket, createCommands[c])
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

// func TestSocket_GetCommandChannel(t *testing.T) {
// 	type fields struct {
// 		Ch chan<- SocketCommand
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   chan<- SocketCommand
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &Socket{
// 				Ch: tt.fields.Ch,
// 			}
// 			if got := s.GetCommandChannel(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Socket.GetCommandChannel() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestSocket_SendCommand(t *testing.T) {
// 	type fields struct {
// 		Ch chan<- SocketCommand
// 	}
// 	type args struct {
// 		cmdType SocketCommandType
// 		params  SocketParams
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    SocketResponse
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &Socket{
// 				Ch: tt.fields.Ch,
// 			}
// 			got, err := s.SendCommand(tt.args.cmdType, tt.args.params)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Socket.SendCommand(%v, %v) error = %v, wantErr %v", tt.args.cmdType, tt.args.params, err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Socket.SendCommand(%v, %v) = %v, want %v", tt.args.cmdType, tt.args.params, got, tt.want)
// 			}
// 		})
// 	}
// }
