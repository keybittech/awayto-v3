package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
)

func (h *Handlers) GetSocketTicket(w http.ResponseWriter, req *http.Request, data *types.GetSocketTicketRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetSocketTicketResponse, error) {
	ticket, err := h.Socket.GetSocketTicket(session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetSocketTicketResponse{Ticket: ticket}, nil
}
