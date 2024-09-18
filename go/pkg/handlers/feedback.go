package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
	"time"
)

func (h *Handlers) PostSiteFeedback(w http.ResponseWriter, req *http.Request, data *types.PostSiteFeedbackRequest) (*types.PostSiteFeedbackResponse, error) {
	session := h.Redis.ReqSession(req)
	_, err := h.Database.Client().Exec(`
		INSERT INTO dbtable_schema.feedback (message, created_sub, created_on)
		VALUES ($1, $2::uuid, $3)
	`, data.GetFeedback(), session.UserSub, time.Now().Local().UTC())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostSiteFeedbackResponse{Success: true}, nil
}
