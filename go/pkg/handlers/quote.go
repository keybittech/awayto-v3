package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (h *Handlers) PostQuote(w http.ResponseWriter, req *http.Request, data *types.PostQuoteRequest) (*types.PostQuoteResponse, error) {
	session := h.Redis.ReqSession(req)

	tx, err := h.Database.Client().Begin()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer tx.Rollback()

	serviceForm, tierForm := data.GetServiceFormVersionSubmission(), data.GetTierFormVersionSubmission()

	for _, form := range []*types.IProtoFormVersionSubmission{serviceForm, tierForm} {
		if form.Submission != nil {
			formSubmission, err := json.Marshal(form.GetSubmission())

			if err != nil {
				return nil, util.ErrCheck(err)
			}

			err = tx.QueryRow(`
				INSERT INTO dbtable_schema.form_version_submissions (form_version_id, submission, created_sub)
				VALUES ($1, $2::jsonb, $3::uuid)
				RETURNING id
			`, form.GetFormVersionId(), formSubmission, session.UserSub).Scan(&form.Id)

			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	var serviceFormId, tierFormId sql.NullString

	if serviceForm.GetId() != "" {
		serviceFormId.Scan(&serviceForm.Id)
	}

	if tierForm.GetId() != "" {
		tierFormId.Scan(&tierForm.Id)
	}

	var quoteId string

	err = tx.QueryRow(`
		INSERT INTO dbtable_schema.quotes (slot_date, schedule_bracket_slot_id, service_tier_id, service_form_version_submission_id, tier_form_version_submission_id, created_sub)
		VALUES ($1::date, $2::uuid, $3::uuid, $4, $5, $6::uuid)
		RETURNING id
	`, data.GetSlotDate(), data.GetScheduleBracketSlotId(), data.GetServiceTierId(), serviceFormId, tierFormId, session.UserSub).Scan(&quoteId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, file := range data.GetFiles() {
		fileRes, err := h.PostFile(w, req, &types.PostFileRequest{File: file})
		if err != nil {
			return nil, util.ErrCheck(err)
		}
		if _, err = tx.Exec(`
				INSERT INTO dbtable_schema.quote_files (quote_id, file_id, created_sub)
				VALUES ($1::uuid, $2::uuid, $3::uuid)
			`, quoteId, fileRes.GetId(), session.UserSub); err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	var staffSub string

	err = tx.QueryRow(`
		SELECT created_sub
		FROM dbtable_schema.schedule_bracket_slots
		WHERE id = $1
	`, data.GetScheduleBracketSlotId()).Scan(&staffSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, util.ErrCheck(err)
	}

	if err := h.Keycloak.RoleCall(http.MethodPost, staffSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), staffSub+"quotes")
	h.Redis.Client().Del(req.Context(), staffSub+"profile/details")

	return &types.PostQuoteResponse{Quote: &types.IQuote{
		Id:                             quoteId,
		ServiceFormVersionSubmissionId: serviceForm.GetId(),
		TierFormVersionSubmissionId:    tierForm.GetId(),
	}}, nil
}

func (h *Handlers) PatchQuote(w http.ResponseWriter, req *http.Request, data *types.PatchQuoteRequest) (*types.PatchQuoteResponse, error) {
	session := h.Redis.ReqSession(req)

	_, err := h.Database.Client().Exec(`
      UPDATE dbtable_schema.quotes
      SET service_tier_id = $2, updated_sub = $3, updated_on = $4 
      WHERE id = $1
	`, data.GetId(), data.GetServiceTierId(), session.UserSub, time.Now().Local().UTC())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchQuoteResponse{Success: true}, nil
}

func (h *Handlers) GetQuotes(w http.ResponseWriter, req *http.Request, data *types.GetQuotesRequest) (*types.GetQuotesResponse, error) {
	session := h.Redis.ReqSession(req)

	var quotes []*types.IQuote
	err := h.Database.QueryRows(&quotes, `
		SELECT q.*
		FROM dbview_schema.enabled_quotes q
		JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = q.schedule_bracket_slot_id
		WHERE sbs.created_sub = $1
	`, session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetQuotesResponse{Quotes: quotes}, nil
}

func (h *Handlers) GetQuoteById(w http.ResponseWriter, req *http.Request, data *types.GetQuoteByIdRequest) (*types.GetQuoteByIdResponse, error) {
	var quotes []*types.IQuote

	err := h.Database.QueryRows(&quotes, `
		SELECT * FROM dbview_schema.enabled_quotes_ext
		WHERE id = $1
	`, data.GetId())

	if err != nil || len(quotes) == 0 {
		return nil, util.ErrCheck(err)
	}

	return &types.GetQuoteByIdResponse{Quote: quotes[0]}, nil
}

func (h *Handlers) DeleteQuote(w http.ResponseWriter, req *http.Request, data *types.DeleteQuoteRequest) (*types.DeleteQuoteResponse, error) {

	_, err := h.Database.Client().Exec(`
		DELETE FROM dbtable_schema.quotes
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteQuoteResponse{Success: true}, nil
}

func (h *Handlers) DisableQuote(w http.ResponseWriter, req *http.Request, data *types.DisableQuoteRequest) (*types.DisableQuoteResponse, error) {
	session := h.Redis.ReqSession(req)
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := h.Database.Client().Exec(`
			UPDATE dbtable_schema.quotes
			SET enabled = false, updated_on = $2, updated_sub = $3
			WHERE id = $1
		`, id, time.Now().Local().UTC(), session.UserSub)

		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"quotes")
	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	return &types.DisableQuoteResponse{Success: true}, nil
}
