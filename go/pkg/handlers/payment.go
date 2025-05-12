package handlers

import (
	json "encoding/json"
	"errors"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostPayment(info ReqInfo, data *types.PostPaymentRequest) (*types.PostPaymentResponse, error) {
	var paymentId string

	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.payments (contact_id, details, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id, contact_id as "contactId", details
	`, data.GetPayment().GetContactId(), data.GetPayment().GetContactId(), info.Session.UserSub).Scan(&paymentId)
	if err != nil || paymentId == "" {
		return nil, util.ErrCheck(err)
	}

	return &types.PostPaymentResponse{Id: paymentId}, nil
}

func (h *Handlers) PatchPayment(info ReqInfo, data *types.PatchPaymentRequest) (*types.PatchPaymentResponse, error) {
	paymentDetails, err := json.Marshal(data.GetPayment().GetDetails())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.payments
		SET contact_id = $2, details = $3, updated_sub = $4, updated_on = $5
		WHERE id = $1
		RETURNING id, contact_id as "contactId", details
	`, data.GetPayment().GetId(), data.GetPayment().GetContactId(), paymentDetails, info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchPaymentResponse{Success: true}, nil
}

func (h *Handlers) GetPayments(info ReqInfo, data *types.GetPaymentsRequest) (*types.GetPaymentsResponse, error) {
	var payments []*types.IPayment
	err := h.Database.QueryRows(info.Ctx, info.Tx, &payments, `SELECT * FROM dbview_schema.enabled_payments`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetPaymentsResponse{Payments: payments}, nil
}

func (h *Handlers) GetPaymentById(info ReqInfo, data *types.GetPaymentByIdRequest) (*types.GetPaymentByIdResponse, error) {
	var payments []*types.IPayment

	err := h.Database.QueryRows(info.Ctx, info.Tx, &payments, `
		SELECT * FROM dbview_schema.enabled_payments
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(payments) == 0 {
		return nil, util.ErrCheck(errors.New("payment not found"))
	}

	return &types.GetPaymentByIdResponse{Payment: payments[0]}, nil
}

func (h *Handlers) DeletePayment(info ReqInfo, data *types.DeletePaymentRequest) (*types.DeletePaymentResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.payments
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeletePaymentResponse{Success: true}, nil
}

func (h *Handlers) DisablePayment(info ReqInfo, data *types.DisablePaymentRequest) (*types.DisablePaymentResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.payments
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisablePaymentResponse{Success: true}, nil
}
