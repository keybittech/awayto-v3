package handlers

import (
	"bytes"
	"fmt"
	"html/template"

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

func (h *Handlers) GetGroupSeatPayments(info ReqInfo, data *types.GetGroupSeatPaymentsRequest) (*types.GetGroupSeatPaymentsResponse, error) {
	seatsReq := util.BatchQuery[types.IGroupSeatPayment](info.Batch, `
		SELECT status, code, seats, amount, "createdOn", "paidOn"
		FROM dbview_schema.enabled_seat_payments
		WHERE "groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupSeatPaymentsResponse{SeatPayments: *seatsReq}, nil
}

const poTemplate = `
<!DOCTYPE html>
<html>
<head>
  <title>Purchase Order {{.Code}}</title>
	<link rel="stylesheet" href="/app/print.css">
</head>
<body>
  <div class="header">
    <div class="title">
      <h1>Purchase Order</h1>
      <span class="status status-po">Status: {{.Status}}</span>
    </div>
    <div class="meta">
      <p><strong>PO #:</strong> {{.Code}}</p>
      <p><strong>Date:</strong> {{.CreatedOn}}</p>
    </div>
  </div>
  
  <div class="payable-section payable-section-po">
    <div>
      <h3>Pay to the order of:</h3>
      <p><strong>{{.PaymentTo}}</strong><br>
      {{.PaymentAddr1}}<br>
      {{.PaymentAddr2}}</p>
    </div>
  </div>

  <table>
    <thead>
      <tr>
        <th>Description</th>
        <th class="text-right">Quantity</th>
        <th class="text-right">Amount</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Group Seat License Expansion</td>
        <td class="text-right">{{.Seats}}</td>
        <td class="text-right">${{.AmountFormatted}}</td>
      </tr>
      <tr class="total-row">
        <td>Total</td>
        <td></td>
        <td class="text-right">${{.AmountFormatted}}</td>
      </tr>
    </tbody>
  </table>

  <div class="footer">
    <p>If you have any questions about this purchase order, please submit a feedback request at the top right of the website.</p>
  </div>

  <center>
    <button id="print-btn" class="btn-print">Print / Save as PDF</button>
  </center>
</body>
</html>
`

func (h *Handlers) GetGroupSeatPurchaseOrder(info ReqInfo, data *types.GetGroupSeatPurchaseOrderRequest) (*types.GetGroupSeatPurchaseOrderResponse, error) {
	payment := util.BatchQueryRow[types.IGroupSeatPayment](info.Batch, `
		SELECT status, code, seats, amount, "createdOn"
		FROM dbview_schema.enabled_seat_payments
		WHERE "groupId" = $1 AND code = $2
	`, info.Session.GetGroupId(), data.GetCode())

	info.Batch.Send(info.Ctx)

	if payment == nil || *payment == nil {
		return nil, util.ErrCheck(util.UserError("Payment not found"))
	}

	p := *payment

	tmplData := struct {
		Code            string
		Status          string
		Seats           int32
		AmountFormatted string
		CreatedOn       string
		PaymentTo       string
		PaymentAddr1    string
		PaymentAddr2    string
	}{
		Code:            p.GetCode(),
		Status:          p.GetStatus(),
		Seats:           p.GetSeats(),
		AmountFormatted: fmt.Sprintf("%.2f", float64(p.Amount)),
		CreatedOn:       p.GetCreatedOn(),
		PaymentTo:       util.E_PAYMENT_TO,
		PaymentAddr1:    util.E_PAYMENT_ADDR1,
		PaymentAddr2:    util.E_PAYMENT_ADDR2,
	}

	t, err := template.New("po").Parse(poTemplate)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, tmplData); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupSeatPurchaseOrderResponse{Html: buf.String()}, nil
}

const receiptTemplate = `
<!DOCTYPE html>
<html>
<head>
  <title>Receipt {{.Code}}</title>
	<link rel="stylesheet" href="/app/print.css">
</head>
<body>
  <div class="header">
    <div class="title">
      <h1>Payment Receipt</h1>
      <span class="status status-receipt">Status: {{.Status}}</span>
    </div>
    <div class="meta">
      <p><strong>Receipt For PO #:</strong> {{.Code}}</p>
      <p><strong>Original PO Date:</strong> {{.CreatedOn}}</p>
    </div>
  </div>
  
  <div class="payable-section payable-section-recepit">
    <div>
      <h3>Payment Received By:</h3>
      <p><strong>{{.PaymentTo}}</strong><br>
      {{.PaymentAddr1}}<br>
      {{.PaymentAddr2}}</p>
    </div>
    <div style="text-align: right;">
      <h3>Payment Details:</h3>
      <p><strong>Date Paid:</strong> {{.PaidOn}}</p>
      <p><strong>Check / Ref #:</strong> {{.CheckNo}}</p>
      <p><strong>Payment Method:</strong> Check</p>
    </div>
  </div>

  <table>
    <thead>
      <tr>
        <th>Description</th>
        <th class="text-right">Quantity</th>
        <th class="text-right">Amount</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Group Seat License Expansion</td>
        <td class="text-right">{{.Seats}}</td>
        <td class="text-right">${{.AmountFormatted}}</td>
      </tr>
      <tr class="total-row">
        <td>Total Paid</td>
        <td></td>
        <td class="text-right">${{.AmountFormatted}}</td>
      </tr>
      <tr class="balance-row">
        <td>Balance Due</td>
        <td></td>
        <td class="text-right">$0.00</td>
      </tr>
    </tbody>
  </table>

  <div class="footer">
    <p>Thank you for your business. This payment has been successfully processed.</p>
  </div>

  <center>
    <button id="print-btn" class="btn-print">Print / Save as PDF</button>
  </center>
</body>
</html>
`

func (h *Handlers) GetGroupSeatReceipt(info ReqInfo, data *types.GetGroupSeatReceiptRequest) (*types.GetGroupSeatReceiptResponse, error) {
	payment := util.BatchQueryRow[types.IGroupSeatPayment](info.Batch, `
		SELECT status, code, seats, amount, "createdOn", "paidOn", "checkNo"
		FROM dbview_schema.enabled_seat_payments
		WHERE "groupId" = $1 AND code = $2
	`, info.Session.GetGroupId(), data.GetCode())

	info.Batch.Send(info.Ctx)

	if payment == nil || *payment == nil {
		return nil, util.ErrCheck(util.UserError("Payment not found"))
	}

	p := *payment

	tmplData := struct {
		Code            string
		Status          string
		Seats           int32
		AmountFormatted string
		CreatedOn       string
		PaymentTo       string
		PaymentAddr1    string
		PaymentAddr2    string
		PaidOn          string
		CheckNo         string
	}{
		Code:            p.GetCode(),
		Status:          p.GetStatus(),
		Seats:           p.GetSeats(),
		AmountFormatted: fmt.Sprintf("%.2f", float64(p.Amount)),
		CreatedOn:       p.GetCreatedOn(),
		PaymentTo:       util.E_PAYMENT_TO,
		PaymentAddr1:    util.E_PAYMENT_ADDR1,
		PaymentAddr2:    util.E_PAYMENT_ADDR2,
		PaidOn:          p.GetPaidOn(),
		CheckNo:         p.GetCheckNo(),
	}

	t, err := template.New("receipt").Parse(receiptTemplate)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, tmplData); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupSeatReceiptResponse{Html: buf.String()}, nil
}
