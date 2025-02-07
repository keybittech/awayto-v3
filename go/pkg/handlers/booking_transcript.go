package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
	"time"
)

func (h *Handlers) PostBookingTranscript(w http.ResponseWriter, req *http.Request, data *types.PostBookingTranscriptRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostBookingTranscriptResponse, error) {
	return &types.PostBookingTranscriptResponse{}, nil
}

func (h *Handlers) PatchBookingTranscript(w http.ResponseWriter, req *http.Request, data *types.PatchBookingTranscriptRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchBookingTranscriptResponse, error) {
	return &types.PatchBookingTranscriptResponse{}, nil
}

func (h *Handlers) GetBookingTranscripts(w http.ResponseWriter, req *http.Request, data *types.GetBookingTranscriptsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetBookingTranscriptsResponse, error) {
	return &types.GetBookingTranscriptsResponse{}, nil
}

func (h *Handlers) GetBookingTranscriptById(w http.ResponseWriter, req *http.Request, data *types.GetBookingTranscriptByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetBookingTranscriptByIdResponse, error) {
	return &types.GetBookingTranscriptByIdResponse{}, nil
}

func (h *Handlers) DeleteBookingTranscript(w http.ResponseWriter, req *http.Request, data *types.DeleteBookingTranscriptRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteBookingTranscriptResponse, error) {
	return &types.DeleteBookingTranscriptResponse{}, nil
}

func (h *Handlers) DisableBookingTranscript(w http.ResponseWriter, req *http.Request, data *types.DisableBookingTranscriptRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableBookingTranscriptResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.bookings
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableBookingTranscriptResponse{Id: data.GetId()}, nil
}
