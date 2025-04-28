package handlers

import (
	"testing"
)

func TestNewHandlers(t *testing.T) {
	tests := []struct {
		name string
		want *Handlers
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := NewHandlers(); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewHandlers() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestRegister(t *testing.T) {
	type args struct {
		// handler TypedProtoHandler[ReqMsg, ResMsg proto.Message]
	}
	tests := []struct {
		name string
		args args
		want ProtoHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := Register(tt.args.handler); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Register(%v) = %v, want %v", tt.args.handler, got, tt.want)
			// }
		})
	}
}
