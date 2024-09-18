package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestAssist(t *testing.T) {
	tests := []HandlersTestCase{
		{
			name: "GetSuggestion works when group ai is true",
			setupMocks: func(hts *HandlersTestSetup) {
				mockSession := &clients.UserSession{GroupAi: true}
				hts.MockRedis.EXPECT().ReqSession(gomock.Any()).Return(mockSession).Times(1)

				hts.MockAi.EXPECT().GetPromptResponse(gomock.Any(), gomock.Any(), gomock.Any()).Return("one|two|three|four|five")
			},
			handlerFunc: func(h *Handlers, w http.ResponseWriter, r *http.Request) (interface{}, error) {
				data := &types.GetSuggestionRequest{Id: "3", Prompt: "test-prompt"}
				return h.GetSuggestion(w, r, data)
			},
			expectedRes: &types.GetSuggestionResponse{PromptResult: []string{"one", "two", "three", "four", "five"}},
		},
		{
			name: "GetSuggestion returns nothing when group ai is false",
			setupMocks: func(hts *HandlersTestSetup) {
				mockSession := &clients.UserSession{GroupAi: false}
				hts.MockRedis.EXPECT().ReqSession(gomock.Any()).Return(mockSession)
			},
			handlerFunc: func(h *Handlers, w http.ResponseWriter, r *http.Request) (interface{}, error) {
				data := &types.GetSuggestionRequest{Id: "3", Prompt: "test-prompt"}
				return h.GetSuggestion(w, r, data)
			},
			expectedRes: &types.GetSuggestionResponse{PromptResult: []string{}},
		},
	}

	RunHandlerTests(t, tests)
}
