package handlers

import (
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) GetSuggestion(info ReqInfo, data *types.GetSuggestionRequest) (*types.GetSuggestionResponse, error) {
	if !info.Session.GetGroupAi() {
		return &types.GetSuggestionResponse{PromptResult: []string{}}, nil

	}
	promptParts := strings.Split(data.GetPrompt(), "!$")
	promptType, err := util.Atoi32(data.Id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	tryRequest := func() ([]string, error) {

		for range 3 {

			resp, err := h.LLM.GetCandidateResponse(info.Ctx, promptParts, types.IPrompts(promptType))
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			if resp == nil {
				continue
			}

			if len(resp.Candidates) < 1 {
				continue
			}

			candidate := resp.Candidates[0]
			if len(candidate.Content.Parts) < 1 {
				continue
			}

			options := strings.Split(strings.Trim(candidate.Content.Parts[0].Text, `"`), "|")
			if len(options) != 5 {
				continue
			}

			return options, nil
		}

		return []string{}, nil
	}

	suggestionResults, err := tryRequest()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetSuggestionResponse{PromptResult: suggestionResults}, nil
}
