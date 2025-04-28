package handlers

import (
	"fmt"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func (h *Handlers) PostGroupSeat(info ReqInfo, data *types.PostGroupSeatRequest) (*types.PostGroupSeatResponse, error) {

	fmt.Printf("Got seat %d", data.GetSeats())

	return nil, nil
}
