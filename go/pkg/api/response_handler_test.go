package api

import (
	"net/http"
	"reflect"
	"testing"
)

func TestProtoResponseHandler(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		results []reflect.Value
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
			got, err := ProtoResponseHandler(tt.args.w, tt.args.results)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProtoResponseHandler(%v, %v) error = %v, wantErr %v", tt.args.w, tt.args.results, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProtoResponseHandler(%v, %v) = %v, want %v", tt.args.w, tt.args.results, got, tt.want)
			}
		})
	}
}

func TestMultipartResponseHandler(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		results []reflect.Value
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
			got, err := MultipartResponseHandler(tt.args.w, tt.args.results)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultipartResponseHandler(%v, %v) error = %v, wantErr %v", tt.args.w, tt.args.results, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MultipartResponseHandler(%v, %v) = %v, want %v", tt.args.w, tt.args.results, got, tt.want)
			}
		})
	}
}
