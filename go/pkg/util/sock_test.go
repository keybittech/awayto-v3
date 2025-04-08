package util

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestGetColonJoined(t *testing.T) {
	type args struct {
		userSub string
		connId  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty values", args: args{userSub: "", connId: ""}, want: ":"},
		{name: "non-nil error", args: args{userSub: "a", connId: "b"}, want: "a:b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetColonJoined(tt.args.userSub, tt.args.connId); got != tt.want {
				t.Errorf("GetSocketId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkGetSocketId(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = GetColonJoined("a", "b")
	}
}

func BenchmarkGetSocketIdNegative(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = GetColonJoined("", "")
	}
}

func TestSplitColonJoined(t *testing.T) {
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
		{name: "empty values", args: args{id: ""}, want: "", want1: "", wantErr: true},
		{name: "valid socket id", args: args{id: "a:b"}, want: "a", want1: "b", wantErr: false},
		{name: "id with no end", args: args{id: "a:"}, want: "", want1: "", wantErr: true},
		{name: "id with no colon", args: args{id: "abc"}, want: "", want1: "", wantErr: true},
		{name: "id with multiple colons", args: args{id: "a:b:c"}, want: "a", want1: "b:c", wantErr: false},
		{name: "id with leading space", args: args{id: " a:b"}, want: " a", want1: "b", wantErr: false},
		{name: "id with trailing space", args: args{id: "a:b "}, want: "a", want1: "b ", wantErr: false},
		{name: "id with only colon", args: args{id: ":"}, want: "", want1: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := SplitColonJoined(tt.args.id)
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

func BenchmarkSplitColonJoined(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_, _, _ = SplitColonJoined("a:b")
	}
}

func BenchmarkSplitColonJoinedNegative(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_, _, _ = SplitColonJoined("")
	}
}

func BenchmarkSplitColonJoinedNegativeColon(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_, _, _ = SplitColonJoined("a:")
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
		{name: "no values", args: args{clientKey: ""}, want: "Kfh9QIsMVZcl6xEPYxPHzW8SZ8w="},
		{name: "test value", args: args{clientKey: "test"}, want: "tNpbgC8ZQDOcSkHAWopKzQjJ1hI="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComputeWebSocketAcceptKey(tt.args.clientKey); got != tt.want {
				t.Errorf("ComputeWebSocketAcceptKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkComputeWebSocketAcceptKey(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = ComputeWebSocketAcceptKey("test")
	}
}

type MockConn struct {
	net.Conn
	data []byte
}

func (m *MockConn) Read(data []byte) (int, error) {
	// Copy data from internal buffer into the provided data slice
	// Simulating a "read" by copying data into the provided slice until it's full or there's no more data
	n := len(data)
	if len(m.data) == 0 {
		return 0, io.EOF // Return EOF when there's no more data to read
	}
	if n > len(m.data) {
		n = len(m.data)
	}
	copy(data[:n], m.data[:n])
	m.data = m.data[n:] // Remove the read part of the data

	return n, nil
}

func (m *MockConn) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	return len(p), nil
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
		{
			name: "Valid message with mask",
			args: args{
				conn: &MockConn{
					data: []byte{
						0x81,                   // first byte (fin, opcode)
						0x85,                   // second byte (mask, payload length)
						0x00, 0x00, 0x00, 0x00, // mask key (0x00000000 for simplicity)
						0x48, 0x65, 0x6c, 0x6c, 0x6f, // payload data "Hello" (XOR'd with the mask)
					},
				},
			},
			want:    []byte("Hello"),
			wantErr: false,
		},
		{
			name: "Empty message",
			args: args{
				conn: &MockConn{
					data: []byte{},
				},
			},
			want:    nil,
			wantErr: true,
		},
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

func BenchmarkReadSocketConnectionMessage(b *testing.B) {
	conn := &MockConn{
		data: []byte{
			0x81,                   // first byte (fin, opcode)
			0x85,                   // second byte (mask, payload length)
			0x00, 0x00, 0x00, 0x00, // mask key (0x00000000 for simplicity)
			0x48, 0x65, 0x6c, 0x6c, 0x6f, // payload data "Hello" (XOR'd with the mask)
		},
	}
	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = ReadSocketConnectionMessage(conn)
	}
}

type MockErrConn struct {
	net.Conn
	data []byte
}

func (m *MockErrConn) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write error")
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
		{
			name: "Write error on connection",
			args: args{
				msg:  []byte("test"),
				conn: &MockErrConn{},
			},
			wantErr: true,
		},
		{
			name: "Zero message",
			args: args{
				msg:  []byte{},
				conn: &MockConn{},
			},
			wantErr: false,
		},
		{
			name: "Valid message (small payload)",
			args: args{
				msg:  []byte("Hello"),
				conn: &MockConn{},
			},
			wantErr: false,
		},
		{
			name: "Valid message (extended length)",
			args: args{
				msg:  make([]byte, 130), // payload larger than 125, so it should use extended length encoding
				conn: &MockConn{},
			},
			wantErr: false,
		},
		{
			name: "Valid message (large payload)",
			args: args{
				msg:  make([]byte, 70000), // payload larger than 65535, it should use 8-byte extended length encoding
				conn: &MockConn{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteSocketConnectionMessage(tt.args.msg, tt.args.conn); (err != nil) != tt.wantErr {
				t.Errorf("WriteSocketConnectionMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkWriteSocketConnectionMessage(b *testing.B) {
	data := []byte("test")
	conn := &MockConn{}
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = WriteSocketConnectionMessage(data, conn)
	}
}

func BenchmarkWriteSocketConnectionMessageLarge(b *testing.B) {
	data := make([]byte, 70000)
	conn := &MockConn{}
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = WriteSocketConnectionMessage(data, conn)
	}
}

func BenchmarkWriteSocketConnectionMessageError(b *testing.B) {
	data := []byte("test")
	conn := &MockErrConn{}
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = WriteSocketConnectionMessage(data, conn)
	}
}

type PayloadData struct {
	InnerData string
}

type TestPayload struct {
	Id   string
	Data PayloadData
}

func TestGenerateMessage(t *testing.T) {
	testMessage := types.SocketMessage{
		Action:     44,
		Topic:      "topic",
		Payload:    "payload",
		Store:      false,
		Historical: false,
		Sender:     "sender",
		Timestamp:  "timestamp",
	}
	storedMessage := testMessage
	storedMessage.Store = true
	historicalMessage := testMessage
	historicalMessage.Historical = true
	badActionMessage := testMessage
	badActionMessage.Action = -44
	type args struct {
		padTo   int
		message *types.SocketMessage
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "standard message", args: args{5, &testMessage}, want: []byte("000024400001f00001f00009timestamp00005topic00006sender00007payload")},
		{name: "stored message", args: args{5, &storedMessage}, want: []byte("000024400001t00001f00009timestamp00005topic00006sender00007payload")},
		{name: "historical message", args: args{5, &historicalMessage}, want: []byte("000024400001f00001t00009timestamp00005topic00006sender00007payload")},
		{name: "bad action message", args: args{5, &badActionMessage}, want: []byte("00003-4400001f00001f00009timestamp00005topic00006sender00007payload")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateMessage(tt.args.padTo, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkGenerateMessage(b *testing.B) {
	padTo := 5
	testMessage := &types.SocketMessage{
		Action:     44,
		Topic:      "topic",
		Payload:    "payload",
		Store:      false,
		Historical: false,
		Sender:     "sender",
		Timestamp:  "timestamp",
	}
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = GenerateMessage(padTo, testMessage)
	}
}

func TestParseMessage(t *testing.T) {
	message := []byte("000024400001f00001f00009timestamp00005topic00006sender00007payload")
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
		{
			name: "parses action",
			args: args{
				padTo:  5,
				cursor: 0,
				data:   message,
			},
			want:    7,    // Expected end of the first parsed value
			want1:   "44", // First value
			wantErr: false,
		},
		{
			name: "parses store",
			args: args{
				padTo:  5,
				cursor: 7,
				data:   message,
			},
			want:    13,  // Expected end after second value
			want1:   "f", // Second value
			wantErr: false,
		},
		{
			name: "parses historical",
			args: args{
				padTo:  5,
				cursor: 13,
				data:   message,
			},
			want:    19,  // Expected end after third value
			want1:   "f", // Third value
			wantErr: false,
		},
		{
			name: "parses timestamp",
			args: args{
				padTo:  5,
				cursor: 19,
				data:   message,
			},
			want:    33,          // Expected end after timestamp value
			want1:   "timestamp", // Timestamp value
			wantErr: false,
		},
		{
			name: "parses topic",
			args: args{
				padTo:  5,
				cursor: 33,
				data:   message,
			},
			want:    43,      // Expected end after topic value
			want1:   "topic", // Topic value
			wantErr: false,
		},
		{
			name: "parses sender",
			args: args{
				padTo:  5,
				cursor: 43,
				data:   message,
			},
			want:    54,       // Expected end after sender value
			want1:   "sender", // Sender value
			wantErr: false,
		},
		{
			name: "parses payload",
			args: args{
				padTo:  5,
				cursor: 54,
				data:   message,
			},
			want:    66,        // Expected end after payload value
			want1:   "payload", // Payload value
			wantErr: false,
		},
		{
			name: "data too short",
			args: args{
				padTo:  5,
				cursor: 0,
				data:   []byte("000094"), // 00009 means the value coming after needs to be 9 characters
			},
			want:    0,
			want1:   "",
			wantErr: true,
		},
		{
			name: "out of range value index",
			args: args{
				padTo:  5,
				cursor: 99, // greater index than the data provided
				data:   []byte("000024400001f00001f00009timestamp"),
			},
			want:    0,
			want1:   "",
			wantErr: true,
		},
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

func BenchmarkParseMessage(b *testing.B) {
	padTo := 5
	cursor := 0
	message := []byte("000024400001f00001f00009timestamp00005topic00006sender00007payload")
	reset(b)
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseMessage(padTo, cursor, message)
	}
}
