package api

import (
	"net/http"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestProtoBodyParser(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		msgType protoreflect.MessageType
	}
	tests := []struct {
		name    string
		args    args
		want    proto.Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProtoBodyParser(tt.args.w, tt.args.req, tt.args.msgType)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProtoBodyParser(%v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.msgType, got, tt.want)
			}
		})
	}
}

func TestMultipartBodyParser(t *testing.T) {
	type args struct {
		w           http.ResponseWriter
		req         *http.Request
		handlerOpts protoreflect.MessageType
	}
	tests := []struct {
		name    string
		args    args
		want    proto.Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MultipartBodyParser(tt.args.w, tt.args.req, tt.args.handlerOpts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultipartBodyParser(%v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.handlerOpts, got, tt.want)
			}
		})
	}
}
