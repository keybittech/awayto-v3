package handlers

import (
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostSiteFeedback(w http.ResponseWriter, req *http.Request, data *types.PostSiteFeedbackRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PostSiteFeedbackResponse, error) {
	_, err := tx.Exec(`
		INSERT INTO dbtable_schema.feedback (message, created_sub, created_on)
		VALUES ($1, $2::uuid, $3)
	`, data.GetFeedback().GetFeedbackMessage(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostSiteFeedbackResponse{Success: true}, nil
}
