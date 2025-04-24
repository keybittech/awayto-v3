package api

import (
	"context"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestAPI_SocketMessageReceiver(t *testing.T) {
	type args struct {
		data []byte
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
			if got := tt.a.SocketMessageReceiver(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.SocketMessageReceiver(%v) = %v, want %v", tt.args.data, got, tt.want)
			}
		})
	}
}

func TestAPI_SocketMessageRouter(t *testing.T) {
	type args struct {
		ctx      context.Context
		connId   string
		socketId string
		sm       *types.SocketMessage
		ds       clients.DbSession
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
			tt.a.SocketMessageRouter(tt.args.ctx, tt.args.connId, tt.args.socketId, tt.args.sm, tt.args.ds)
		})
	}
}
