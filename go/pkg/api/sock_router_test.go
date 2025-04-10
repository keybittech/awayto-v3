package api

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestAPI_SocketMessageReceiver(t *testing.T) {
	type args struct {
		userSub string
		data    []byte
	}
	tests := []struct {
		name string
		a    *API
		args args
		want *types.SocketMessage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.SocketMessageReceiver(tt.args.userSub, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.SocketMessageReceiver(%v, %v) = %v, want %v", tt.args.userSub, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestAPI_SocketMessageRouter(t *testing.T) {
	type args struct {
		sm       *types.SocketMessage
		connId   string
		socketId string
		userSub  string
		groupId  string
		roles    string
	}
	tests := []struct {
		name string
		a    *API
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.SocketMessageRouter(tt.args.sm, tt.args.connId, tt.args.socketId, tt.args.userSub, tt.args.groupId, tt.args.roles)
		})
	}
}
