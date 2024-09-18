package handlers

import (
	"av3api/pkg/types"
	"net/http"
)

func (h *Handlers) GetSocketTicket(w http.ResponseWriter, req *http.Request, data *types.GetSocketTicketRequest) (*types.GetSocketTicketResponse, error) {
	session := h.Redis.ReqSession(req)

	ticket, err := h.Socket.GetSocketTicket(session.UserSub)
	if err != nil {
		return nil, err
	}

	return &types.GetSocketTicketResponse{Ticket: ticket}, nil
}
