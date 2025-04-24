package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostBooking(w http.ResponseWriter, req *http.Request, data *types.PostBookingRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PostBookingResponse, error) {
	newBookings := make([]*types.IBooking, 0)

	var scheduleBracketSlotId string
	for i, booking := range data.Bookings {
		if i == 0 {
			scheduleBracketSlotId = booking.Quote.ScheduleBracketSlotId
		} else {
			if booking.Quote.ScheduleBracketSlotId != scheduleBracketSlotId {
				return nil, util.ErrCheck(util.UserError("Only appointments of the same date and time may be batch approved."))
			}
		}
	}

	var isOwner bool
	err := tx.QueryRow(req.Context(), `
		SELECT EXISTS(
			SELECT 1 FROM dbtable_schema.schedule_bracket_slots
			WHERE id = $1 AND created_sub = $2
		)
	`, scheduleBracketSlotId, session.UserSub).Scan(&isOwner)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	if !isOwner {
		return nil, util.ErrCheck(errors.New("sub " + session.UserSub + " attempted to approve non-owned sbsid " + scheduleBracketSlotId))
	}

	for _, booking := range data.Bookings {
		var newBooking types.IBooking
		err := tx.QueryRow(req.Context(), `
			INSERT INTO dbtable_schema.bookings (quote_id, slot_date, schedule_bracket_slot_id, created_sub)
			VALUES ($1::uuid, $2::date, $3::uuid, $4::uuid)
			RETURNING id
		`, booking.Quote.Id, booking.Quote.SlotDate, scheduleBracketSlotId, session.UserSub).Scan(&newBooking.Id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		var quoteUserSub string
		err = tx.QueryRow(req.Context(), `
			SELECT created_sub
			FROM dbtable_schema.quotes
			WHERE id = $1
		`, booking.Quote.Id).Scan(&quoteUserSub)

		h.Redis.Client().Del(req.Context(), quoteUserSub+"profile/details")

		if err := h.Socket.RoleCall(quoteUserSub); err != nil {
			return nil, util.ErrCheck(err)
		}

		newBookings = append(newBookings, &newBooking)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	return &types.PostBookingResponse{Bookings: newBookings}, nil
}

func (h *Handlers) PatchBooking(w http.ResponseWriter, req *http.Request, data *types.PatchBookingRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PatchBookingResponse, error) {
	var updatedBookings []*types.IBooking
	err := h.Database.QueryRows(req.Context(), tx, &updatedBookings, `
		UPDATE dbtable_schema.bookings
		SET service_tier_id = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
	`, data.Booking.Id, data.Booking.Quote.ServiceTierId, session.UserSub, time.Now())

	if err != nil || len(updatedBookings) == 0 {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchBookingResponse{Success: true}, nil
}

func (h *Handlers) GetBookings(w http.ResponseWriter, req *http.Request, data *types.GetBookingsRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetBookingsResponse, error) {
	bookings := []*types.IBooking{}
	err := h.Database.QueryRows(req.Context(), tx, &bookings, `
		SELECT eb.*
		FROM dbview_schema.enabled_bookings eb
		JOIN dbtable_schema.bookings b ON b.id = eb.id
		LEFT JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = eb.schedule_bracket_slot_id
		WHERE b.created_sub = $1 OR sbs.created_sub = $1
	`, session.UserSub)
	return &types.GetBookingsResponse{Bookings: bookings}, err
}

func (h *Handlers) GetBookingById(w http.ResponseWriter, req *http.Request, data *types.GetBookingByIdRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetBookingByIdResponse, error) {
	var bookings []*types.IBooking
	err := h.Database.QueryRows(req.Context(), tx, &bookings, `
		SELECT * FROM dbview_schema.enabled_bookings
		WHERE id = $1
	`, data.Id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(bookings) == 0 {
		return nil, util.ErrCheck(util.UserError("No booking found."))
	}

	return &types.GetBookingByIdResponse{Booking: bookings[0]}, err
}

func (h *Handlers) GetBookingFiles(w http.ResponseWriter, req *http.Request, data *types.GetBookingFilesRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetBookingFilesResponse, error) {
	files := []*types.IFile{}
	err := h.Database.QueryRows(req.Context(), tx, &files, `
		SELECT f.name, f.uuid, f."mimeType"
		FROM dbview_schema.enabled_files f
		JOIN dbtable_schema.quote_files qf ON qf.file_id = f.id
		JOIN dbtable_schema.bookings b ON b.quote_id = qf.quote_id
		WHERE b.id = $1
	`, data.Id)
	return &types.GetBookingFilesResponse{Files: files}, err
}

func (h *Handlers) PatchBookingRating(w http.ResponseWriter, req *http.Request, data *types.PatchBookingRatingRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PatchBookingRatingResponse, error) {
	_, err := tx.Exec(req.Context(), `
		UPDATE dbtable_schema.bookings
		SET rating = $2
		WHERE id = $1
	`, data.Id, data.Rating)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"bookings/"+data.Id)

	return &types.PatchBookingRatingResponse{Success: true}, nil
}

func (h *Handlers) DeleteBooking(w http.ResponseWriter, req *http.Request, data *types.DeleteBookingRequest, session *types.UserSession, tx *clients.PoolTx) (*types.DeleteBookingResponse, error) {
	_, err := tx.Exec(req.Context(), `
		DELETE FROM dbtable_schema.bookings
		WHERE id = $1
	`, data.GetId())
	return &types.DeleteBookingResponse{Id: data.Id}, err
}

func (h *Handlers) DisableBooking(w http.ResponseWriter, req *http.Request, data *types.DisableBookingRequest, session *types.UserSession, tx *clients.PoolTx) (*types.DisableBookingResponse, error) {
	_, err := tx.Exec(req.Context(), `
		UPDATE dbtable_schema.bookings
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	return &types.DisableBookingResponse{Id: data.Id}, err
}
