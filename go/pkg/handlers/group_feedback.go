package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"net/http"
	"time"
)

func (h *Handlers) PostGroupFeedback(w http.ResponseWriter, req *http.Request, data *types.PostGroupFeedbackRequest, session *types.UserSession, tx clients.IDatabaseTx) (*types.PostGroupFeedbackResponse, error) {
	_, err := tx.Exec(`
		INSERT INTO dbtable_schema.group_feedback (message, group_id, created_sub, created_on)
		VALUES ($1, $2::uuid, $3::uuid, $4)
	`, data.GetFeedback().GetFeedbackMessage(), session.GroupId, session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/feedback")

	return &types.PostGroupFeedbackResponse{Success: true}, nil
}

func (h *Handlers) GetGroupFeedback(w http.ResponseWriter, req *http.Request, data *types.GetGroupFeedbackRequest, session *types.UserSession, tx clients.IDatabaseTx) (*types.GetGroupFeedbackResponse, error) {
	var feedback []*types.IFeedback

	err := tx.QueryRows(&feedback, `
		SELECT f.id, f.message as "feedbackMessage", f.created_on as "createdOn"
		FROM dbtable_schema.group_feedback f
		JOIN dbtable_schema.users u ON u.sub = f.created_sub
		WHERE f.group_id = $1
		ORDER BY f.created_on DESC
	`, session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupFeedbackResponse{Feedback: feedback}, nil
}
