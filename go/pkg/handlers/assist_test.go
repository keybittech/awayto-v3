package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestAssist(t *testing.T) {
	tests := []HandlersTestCase{
		{
			name: "GetSuggestion works when group ai is true",
			setupMocks: func(hts *HandlersTestSetup) {
				// mockSession := &types.UserSession{GroupAi: true}
				// hts.MockRedis.EXPECT().ReqSession(gomock.Any()).Return(mockSession).Times(1)
				hts.UserSession.GroupAi = true

				hts.MockAi.EXPECT().GetPromptResponse(gomock.Any(), gomock.Any(), gomock.Any()).Return("one|two|three|four|five", nil)
			},
			handlerFunc: func(h *Handlers, w http.ResponseWriter, r *http.Request, session *types.UserSession, tx *interfaces.MockIDatabaseTx) (interface{}, error) {
				data := &types.GetSuggestionRequest{Id: "3", Prompt: "test-prompt"}
				return h.GetSuggestion(w, r, data, session, tx)
			},
			expectedRes: &types.GetSuggestionResponse{PromptResult: []string{"one", "two", "three", "four", "five"}},
		},
		{
			name: "GetSuggestion returns nothing when group ai is false",
			setupMocks: func(hts *HandlersTestSetup) {
				// mockSession := &types.UserSession{GroupAi: false}
				// hts.MockRedis.EXPECT().ReqSession(gomock.Any()).Return(mockSession)
				hts.UserSession.GroupAi = false
			},
			handlerFunc: func(h *Handlers, w http.ResponseWriter, r *http.Request, session *types.UserSession, tx *interfaces.MockIDatabaseTx) (interface{}, error) {
				data := &types.GetSuggestionRequest{Id: "3", Prompt: "test-prompt"}
				return h.GetSuggestion(w, r, data, session, tx)
			},
			expectedRes: &types.GetSuggestionResponse{PromptResult: []string{}},
		},
	}

	RunHandlerTests(t, tests)
}

func TestHandlers_PostPrompt(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PostPromptRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
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
		tx      interfaces.IDatabaseTx
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
