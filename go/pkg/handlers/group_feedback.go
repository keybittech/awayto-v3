package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupFeedback(info ReqInfo, data *types.PostGroupFeedbackRequest) (*types.PostGroupFeedbackResponse, error) {
	util.BatchExec(info.Batch, `
		INSERT INTO dbtable_schema.group_feedback (message, group_id, created_sub)
		VALUES ($1, $2::uuid, $3::uuid)
	`, data.Feedback.FeedbackMessage, info.Session.GroupId, info.Session.UserSub)

	info.Batch.Send(info.Ctx)

	return &types.PostGroupFeedbackResponse{Success: true}, nil
}

func (h *Handlers) GetGroupFeedback(info ReqInfo, data *types.GetGroupFeedbackRequest) (*types.GetGroupFeedbackResponse, error) {
	feedback := util.BatchQuery[types.IFeedback](info.Batch, `
		SELECT f.id, f.message as "feedbackMessage", f.created_on as "createdOn"
		FROM dbtable_schema.group_feedback f
		WHERE f.group_id = $1
		ORDER BY f.created_on DESC
	`, info.Session.GroupId)

	info.Batch.Send(info.Ctx)

	return &types.GetGroupFeedbackResponse{Feedback: *feedback}, nil
}
