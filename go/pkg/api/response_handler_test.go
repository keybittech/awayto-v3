package api

import (
	"net/http"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestProtoResponseHandler(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		results proto.Message
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProtoResponseHandler(tt.args.w, tt.args.results)
			if got != tt.want {
				t.Errorf("ProtoResponseHandler(%v, %v) = %v, want %v", tt.args.w, tt.args.results, got, tt.want)
			}
		})
	}
}

func TestMultipartResponseHandler(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		results proto.Message
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MultipartResponseHandler(tt.args.w, tt.args.results)
			if got != tt.want {
				t.Errorf("MultipartResponseHandler(%v, %v) = %v, want %v", tt.args.w, tt.args.results, got, tt.want)
			}
		})
	}
}
