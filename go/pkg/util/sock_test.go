package util

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/proto"
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
	for b.Loop() {
		_ = GetColonJoined("a", "b")
	}
}

func BenchmarkGetSocketIdNegative(b *testing.B) {
	reset(b)
	for b.Loop() {
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
	for b.Loop() {
		_, _, _ = SplitColonJoined("a:b")
	}
}

func BenchmarkSplitColonJoinedNegative(b *testing.B) {
	reset(b)
	for b.Loop() {
		_, _, _ = SplitColonJoined("")
	}
}

func BenchmarkSplitColonJoinedNegativeColon(b *testing.B) {
	reset(b)
	for b.Loop() {
		_, _, _ = SplitColonJoined("a:")
	}
}

type PayloadData struct {
	InnerData string
}

type TestPayload struct {
	Data PayloadData
	Id   string
}

func TestGenerateMessage(t *testing.T) {
	testMessage := &types.SocketMessage{
		Action:     44,
		Topic:      "topic",
		Payload:    "payload",
		Store:      false,
		Historical: false,
		Sender:     "sender",
		Timestamp:  "timestamp",
	}
	storedMessage := proto.Clone(testMessage).(*types.SocketMessage)
	storedMessage.Store = true
	historicalMessage := proto.Clone(testMessage).(*types.SocketMessage)
	historicalMessage.Historical = true
	type args struct {
		padTo   int
		message *types.SocketMessage
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "standard message", args: args{5, testMessage}, want: []byte("000024400001f00001f00009timestamp00005topic00006sender00007payload")},
		{name: "stored message", args: args{5, storedMessage}, want: []byte("000024400001t00001f00009timestamp00005topic00006sender00007payload")},
		{name: "historical message", args: args{5, historicalMessage}, want: []byte("000024400001f00001t00009timestamp00005topic00006sender00007payload")},
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
	for b.Loop() {
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
	for b.Loop() {
		_, _, _ = ParseMessage(padTo, cursor, message)
	}
}
