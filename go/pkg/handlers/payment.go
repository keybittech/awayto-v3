package handlers

// func (h *Handlers) PostPayment(info ReqInfo, data *types.PostPaymentRequest) (*types.PostPaymentResponse, error) {
// 	paymentDetails, err := protojson.Marshal(data.Payment.Details)
// 	if err != nil {
// 		return nil, util.ErrCheck(err)
// 	}
//
// 	paymentInsert := util.BatchQueryRow[types.ILookup](info.Batch, `
// 		INSERT INTO dbtable_schema.seat_payments (details, created_sub)
// 		VALUES ($1, $2, $3::uuid)
// 		RETURNING id
// 	`, paymentDetails, info.Session.GetUserSub())
//
// 	info.Batch.Send(info.Ctx)
//
// 	return &types.PostPaymentResponse{Id: (*paymentInsert).Id}, nil
// }
//
// func (h *Handlers) PatchPayment(info ReqInfo, data *types.PatchPaymentRequest) (*types.PatchPaymentResponse, error) {
// 	paymentDetails, err := protojson.Marshal(data.Payment.Details)
// 	if err != nil {
// 		return nil, util.ErrCheck(err)
// 	}
//
// 	util.BatchExec(info.Batch, `
// 		UPDATE dbtable_schema.seat_payments
// 		SET details = $2, updated_sub = $3, updated_on = $4
// 		WHERE id = $1
// 	`, data.Payment.Id, paymentDetails, info.Session.GetUserSub(), time.Now())
//
// 	info.Batch.Send(info.Ctx)
//
// 	return &types.PatchPaymentResponse{Success: true}, nil
// }
//
// func (h *Handlers) GetPayments(info ReqInfo, data *types.GetPaymentsRequest) (*types.GetPaymentsResponse, error) {
// 	payments := util.BatchQuery[types.IPayment](info.Batch, `
// 		SELECT id, "createdOn"
// 		FROM dbview_schema.enabled_seat_payments
// 		WHERE "createdSub" = $1
// 	`, info.Session.GetUserSub())
//
// 	info.Batch.Send(info.Ctx)
//
// 	return &types.GetPaymentsResponse{Payments: *payments}, nil
// }
//
// func (h *Handlers) GetPaymentById(info ReqInfo, data *types.GetPaymentByIdRequest) (*types.GetPaymentByIdResponse, error) {
// 	payment := util.BatchQueryRow[types.IPayment](info.Batch, `
// 		SELECT id, details, "createdOn"
// 		FROM dbview_schema.enabled_seat_payments
// 		WHERE id = $1
// 	`, data.Id)
//
// 	info.Batch.Send(info.Ctx)
//
// 	return &types.GetPaymentByIdResponse{Payment: *payment}, nil
// }
//
// func (h *Handlers) DeletePayment(info ReqInfo, data *types.DeletePaymentRequest) (*types.DeletePaymentResponse, error) {
// 	util.BatchExec(info.Batch, `
// 		DELETE FROM dbtable_schema.seat_payments
// 		WHERE id = $1
// 	`, data.Id)
//
// 	info.Batch.Send(info.Ctx)
//
// 	return &types.DeletePaymentResponse{Success: true}, nil
// }
//
// func (h *Handlers) DisablePayment(info ReqInfo, data *types.DisablePaymentRequest) (*types.DisablePaymentResponse, error) {
// 	util.BatchExec(info.Batch, `
// 		UPDATE dbtable_schema.seat_payments
// 		SET enabled = false, updated_on = $2, updated_sub = $3
// 		WHERE id = $1
// 	`, data.Id, time.Now(), info.Session.GetUserSub())
//
// 	info.Batch.Send(info.Ctx)
//
// 	return &types.DisablePaymentResponse{Success: true}, nil
// }
