package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostQuote(info ReqInfo, data *types.PostQuoteRequest) (*types.PostQuoteResponse, error) {
	var slotReserved bool
	err := info.Tx.QueryRow(info.Ctx, `
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

			err = info.Tx.QueryRow(info.Ctx, `
				INSERT INTO dbtable_schema.form_version_submissions (form_version_id, submission, created_sub)
				VALUES ($1, $2::jsonb, $3::uuid)
				RETURNING id
			`, form.GetFormVersionId(), formSubmission, info.Session.UserSub).Scan(&form.Id)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	var quoteId string

	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.quotes (slot_date, schedule_bracket_slot_id, service_tier_id, service_form_version_submission_id, tier_form_version_submission_id, created_sub, group_id)
		VALUES ($1::date, $2::uuid, $3::uuid, $4, $5, $6::uuid, $7::uuid)
		RETURNING id
	`, data.SlotDate, data.ScheduleBracketSlotId, data.ServiceTierId, util.NewNullString(serviceForm.Id), util.NewNullString(tierForm.Id), info.Session.UserSub, info.Session.GroupId).Scan(&quoteId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, file := range data.GetFiles() {
		fileRes, err := h.PostFile(info, &types.PostFileRequest{File: file})
		if err != nil {
			return nil, util.ErrCheck(err)
		}
		_, err = info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.quote_files (quote_id, file_id, created_sub)
			VALUES ($1::uuid, $2::uuid, $3::uuid)
		`, quoteId, fileRes.GetId(), info.Session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	var staffSub string

	err = info.Tx.QueryRow(info.Ctx, `
		SELECT created_sub
		FROM dbtable_schema.schedule_bracket_slots
		WHERE id = $1
	`, data.GetScheduleBracketSlotId()).Scan(&staffSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, staffSub+"quotes")
	h.Redis.Client().Del(info.Ctx, staffSub+"profile/details")

	if err := h.Socket.RoleCall(info.Ctx, staffSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostQuoteResponse{
		Quote: &types.IQuote{
			Id:                             quoteId,
			SlotDate:                       data.SlotDate,
			ScheduleBracketSlotId:          data.ScheduleBracketSlotId,
			ServiceFormVersionSubmissionId: &serviceForm.Id,
			TierFormVersionSubmissionId:    &tierForm.Id,
		},
	}, nil
}

func (h *Handlers) PatchQuote(info ReqInfo, data *types.PatchQuoteRequest) (*types.PatchQuoteResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
      UPDATE dbtable_schema.quotes
      SET service_tier_id = $2, updated_sub = $3, updated_on = $4 
      WHERE id = $1
	`, data.GetId(), data.GetServiceTierId(), info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchQuoteResponse{Success: true}, nil
}

func (h *Handlers) GetQuotes(info ReqInfo, data *types.GetQuotesRequest) (*types.GetQuotesResponse, error) {
	quotes, err := clients.QueryProtos[types.IQuote](info.Ctx, info.Tx, `
		SELECT
			q.id,
			q."startTime",
			q."slotDate"::TEXT,
			q."scheduleBracketSlotId",
			q."serviceTierId",
			q."serviceTierName",
			q."serviceName",
			q."serviceFormVersionSubmissionId",
			q."tierFormVersionSubmissionId",
			q."createdOn"
		FROM dbview_schema.enabled_quotes q
		JOIN dbtable_schema.schedule_bracket_slots sbs ON sbs.id = q."scheduleBracketSlotId"
		WHERE sbs.created_sub = $1
	`, info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetQuotesResponse{Quotes: quotes}, nil
}

func (h *Handlers) GetQuoteById(info ReqInfo, data *types.GetQuoteByIdRequest) (*types.GetQuoteByIdResponse, error) {
	quote, err := clients.QueryProto[types.IQuote](info.Ctx, info.Tx, `
		SELECT id, "slotDate", "serviceFormVersionSubmission", "tierFormVersionSubmission", "createdOn"
		FROM dbview_schema.enabled_quotes_ext
		WHERE id = $1
	`, data.Id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetQuoteByIdResponse{Quote: quote}, nil
}

func (h *Handlers) DeleteQuote(info ReqInfo, data *types.DeleteQuoteRequest) (*types.DeleteQuoteResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.quotes
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteQuoteResponse{Success: true}, nil
}

func (h *Handlers) DisableQuote(info ReqInfo, data *types.DisableQuoteRequest) (*types.DisableQuoteResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := info.Tx.Exec(info.Ctx, `
			UPDATE dbtable_schema.quotes
			SET enabled = false, updated_on = $2, updated_sub = $3
			WHERE id = $1
		`, id, time.Now().Local().UTC(), info.Session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"quotes")
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")

	return &types.DisableQuoteResponse{Success: true}, nil
}
