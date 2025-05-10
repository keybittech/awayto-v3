package api

import (
	"net"
	"testing"
)

func TestAPI_InitUnixServer(t *testing.T) {
	type args struct {
		unixPath string
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
			tt.a.InitUnixServer(tt.args.unixPath)
		})
	}
}

func TestAPI_HandleUnixConnection(t *testing.T) {
	type args struct {
		conn net.Conn
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
			tt.a.HandleUnixConnection(tt.args.conn)
		})
	}
}
