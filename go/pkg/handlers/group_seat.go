package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupSeat(info ReqInfo, data *types.PostGroupSeatRequest) (*types.PostGroupSeatResponse, error) {
	seats := data.GetSeats()
	amount := 10 * seats

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
	util.BatchExec(batch, `
		INSERT INTO dbtable_schema.seat_payments (group_id, created_sub, code, seats, amount)
		VALUES ($1::uuid, $2::uuid, 'unset_code', $3, $4)
		RETURNING id 
	`, info.Session.GetGroupId(), info.Session.GetUserSub(), seats, amount)

	batch.Send(info.Ctx)

	return nil, nil
}

func (h *Handlers) GetGroupSeats(info ReqInfo, data *types.GetGroupSeatsRequest) (*types.GetGroupSeatsResponse, error) {
	seatsReq := util.BatchQueryRow[types.IGroupSeat](info.Batch, `
		SELECT balance
		FROM dbtable_schema.group_seats
		WHERE group_id = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupSeatsResponse{Balance: (*seatsReq).GetBalance()}, nil
}
