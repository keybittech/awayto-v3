package api

import (
	"net/http"
	"testing"
)

func TestAPI_InitSockServer(t *testing.T) {
	type args struct {
		mux *http.ServeMux
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
			tt.a.InitSockServer(tt.args.mux)
		})
	}
}

func TestAPI_TearDownSocketConnection(t *testing.T) {
	type args struct {
		socketId string
		connId   string
		userSub  string
	}
	tests := []struct {
		name string
		a    *API
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.a.TearDownSocketConnection(tt.args.socketId, tt.args.connId, tt.args.userSub); got != tt.want {
			// 	t.Errorf("API.TearDownSocketConnection(%v, %v, %v) = %v, want %v", tt.args.socketId, tt.args.connId, tt.args.userSub, got, tt.want)
			// }
		})
	}
}
