package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupFeedback(info ReqInfo, data *types.PostGroupFeedbackRequest) (*types.PostGroupFeedbackResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		INSERT INTO dbtable_schema.group_feedback (message, group_id, created_sub, created_on)
		VALUES ($1, $2::uuid, $3::uuid, $4)
	`, data.GetFeedback().GetFeedbackMessage(), info.Session.GroupId, info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/feedback")

	return &types.PostGroupFeedbackResponse{Success: true}, nil
}

func (h *Handlers) GetGroupFeedback(info ReqInfo, data *types.GetGroupFeedbackRequest) (*types.GetGroupFeedbackResponse, error) {
	var feedback []*types.IFeedback

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &feedback, `
		SELECT f.id, f.message as "feedbackMessage", f.created_on as "createdOn"
		FROM dbtable_schema.group_feedback f
		JOIN dbtable_schema.users u ON u.sub = f.created_sub
		WHERE f.group_id = $1
		ORDER BY f.created_on DESC
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupFeedbackResponse{Feedback: feedback}, nil
}
