package api

import (
	"net"
	"testing"
)

func TestAPI_InitTurnTCP(t *testing.T) {
	type args struct {
		listenerPort int
		internalPort int
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
			tt.a.InitTurnTCP(tt.args.listenerPort, tt.args.internalPort)
		})
	}
}

func TestHandleTurnTCPConnection(t *testing.T) {
	type args struct {
		conn         net.Conn
		internalPort int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleTurnTCPConnection(tt.args.conn, tt.args.internalPort)
		})
	}
}

func TestAPI_InitTurnUDP(t *testing.T) {
	type args struct {
		listenerPort int
		internalPort int
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
			tt.a.InitTurnUDP(tt.args.listenerPort, tt.args.internalPort)
		})
	}
}

func TestHandleTurnUDPConnection(t *testing.T) {
	type args struct {
		listenerConn *net.UDPConn
		internalConn *net.UDPConn
		listenerAddr *net.UDPAddr
		internalAddr *net.UDPAddr
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleTurnUDPConnection(tt.args.listenerConn, tt.args.internalConn, tt.args.listenerAddr, tt.args.internalAddr)
		})
	}
}
