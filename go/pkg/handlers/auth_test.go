package handlers

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_AuthWebhook_REGISTER(t *testing.T) {
	type args struct {
		req       *http.Request
		authEvent *types.AuthEvent
		session   *types.UserSession
		tx        *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.AuthWebhook_REGISTER(tt.args.req, tt.args.authEvent, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.AuthWebhook_REGISTER(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.req, tt.args.authEvent, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Handlers.AuthWebhook_REGISTER(%v, %v, %v, %v) = %v, want %v", tt.args.req, tt.args.authEvent, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_AuthWebhook_REGISTER_VALIDATE(t *testing.T) {
	type args struct {
		req       *http.Request
		authEvent *types.AuthEvent
		session   *types.UserSession
		tx        *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.AuthWebhook_REGISTER_VALIDATE(tt.args.req, tt.args.authEvent, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.AuthWebhook_REGISTER_VALIDATE(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.req, tt.args.authEvent, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Handlers.AuthWebhook_REGISTER_VALIDATE(%v, %v, %v, %v) = %v, want %v", tt.args.req, tt.args.authEvent, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}
