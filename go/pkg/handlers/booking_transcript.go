package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostBookingTranscript(info ReqInfo, data *types.PostBookingTranscriptRequest) (*types.PostBookingTranscriptResponse, error) {
	return &types.PostBookingTranscriptResponse{}, nil
}

func (h *Handlers) PatchBookingTranscript(info ReqInfo, data *types.PatchBookingTranscriptRequest) (*types.PatchBookingTranscriptResponse, error) {
	return &types.PatchBookingTranscriptResponse{}, nil
}

func (h *Handlers) GetBookingTranscripts(info ReqInfo, data *types.GetBookingTranscriptsRequest) (*types.GetBookingTranscriptsResponse, error) {
	return &types.GetBookingTranscriptsResponse{}, nil
}

func (h *Handlers) GetBookingTranscriptById(info ReqInfo, data *types.GetBookingTranscriptByIdRequest) (*types.GetBookingTranscriptByIdResponse, error) {
	return &types.GetBookingTranscriptByIdResponse{}, nil
}

func (h *Handlers) DeleteBookingTranscript(info ReqInfo, data *types.DeleteBookingTranscriptRequest) (*types.DeleteBookingTranscriptResponse, error) {
	return &types.DeleteBookingTranscriptResponse{}, nil
}

func (h *Handlers) DisableBookingTranscript(info ReqInfo, data *types.DisableBookingTranscriptRequest) (*types.DisableBookingTranscriptResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.bookings
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), info.Session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableBookingTranscriptResponse{Id: data.GetId()}, nil
}
