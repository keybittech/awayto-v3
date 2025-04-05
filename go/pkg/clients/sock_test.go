package clients

import (
	"net"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestInitSocket(t *testing.T) {
	tests := []struct {
		name string
		want interfaces.ISocket
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitSocket(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitSocket() = %v, want %v", got, tt.want)
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
		s       *Socket
		args    args
		want    func()
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.InitConnection(tt.args.conn, tt.args.userSub, tt.args.ticket)
			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.InitConnection(%v, %v, %v) error = %v, wantErr %v", tt.args.conn, tt.args.userSub, tt.args.ticket, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Socket.InitConnection(%v, %v, %v) = %v, want %v", tt.args.conn, tt.args.userSub, tt.args.ticket, got, tt.want)
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
		s       *Socket
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetSocketTicket(tt.args.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.GetSocketTicket(%v) error = %v, wantErr %v", tt.args.session, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Socket.GetSocketTicket(%v) = %v, want %v", tt.args.session, got, tt.want)
			}
		})
	}
}

func TestSocket_SendMessageBytes(t *testing.T) {
	type args struct {
		messageBytes []byte
		targets      string
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
			if err := tt.s.SendMessageBytes(tt.args.messageBytes, tt.args.targets); (err != nil) != tt.wantErr {
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
		s       *Socket
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.SendMessage(tt.args.message, tt.args.targets); (err != nil) != tt.wantErr {
				t.Errorf("Socket.SendMessage(%v, %v) error = %v, wantErr %v", tt.args.message, tt.args.targets, err, tt.wantErr)
			}
		})
	}
}

func TestSocket_GetSubscriberByTicket(t *testing.T) {
	type args struct {
		ticket string
	}
	tests := []struct {
		name    string
		s       *Socket
		args    args
		want    *types.Subscriber
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetSubscriberByTicket(tt.args.ticket)
			if (err != nil) != tt.wantErr {
				t.Errorf("Socket.GetSubscriberByTicket(%v) error = %v, wantErr %v", tt.args.ticket, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Socket.GetSubscriberByTicket(%v) = %v, want %v", tt.args.ticket, got, tt.want)
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
		s    *Socket
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.AddSubscribedTopic(tt.args.userSub, tt.args.topic, tt.args.targets)
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
		s    *Socket
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.DeleteSubscribedTopic(tt.args.userSub, tt.args.topic)
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
		s    *Socket
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.HasTopicSubscription(tt.args.userSub, tt.args.topic); got != tt.want {
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
		s       *Socket
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.RoleCall(tt.args.userSub); (err != nil) != tt.wantErr {
				t.Errorf("Socket.RoleCall(%v) error = %v, wantErr %v", tt.args.userSub, err, tt.wantErr)
			}
		})
	}
}
