package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupSeat(info ReqInfo, data *types.PostGroupSeatRequest) (*types.PostGroupSeatResponse, error) {
	seats := data.GetSeats()
	amount := 10 * seats

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
	paymentReq := util.BatchQueryRow[types.IPayment](batch, `
		INSERT INTO dbtable_schema.payments (group_id, created_sub, code, amount)
		VALUES ($1, $2::uuid, 'unset_code', $3)
		RETURNING id 
	`, info.Session.GetGroupId(), info.Session.GetUserSub(), amount)

	batch.Send(info.Ctx)

	paymentId := (*paymentReq).GetId()

	var paymentCode string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT code
		FROM dbtable_schema.payments
		WHERE id = $1
	`, paymentId).Scan(&paymentCode)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	batch.Reset()
	util.BatchExec(batch, `
		INSERT INTO dbtable_schema.group_seats (group_id, payment_id, seats, created_sub)
		VALUES ($1, $2::uuid, $3, $4::uuid)
	`, info.Session.GetGroupId(), paymentId, seats, info.Session.GetUserSub())
	batch.Send(info.Ctx)

	return nil, nil
}
