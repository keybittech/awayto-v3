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

const commonStyles = `
{{define "common_css"}}
  body { 
    font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; 
    padding: 40px; 
    color: #333; 
    max-width: 800px; 
    margin: 0 auto; 
  }
  .header { display: flex; justify-content: space-between; border-bottom: 2px solid #eee; padding-bottom: 20px; margin-bottom: 30px; }
  .title h1 { margin: 0; color: #2c3e50; font-size: 24px; text-transform: uppercase; letter-spacing: 1px; }
  .meta { text-align: right; }
  .meta p { margin: 2px 0; font-size: 14px; }
  
  .payable-section { 
    padding: 20px; 
    border-radius: 4px; 
    margin-bottom: 30px; 
    border-left-width: 4px; 
    border-left-style: solid; 
    display: flex; 
    justify-content: space-between;
  }
  .payable-section div { flex: 1; }
  .payable-section h3 { margin-top: 0; margin-bottom: 10px; font-size: 14px; text-transform: uppercase; }
  .payable-section p { margin: 0; line-height: 1.5; font-size: 15px; }

  table { width: 100%; border-collapse: collapse; }
  th { text-align: left; padding: 12px; background: #2c3e50; color: white; font-size: 12px; text-transform: uppercase; }
  td { padding: 12px; border-bottom: 1px solid #eee; }
  .total-row td { border-top: 2px solid #333; font-weight: bold; font-size: 18px; }
  .text-right { text-align: right; }

  .status { padding: 5px 10px; font-size: 12px; font-weight: bold; display: inline-block; margin-top: 5px; }
  
  .footer { padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #7f8c8d; text-align: center; }
  
  .btn-print { margin-top:20px; padding: 10px 25px; cursor: pointer; background: #2c3e50; color: white; border: none; border-radius: 4px; font-size: 14px; }

  @media print { 
    @page { margin: 0; size: auto; }
    body { margin: 1cm; padding: 20px; max-width: 100%; width: auto; } 
    #print-btn { display: none; } 
    .payable-section { background: none !important; border: 1px solid #ccc !important; }
    * { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  }
{{end}}
`

const poTemplate = `
<!DOCTYPE html>
<html>
<head>
  <title>Purchase Order {{.Code}}</title>
  <style>
    {{template "common_css" .}}
    .payable-section { background: #f9f9f9; border-left-color: #2c3e50; }
    .status { background: #eee; }
  </style>
</head>
<body>
  <div class="header">
    <div class="title">
      <h1>Purchase Order</h1>
      <span class="status">Status: {{.Status}}</span>
    </div>
    <div class="meta">
      <p><strong>PO #:</strong> {{.Code}}</p>
      <p><strong>Date:</strong> {{.CreatedOn}}</p>
    </div>
  </div>
  
  <div class="payable-section">
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

	t, err := template.New("po").Parse(commonStyles + poTemplate)
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
  <style>
		{{template "common_css" .}}
    .payable-section { background: #f0fdf4; border-left-color: #166534; }
    .status { background: #166534; color: white; }
    .balance-row td { font-size: 16px; border-bottom: none; }
  </style>
</head>
<body>
  <div class="header">
    <div class="title">
      <h1>Payment Receipt</h1>
      <span class="status">Status: {{.Status}}</span>
    </div>
    <div class="meta">
      <p><strong>Receipt For PO #:</strong> {{.Code}}</p>
      <p><strong>Original PO Date:</strong> {{.CreatedOn}}</p>
    </div>
  </div>
  
  <div class="payable-section">
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

	t, err := template.New("receipt").Parse(commonStyles + receiptTemplate)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, tmplData); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupSeatReceiptResponse{Html: buf.String()}, nil
}
