package handlers

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/mocks"
	"github.com/keybittech/awayto-v3/go/pkg/types"

	"github.com/golang/mock/gomock"
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
			handlerFunc: func(h *Handlers, w http.ResponseWriter, r *http.Request, session *types.UserSession, tx *mocks.MockIDatabaseTx) (interface{}, error) {
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
			handlerFunc: func(h *Handlers, w http.ResponseWriter, r *http.Request, session *types.UserSession, tx *mocks.MockIDatabaseTx) (interface{}, error) {
				data := &types.GetSuggestionRequest{Id: "3", Prompt: "test-prompt"}
				return h.GetSuggestion(w, r, data, session, tx)
			},
			expectedRes: &types.GetSuggestionResponse{PromptResult: []string{}},
		},
	}

	RunHandlerTests(t, tests)
}
