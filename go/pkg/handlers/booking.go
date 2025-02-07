package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
	"time"
)

func (h *Handlers) PostBooking(w http.ResponseWriter, req *http.Request, data *types.PostBookingRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostBookingResponse, error) {
	newBookings := make([]*types.IBooking, 0)

	for _, booking := range data.Bookings {
		var newBooking types.IBooking
		err := tx.QueryRow(`
			INSERT INTO dbtable_schema.bookings (quote_id, slot_date, schedule_bracket_slot_id, created_sub)
			VALUES ($1::uuid, $2::date, $3::uuid, $4::uuid)
			RETURNING id
		`, booking.GetQuote().GetId(), booking.GetQuote().GetSlotDate(), booking.GetQuote().GetScheduleBracketSlotId(), session.UserSub).Scan(&newBooking.Id)

		if err != nil {
			return nil, util.ErrCheck(err)
		}
		newBookings = append(newBookings, &newBooking)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	return &types.PostBookingResponse{Bookings: newBookings}, nil
}

func (h *Handlers) PatchBooking(w http.ResponseWriter, req *http.Request, data *types.PatchBookingRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchBookingResponse, error) {
	var updatedBookings []*types.IBooking
	err := h.Database.QueryRows(&updatedBookings, `
		UPDATE dbtable_schema.bookings
		SET service_tier_id = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
		RETURNING id, service_tier_id, payment_id, agreement, description
	`, data.Booking.GetId(), data.Booking.GetQuote().GetServiceTierId(), session.UserSub, time.Now().Local().UTC())

	if err != nil || len(updatedBookings) == 0 {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchBookingResponse{Success: true}, nil
}

func (h *Handlers) GetBookings(w http.ResponseWriter, req *http.Request, data *types.GetBookingsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetBookingsResponse, error) {
	bookings := []*types.IBooking{}
	err := h.Database.QueryRows(&bookings, `
		SELECT eb.*
		FROM dbview_schema.enabled_bookings eb
		JOIN dbtable_schema.bookings b ON b.id = eb.id
		LEFT JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = eb.schedule_bracket_slot_id
		WHERE b.created_sub = $1 OR sbs.created_sub = $1
	`, session.UserSub)
	return &types.GetBookingsResponse{Bookings: bookings}, err
}

func (h *Handlers) GetBookingById(w http.ResponseWriter, req *http.Request, data *types.GetBookingByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetBookingByIdResponse, error) {
	var bookings []*types.IBooking
	err := h.Database.QueryRows(&bookings, `
		SELECT * FROM dbview_schema.enabled_bookings
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetBookingByIdResponse{Booking: bookings[0]}, err
}

func (h *Handlers) GetBookingFiles(w http.ResponseWriter, req *http.Request, data *types.GetBookingFilesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetBookingFilesResponse, error) {
	files := []*types.IFile{}
	err := h.Database.QueryRows(&files, `
		SELECT f.name, f.uuid, f."mimeType"
		FROM dbview_schema.enabled_files f
		JOIN dbtable_schema.quote_files qf ON qf.file_id = f.id
		JOIN dbtable_schema.bookings b ON b.quote_id = qf.quote_id
		WHERE b.id = $1
	`, data.GetId())
	return &types.GetBookingFilesResponse{Files: files}, err
}

func (h *Handlers) PatchBookingRating(w http.ResponseWriter, req *http.Request, data *types.PatchBookingRatingRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchBookingRatingResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.bookings
		SET rating = $2
		WHERE id = $1
	`, data.GetId(), data.GetRating())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchBookingRatingResponse{Success: true}, nil
}

func (h *Handlers) DeleteBooking(w http.ResponseWriter, req *http.Request, data *types.DeleteBookingRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteBookingResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.bookings
		WHERE id = $1
	`, data.GetId())
	return &types.DeleteBookingResponse{Id: data.Id}, err
}

func (h *Handlers) DisableBooking(w http.ResponseWriter, req *http.Request, data *types.DisableBookingRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableBookingResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.bookings
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	return &types.DisableBookingResponse{Id: data.Id}, err
}
