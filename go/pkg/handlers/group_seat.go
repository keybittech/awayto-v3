package handlers

import (
	"av3api/pkg/types"
	"fmt"
	"net/http"
)

func (h *Handlers) PostGroupSeat(w http.ResponseWriter, r *http.Request, data *types.PostGroupSeatRequest) (*types.PostGroupSeatResponse, error) {

	fmt.Printf("Got seat %d", data.GetSeats())

	return nil, nil
}
