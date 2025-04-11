package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostPrompt(w http.ResponseWriter, req *http.Request, data *types.PostPromptRequest, session *types.UserSession, tx *sql.Tx) (*types.PostPromptResponse, error) {

	// TODO
	// if rateLimited, err := h.Redis.Client().RateLimitResource(req.Context(), data.UserSub, "prompt", 25, 86400); err != nil || rateLimited {
	// 	if err != nil {
	// 		return nil, util.ErrCheck(err)
	// 	}
	// 	return &types.PostPromptResponse{PromptResult: []string{}}, nil
	// }
	//
	// promptParts := strings.Split(data.Prompt, "!$")
	// response, err := h.AI.UseAI(req.Context(), data.Id, promptParts...)

	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }

	// promptResults := strings.Split(response.Message, "|")
	trimmedResults := make([]string, 0)
	//
	// for _, result := range promptResults {
	// 	trimmedResult := strings.TrimSpace(result)
	// 	if trimmedResult != "" {
	// 		trimmedResults = append(trimmedResults, trimmedResult)
	// 	}
	// }

	return &types.PostPromptResponse{PromptResult: trimmedResults}, nil
}

func (h *Handlers) GetSuggestion(w http.ResponseWriter, req *http.Request, data *types.GetSuggestionRequest, session *types.UserSession, tx *sql.Tx) (*types.GetSuggestionResponse, error) {

	if session.GroupAi {
		promptParts := strings.Split(data.GetPrompt(), "!$")
		promptType, err := strconv.Atoi(data.GetId())
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		tryRequest := func() ([]string, error) {

			for i := 0; i < 3; i++ {

				resp, err := h.Ai.GetPromptResponse(req.Context(), promptParts, types.IPrompts(promptType))
				if err != nil {
					return nil, util.ErrCheck(err)
				}

				if resp == "" {
					continue
				}

				lowerCheck := strings.ToLower(resp)

				if strings.Contains(lowerCheck, "sorry") {
					continue
				}

				if strings.Contains(lowerCheck, "result:") {
					resp = strings.Split(resp, "Result:")[1]
				}

				suggestions := strings.Split(resp, "|")
				for _, str := range suggestions {
					str = strings.TrimSpace(str)
				}

				if len(suggestions) != 5 {
					continue
				}

				return suggestions, nil
			}

			return nil, nil
		}

		suggestionResults, err := tryRequest()
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		return &types.GetSuggestionResponse{PromptResult: suggestionResults}, nil
	}

	return &types.GetSuggestionResponse{PromptResult: []string{}}, nil
}
