package api

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func TestAPI_HandleRequest(t *testing.T) {
	type args struct {
		handlerOpts *util.HandlerOptions
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
			if got := tt.a.HandleRequest(tt.args.handlerOpts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.HandleRequest(%v) = %v, want %v", tt.args.handlerOpts, got, tt.want)
			}
		})
	}
}
