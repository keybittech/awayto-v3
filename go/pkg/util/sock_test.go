package util

import (
	"net"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestGetSocketId(t *testing.T) {
	type args struct {
		userSub string
		connId  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSocketId(tt.args.userSub, tt.args.connId); got != tt.want {
				t.Errorf("GetSocketId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitSocketId(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := SplitSocketId(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitSocketId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SplitSocketId() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SplitSocketId() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestComputeWebSocketAcceptKey(t *testing.T) {
	type args struct {
		clientKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComputeWebSocketAcceptKey(tt.args.clientKey); got != tt.want {
				t.Errorf("ComputeWebSocketAcceptKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadSocketConnectionMessage(t *testing.T) {
	type args struct {
		conn net.Conn
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadSocketConnectionMessage(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadSocketConnectionMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadSocketConnectionMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteSocketConnectionMessage(t *testing.T) {
	type args struct {
		msg  []byte
		conn net.Conn
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
			if err := WriteSocketConnectionMessage(tt.args.msg, tt.args.conn); (err != nil) != tt.wantErr {
				t.Errorf("WriteSocketConnectionMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateMessage(t *testing.T) {
	type args struct {
		padTo   int
		message *types.SocketMessage
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateMessage(tt.args.padTo, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	type args struct {
		padTo  int
		cursor int
		data   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseMessage(tt.args.padTo, tt.args.cursor, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseMessage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseMessage() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
