package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"net/http"
)

func (h *Handlers) GetSocketTicket(w http.ResponseWriter, req *http.Request, data *types.GetSocketTicketRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetSocketTicketResponse, error) {
	ticket, err := h.Socket.GetSocketTicket(session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetSocketTicketResponse{Ticket: ticket}, nil
}
