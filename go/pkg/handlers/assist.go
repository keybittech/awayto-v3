package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
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
			util.ErrCheck(err)
			return nil, util.ErrCheck(err)
		}

		resp := h.Ai.GetPromptResponse(req.Context(), promptParts, types.IPrompts(promptType))

		if strings.Contains(resp, "Result:") {
			resp = strings.Split(resp, "Result:")[1]
		}

		suggestions := []string{}
		for _, str := range strings.Split(resp, "|") {
			suggestions = append(suggestions, strings.TrimSpace(str))
		}

		return &types.GetSuggestionResponse{PromptResult: suggestions}, nil
	}

	return &types.GetSuggestionResponse{PromptResult: []string{}}, nil
}
