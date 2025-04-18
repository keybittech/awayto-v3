package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostQuote(w http.ResponseWriter, req *http.Request, data *types.PostQuoteRequest, session *types.UserSession, tx *sql.Tx) (*types.PostQuoteResponse, error) {
	var slotReserved bool
	err := tx.QueryRow(`
		SELECT dbfunc_schema.is_slot_taken($1, $2)
	`, data.ScheduleBracketSlotId, data.SlotDate).Scan(&slotReserved)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if slotReserved {
		return nil, util.ErrCheck(util.UserError("The selected time has already been taken. Please select a new time."))
	}

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

	var quoteId string

	err = tx.QueryRow(`
		INSERT INTO dbtable_schema.quotes (slot_date, schedule_bracket_slot_id, service_tier_id, service_form_version_submission_id, tier_form_version_submission_id, created_sub, group_id)
		VALUES ($1::date, $2::uuid, $3::uuid, $4, $5, $6::uuid, $7::uuid)
		RETURNING id
	`, data.SlotDate, data.ScheduleBracketSlotId, data.ServiceTierId, util.NewNullString(serviceForm.Id), util.NewNullString(tierForm.Id), session.UserSub, session.GroupId).Scan(&quoteId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, file := range data.GetFiles() {
		fileRes, err := h.PostFile(w, req, &types.PostFileRequest{File: file}, session, tx)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
		_, err = tx.Exec(`
			INSERT INTO dbtable_schema.quote_files (quote_id, file_id, created_sub)
			VALUES ($1::uuid, $2::uuid, $3::uuid)
		`, quoteId, fileRes.GetId(), session.UserSub)
		if err != nil {
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

	h.Redis.Client().Del(req.Context(), staffSub+"quotes")
	h.Redis.Client().Del(req.Context(), staffSub+"profile/details")

	if err := h.Socket.RoleCall(staffSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostQuoteResponse{
		Quote: &types.IQuote{
			Id:                             quoteId,
			SlotDate:                       data.SlotDate,
			ScheduleBracketSlotId:          data.ScheduleBracketSlotId,
			ServiceFormVersionSubmissionId: serviceForm.GetId(),
			TierFormVersionSubmissionId:    tierForm.GetId(),
		},
	}, nil
}

func (h *Handlers) PatchQuote(w http.ResponseWriter, req *http.Request, data *types.PatchQuoteRequest, session *types.UserSession, tx *sql.Tx) (*types.PatchQuoteResponse, error) {
	_, err := tx.Exec(`
      UPDATE dbtable_schema.quotes
      SET service_tier_id = $2, updated_sub = $3, updated_on = $4 
      WHERE id = $1
	`, data.GetId(), data.GetServiceTierId(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchQuoteResponse{Success: true}, nil
}

func (h *Handlers) GetQuotes(w http.ResponseWriter, req *http.Request, data *types.GetQuotesRequest, session *types.UserSession, tx *sql.Tx) (*types.GetQuotesResponse, error) {
	var quotes []*types.IQuote
	err := h.Database.QueryRows(tx, &quotes, `
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

func (h *Handlers) GetQuoteById(w http.ResponseWriter, req *http.Request, data *types.GetQuoteByIdRequest, session *types.UserSession, tx *sql.Tx) (*types.GetQuoteByIdResponse, error) {
	var quotes []*types.IQuote

	err := h.Database.QueryRows(tx, &quotes, `
		SELECT * FROM dbview_schema.enabled_quotes_ext
		WHERE id = $1
	`, data.GetId())
	if err != nil || len(quotes) == 0 {
		return nil, util.ErrCheck(err)
	}

	return &types.GetQuoteByIdResponse{Quote: quotes[0]}, nil
}

func (h *Handlers) DeleteQuote(w http.ResponseWriter, req *http.Request, data *types.DeleteQuoteRequest, session *types.UserSession, tx *sql.Tx) (*types.DeleteQuoteResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.quotes
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteQuoteResponse{Success: true}, nil
}

func (h *Handlers) DisableQuote(w http.ResponseWriter, req *http.Request, data *types.DisableQuoteRequest, session *types.UserSession, tx *sql.Tx) (*types.DisableQuoteResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := tx.Exec(`
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
