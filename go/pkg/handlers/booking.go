package handlers

import (
	"errors"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostBooking(info ReqInfo, data *types.PostBookingRequest) (*types.PostBookingResponse, error) {
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
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT EXISTS(
			SELECT 1 FROM dbtable_schema.schedule_bracket_slots
			WHERE id = $1 AND created_sub = $2
		)
	`, scheduleBracketSlotId, info.Session.UserSub).Scan(&isOwner)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	if !isOwner {
		return nil, util.ErrCheck(errors.New("sub " + info.Session.UserSub + " attempted to approve non-owned sbsid " + scheduleBracketSlotId))
	}

	for _, booking := range data.Bookings {
		var quoteCreatedSub string
		err = info.Tx.QueryRow(info.Ctx, `
			SELECT created_sub
			FROM dbtable_schema.quotes
			WHERE id = $1
		`, booking.Quote.Id).Scan(&quoteCreatedSub)

		var newBooking types.IBooking
		err := info.Tx.QueryRow(info.Ctx, `
			INSERT INTO dbtable_schema.bookings (quote_id, slot_date, schedule_bracket_slot_id, created_sub, quote_created_sub)
			VALUES ($1::uuid, $2::date, $3::uuid, $4::uuid, $5::uuid)
			RETURNING id
		`, booking.Quote.Id, booking.Quote.SlotDate, scheduleBracketSlotId, info.Session.UserSub, quoteCreatedSub).Scan(&newBooking.Id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		h.Redis.Client().Del(info.Ctx, quoteCreatedSub+"profile/details")

		if err := h.Socket.RoleCall(quoteCreatedSub); err != nil {
			return nil, util.ErrCheck(err)
		}

		newBookings = append(newBookings, &newBooking)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")
	return &types.PostBookingResponse{Bookings: newBookings}, nil
}

func (h *Handlers) PatchBooking(info ReqInfo, data *types.PatchBookingRequest) (*types.PatchBookingResponse, error) {
	var updatedBookings []*types.IBooking
	err := h.Database.QueryRows(info.Ctx, info.Tx, &updatedBookings, `
		UPDATE dbtable_schema.bookings
		SET service_tier_id = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
	`, data.Booking.Id, data.Booking.Quote.ServiceTierId, info.Session.UserSub, time.Now())

	if err != nil || len(updatedBookings) == 0 {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchBookingResponse{Success: true}, nil
}

func (h *Handlers) GetBookings(info ReqInfo, data *types.GetBookingsRequest) (*types.GetBookingsResponse, error) {
	bookings := []*types.IBooking{}
	err := h.Database.QueryRows(info.Ctx, info.Tx, &bookings, `
		SELECT eb.*
		FROM dbview_schema.enabled_bookings eb
		JOIN dbtable_schema.bookings b ON b.id = eb.id
		LEFT JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = eb.schedule_bracket_slot_id
		WHERE b.created_sub = $1 OR sbs.created_sub = $1
	`, info.Session.UserSub)
	return &types.GetBookingsResponse{Bookings: bookings}, err
}

func (h *Handlers) GetBookingById(info ReqInfo, data *types.GetBookingByIdRequest) (*types.GetBookingByIdResponse, error) {
	var bookings []*types.IBooking
	err := h.Database.QueryRows(info.Ctx, info.Tx, &bookings, `
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

func (h *Handlers) GetBookingFiles(info ReqInfo, data *types.GetBookingFilesRequest) (*types.GetBookingFilesResponse, error) {
	files := []*types.IFile{}
	err := h.Database.QueryRows(info.Ctx, info.Tx, &files, `
		SELECT f.name, f.uuid, f."mimeType"
		FROM dbview_schema.enabled_files f
		JOIN dbtable_schema.quote_files qf ON qf.file_id = f.id
		JOIN dbtable_schema.bookings b ON b.quote_id = qf.quote_id
		WHERE b.id = $1
	`, data.Id)
	return &types.GetBookingFilesResponse{Files: files}, err
}

func (h *Handlers) PatchBookingRating(info ReqInfo, data *types.PatchBookingRatingRequest) (*types.PatchBookingRatingResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.bookings
		SET rating = $2
		WHERE id = $1
	`, data.Id, data.Rating)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"bookings/"+data.Id)

	return &types.PatchBookingRatingResponse{Success: true}, nil
}

func (h *Handlers) DeleteBooking(info ReqInfo, data *types.DeleteBookingRequest) (*types.DeleteBookingResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.bookings
		WHERE id = $1
	`, data.GetId())
	return &types.DeleteBookingResponse{Id: data.Id}, err
}

func (h *Handlers) DisableBooking(info ReqInfo, data *types.DisableBookingRequest) (*types.DisableBookingResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.bookings
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), info.Session.UserSub)
	return &types.DisableBookingResponse{Id: data.Id}, err
}
