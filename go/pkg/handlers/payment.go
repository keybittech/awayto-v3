package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func (h *Handlers) PostPayment(w http.ResponseWriter, req *http.Request, data *types.PostPaymentRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostPaymentResponse, error) {
	var paymentId string

	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.payments (contact_id, details, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id, contact_id as "contactId", details
	`, data.GetPayment().GetContactId(), data.GetPayment().GetContactId(), session.UserSub).Scan(&paymentId)
	if err != nil || paymentId == "" {
		return nil, util.ErrCheck(err)
	}

	return &types.PostPaymentResponse{Id: paymentId}, nil
}

func (h *Handlers) PatchPayment(w http.ResponseWriter, req *http.Request, data *types.PatchPaymentRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchPaymentResponse, error) {
	paymentDetails, err := json.Marshal(data.GetPayment().GetDetails())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		UPDATE dbtable_schema.payments
		SET contact_id = $2, details = $3, updated_sub = $4, updated_on = $5
		WHERE id = $1
		RETURNING id, contact_id as "contactId", details
	`, data.GetPayment().GetId(), data.GetPayment().GetContactId(), paymentDetails, session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchPaymentResponse{Success: true}, nil
}

func (h *Handlers) GetPayments(w http.ResponseWriter, req *http.Request, data *types.GetPaymentsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetPaymentsResponse, error) {
	var payments []*types.IPayment
	err := tx.QueryRows(&payments, `SELECT * FROM dbview_schema.enabled_payments`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetPaymentsResponse{Payments: payments}, nil
}

func (h *Handlers) GetPaymentById(w http.ResponseWriter, req *http.Request, data *types.GetPaymentByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetPaymentByIdResponse, error) {
	var payments []*types.IPayment

	err := tx.QueryRows(&payments, `
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

func (h *Handlers) DeletePayment(w http.ResponseWriter, req *http.Request, data *types.DeletePaymentRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeletePaymentResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.payments
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeletePaymentResponse{Success: true}, nil
}

func (h *Handlers) DisablePayment(w http.ResponseWriter, req *http.Request, data *types.DisablePaymentRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisablePaymentResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.payments
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisablePaymentResponse{Success: true}, nil
}
