package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func (h *Handlers) PostPrompt(w http.ResponseWriter, req *http.Request, data *types.PostPromptRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostPromptResponse, error) {

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

func (h *Handlers) GetSuggestion(w http.ResponseWriter, req *http.Request, data *types.GetSuggestionRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetSuggestionResponse, error) {

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
					continue
				}

				if strings.Contains(resp, "Result:") {
					resp = strings.Split(resp, "Result:")[1]
				}

				r, err := regexp.Compile(`[-:\.]`)
				if err != nil {
					continue
				}
				if r.MatchString(resp) {
					continue
				}

				suggestions := strings.Split(resp, "|")
				for _, str := range suggestions {
					str = strings.TrimSpace(str)
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
