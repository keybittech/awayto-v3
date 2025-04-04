package util

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestUnmarshalProto(t *testing.T) {
	type args struct {
		req *http.Request
		pb  protoreflect.ProtoMessage
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
			if err := UnmarshalProto(tt.args.req, tt.args.pb); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalProto() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecomposeProto(t *testing.T) {
	type args struct {
		msg proto.Message
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 []string
		want2 []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := DecomposeProto(tt.args.msg)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecomposeProto() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("DecomposeProto() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("DecomposeProto() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestParseHandlerOptions(t *testing.T) {
	type args struct {
		md protoreflect.MethodDescriptor
	}
	tests := []struct {
		name string
		args args
		want *HandlerOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseHandlerOptions(tt.args.md); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseHandlerOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseProtoQueryParams(t *testing.T) {
	type args struct {
		pbVal       reflect.Value
		queryParams url.Values
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ParseProtoQueryParams(tt.args.pbVal, tt.args.queryParams)
		})
	}
}

func TestParseProtoPathParams(t *testing.T) {
	type args struct {
		pbVal             reflect.Value
		methodParameters  []string
		requestParameters []string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ParseProtoPathParams(tt.args.pbVal, tt.args.methodParameters, tt.args.requestParameters)
		})
	}
}
