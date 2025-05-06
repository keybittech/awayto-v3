package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostSiteFeedback(info ReqInfo, data *types.PostSiteFeedbackRequest) (*types.PostSiteFeedbackResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.feedback (message, created_sub, created_on)
		VALUES ($1, $2::uuid, $3)
	`, data.GetFeedback().GetFeedbackMessage(), info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostSiteFeedbackResponse{Success: true}, nil
}
