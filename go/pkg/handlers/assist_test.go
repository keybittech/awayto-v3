package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostPrompt(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PostPromptRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostPromptResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostPrompt(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostPrompt(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostPrompt(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetSuggestion(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetSuggestionRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetSuggestionResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetSuggestion(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetSuggestion(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetSuggestion(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}
