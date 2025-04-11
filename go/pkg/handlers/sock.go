package handlers

import (
	"database/sql"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) GetSocketTicket(w http.ResponseWriter, req *http.Request, data *types.GetSocketTicketRequest, session *types.UserSession, tx *sql.Tx) (*types.GetSocketTicketResponse, error) {
	ticket, err := h.Socket.GetSocketTicket(session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetSocketTicketResponse{Ticket: ticket}, nil
}
