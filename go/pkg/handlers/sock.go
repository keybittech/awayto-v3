package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) GetSocketTicket(info ReqInfo, data *types.GetSocketTicketRequest) (*types.GetSocketTicketResponse, error) {
	ticket, err := h.Socket.GetSocketTicket(info.Session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetSocketTicketResponse{Ticket: ticket}, nil
}
