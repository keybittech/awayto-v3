package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostSiteFeedback(info ReqInfo, data *types.PostSiteFeedbackRequest) (*types.PostSiteFeedbackResponse, error) {
	util.BatchExec(info.Batch, `
		INSERT INTO dbtable_schema.feedback (message, created_sub)
		VALUES ($1, $2::uuid)
	`, data.Feedback.FeedbackMessage, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.PostSiteFeedbackResponse{Success: true}, nil
}
