package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_AuthWebhook_REGISTER(t *testing.T) {
	type args struct {
		info      ReqInfo
		authEvent *types.AuthEvent
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.AuthWebhookResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.AuthWebhook_REGISTER(tt.args.info, tt.args.authEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.AuthWebhook_REGISTER(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.authEvent, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.AuthWebhook_REGISTER(%v, %v) = %v, want %v", tt.args.info, tt.args.authEvent, got, tt.want)
			}
		})
	}
}

func TestHandlers_AuthWebhook_REGISTER_VALIDATE(t *testing.T) {
	type args struct {
		info      ReqInfo
		authEvent *types.AuthEvent
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.AuthWebhookResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.AuthWebhook_REGISTER_VALIDATE(tt.args.info, tt.args.authEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.AuthWebhook_REGISTER_VALIDATE(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.authEvent, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.AuthWebhook_REGISTER_VALIDATE(%v, %v) = %v, want %v", tt.args.info, tt.args.authEvent, got, tt.want)
			}
		})
	}
}
