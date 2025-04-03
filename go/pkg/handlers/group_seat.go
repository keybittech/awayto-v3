package handlers

import (
	"fmt"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func (h *Handlers) PostGroupSeat(w http.ResponseWriter, r *http.Request, data *types.PostGroupSeatRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PostGroupSeatResponse, error) {

	fmt.Printf("Got seat %d", data.GetSeats())

	return nil, nil
}
