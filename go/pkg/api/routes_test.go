package api

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestAPI_HandleRequest(t *testing.T) {
	type args struct {
		serviceMethod protoreflect.MethodDescriptor
	}
	tests := []struct {
		name string
		a    *API
		args args
		want SessionHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.HandleRequest(tt.args.serviceMethod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.HandleRequest(%v) = %v, want %v", tt.args.serviceMethod, got, tt.want)
			}
		})
	}
}
