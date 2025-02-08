package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"fmt"
	"net/http"
)

func (h *Handlers) PostGroupSeat(w http.ResponseWriter, r *http.Request, data *types.PostGroupSeatRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostGroupSeatResponse, error) {

	fmt.Printf("Got seat %d", data.GetSeats())

	return nil, nil
}
