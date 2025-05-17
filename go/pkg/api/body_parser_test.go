package api

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestProtoBodyParser(t *testing.T) {
	type args struct {
		w           http.ResponseWriter
		req         *http.Request
		handlerOpts *util.HandlerOptions
	}
	tests := []struct {
		name    string
		args    args
		want    protoreflect.ProtoMessage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProtoBodyParser(tt.args.w, tt.args.req, tt.args.handlerOpts)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProtoBodyParser(%v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.handlerOpts, got, tt.want)
			}
		})
	}
}

func TestMultipartBodyParser(t *testing.T) {
	type args struct {
		w           http.ResponseWriter
		req         *http.Request
		handlerOpts *util.HandlerOptions
	}
	tests := []struct {
		name    string
		args    args
		want    protoreflect.ProtoMessage
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
