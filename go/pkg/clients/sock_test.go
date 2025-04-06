package clients

import (
	"log"
	"net"
	"strings"
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
	testSocket = InitSocket().(*Socket)
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
			_, err := testSocket.InitConnection(tt.args.conn, tt.args.userSub, tt.args.ticket)
			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.InitConnection(%v, %v, %v) error = %v, wantErr %v", tt.args.conn, tt.args.userSub, tt.args.ticket, err, tt.wantErr)
				return
			}

		})
	}
}

func TestSocket_TicketRemovalBehavior(t *testing.T) {
	t.Run("test ticket after renewal", func(t *testing.T) {
		socket := InitSocket().(*Socket)
		ticket, _ := socket.GetSocketTicket(testSocketUserSession)

		_, err := socket.InitConnection(interfaces.NewNullConn(), testSocketUserSession.UserSub, ticket)
		if err != nil {
			t.Errorf("Failed with first ticket: %v", err)
		}

		_, err = socket.InitConnection(interfaces.NewNullConn(), testSocketUserSession.UserSub, ticket)
		if err == nil {
			t.Error("Able to reuse ticket")
		}
	})
}

func TestSocket_SendMessageBytes(t *testing.T) {
	type args struct {
		messageBytes []byte
		targets      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "send a basic message", args: args{[]byte("PING"), connId}, wantErr: false},
		{name: "errors when no targets provided", args: args{[]byte("PING"), ""}, wantErr: true},
		{name: "errors when no targets found", args: args{[]byte("PING"), "bad-target"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := testSocket.SendMessageBytes(tt.args.messageBytes, tt.args.targets); (err != nil) != tt.wantErr {
				t.Errorf("Socket.SendMessageBytes(%v, %v) error = %v, wantErr %v", tt.args.messageBytes, tt.args.targets, err, tt.wantErr)
			}
		})
	}
}

func TestSocket_SendMessage(t *testing.T) {
	type args struct {
		message *types.SocketMessage
		targets string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "sends normal message", args: args{testSocketMessage, connId}, wantErr: false},
		{name: "sends partial message", args: args{&types.SocketMessage{}, connId}, wantErr: false},
		{name: "requires message object", args: args{nil, connId}, wantErr: true},
		{name: "requires targets to send to", args: args{testSocketMessage, ""}, wantErr: true},
		{name: "requires valid targets to send to", args: args{testSocketMessage, connId}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testSocket.SendMessage(tt.args.message, tt.args.targets); (err != nil) != tt.wantErr {
				t.Errorf("Socket.SendMessage(%v, %v) error = %v, wantErr %v", tt.args.message, tt.args.targets, err, tt.wantErr)
			}
		})
	}
}

func TestSocket_GetSubscriberByTicket(t *testing.T) {
	renewedTicket, _ := testSocket.GetSocketTicket(testSocketUserSession)
	type args struct {
		ticket string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success with valid ticket", args: args{renewedTicket}, wantErr: false},
		{name: "errors with no ticket", args: args{""}, wantErr: true},
		{name: "errors with invalid ticket", args: args{"a:"}, wantErr: true},
		{name: "errors when no target found", args: args{"a:b"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newSocket := testSocket
			if tt.name == "success with valid ticket" {
				newSocket = InitSocket().(*Socket)
				renewedTicket, _ := newSocket.GetSocketTicket(testSocketUserSession)
				tt.args.ticket = renewedTicket
			}
			got, err := newSocket.GetSubscriberByTicket(tt.args.ticket)
			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.GetSubscriberByTicket(%v) error = %v, wantErr %v", tt.args.ticket, err, tt.wantErr)
				return
			}

			if err == nil {
				if got == nil {
					t.Errorf("Socket.GetSubscriberByTicket(%v) returned nil", tt.args.ticket)
				}

				// Verify the returned subscriber has the expected ticket
				if _, ok := got.Tickets[strings.Split(tt.args.ticket, ":")[0]]; !ok {
					t.Errorf("Socket.GetSubscriberByTicket(%v) returned subscriber without matching ticket", tt.args.ticket)
				}
			}
		})
	}
}

func TestSocket_AddSubscribedTopic(t *testing.T) {
	type args struct {
		userSub string
		topic   string
		targets string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSocket.AddSubscribedTopic(tt.args.userSub, tt.args.topic, tt.args.targets)
		})
	}
}

func TestSocket_DeleteSubscribedTopic(t *testing.T) {
	type args struct {
		userSub string
		topic   string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSocket.DeleteSubscribedTopic(tt.args.userSub, tt.args.topic)
		})
	}
}

func TestSocket_HasTopicSubscription(t *testing.T) {
	type args struct {
		userSub string
		topic   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := testSocket.HasTopicSubscription(tt.args.userSub, tt.args.topic); got != tt.want {
				t.Errorf("Socket.HasTopicSubscription(%v, %v) = %v, want %v", tt.args.userSub, tt.args.topic, got, tt.want)
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
