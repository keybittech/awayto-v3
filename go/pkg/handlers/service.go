package handlers

import (
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/lib/pq"
)

func (h *Handlers) PostService(info ReqInfo, data *types.PostServiceRequest) (*types.PostServiceResponse, error) {
	service := data.GetService()

	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.services (name, cost, form_id, survey_id, created_sub)
		VALUES ($1, $2::integer, $3, $4, $5::uuid)
		ON CONFLICT (name, created_sub) DO UPDATE
		SET enabled = true, cost = $2::integer, form_id = $3, survey_id = $4
		RETURNING id
	`, service.GetName(), service.Cost, service.FormId, service.SurveyId, info.Session.GetUserSub()).Scan(&service.Id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.PatchService(info, &types.PatchServiceRequest{Service: service})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostServiceResponse{Id: service.Id}, nil
}

func (h *Handlers) PatchService(info ReqInfo, data *types.PatchServiceRequest) (*types.PatchServiceResponse, error) {
	service := data.GetService()

	// insert new tiers, re-enabling if conflicting
	insertedTierIds := make([]string, 0)
	for _, tier := range service.GetTiers() {
		var tierId string

		err := info.Tx.QueryRow(info.Ctx, `
			WITH input_rows(name, service_id, multiplier, form_id, survey_id, created_sub) as (VALUES ($1, $2::uuid, $3::decimal, $4::uuid, $5::uuid, $6::uuid)), ins AS (
				INSERT INTO dbtable_schema.service_tiers (name, service_id, multiplier, form_id, survey_id, created_sub)
				SELECT name, service_id, multiplier, form_id, survey_id, created_sub FROM input_rows
				ON CONFLICT (name, service_id) DO UPDATE
				SET enabled = true, multiplier = $3::decimal, form_id = $4::uuid, survey_id = $5::uuid, updated_sub = $6::uuid, updated_on = $7
				RETURNING id
			)
			SELECT id
			FROM ins
			UNION ALL
			SELECT st.id
			FROM input_rows
			JOIN dbtable_schema.service_tiers st USING (name, service_id)
		`, tier.GetName(), service.Id, tier.GetMultiplier(), tier.FormId, tier.SurveyId, info.Session.GetUserSub(), time.Now()).Scan(&tierId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		insertedTierIds = append(insertedTierIds, tierId)

		// insert new tier addons, enabling on conflict
		insertedTierAddonIds := make([]string, 0)
		for _, addon := range tier.GetAddons() {
			_, err = info.Tx.Exec(info.Ctx, `
				INSERT INTO dbtable_schema.service_tier_addons (service_addon_id, service_tier_id, created_sub)
				VALUES ($1, $2, $3::uuid)
				ON CONFLICT (service_addon_id, service_tier_id) DO UPDATE
				SET enabled = true
			`, addon.GetId(), tierId, info.Session.GetUserSub())
			if err != nil {
				return nil, util.ErrCheck(err)
			}
			insertedTierAddonIds = append(insertedTierAddonIds, addon.GetId())
		}

		// delete old addons, never referenced beyond the tier
		_, err = info.Tx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.service_tier_addons
			WHERE service_addon_id IN (
				SELECT service_addon_id
				FROM dbtable_schema.service_tier_addons
				WHERE service_tier_id = $1
				AND service_addon_id NOT IN (SELECT unnest($2::uuid[]))
			)
		`, tierId, pq.Array(insertedTierAddonIds))
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	// disable tiers that were not inserted or re-enabled
	// may be referenced by a quote, which must show accurate info at the time of the request
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.service_tiers
		SET enabled = false
		WHERE id IN (
			SELECT id
			FROM dbtable_schema.service_tiers
			WHERE service_id = $1
			AND id NOT IN (SELECT unnest($2::uuid[]))
		)
	`, service.Id, pq.Array(insertedTierIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// update service
	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.services
		SET name = $2, form_id = $3, survey_id = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
	`, service.GetId(), service.GetName(), service.FormId, service.SurveyId, info.Session.GetUserSub(), time.Now())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchServiceResponse{Success: true}, nil
}

func (h *Handlers) GetServices(info ReqInfo, data *types.GetServicesRequest) (*types.GetServicesResponse, error) {
	services := util.BatchQuery[types.IService](info.Batch, `
		SELECT id, name, "formId", "surveyId", "createdOn"
		FROM dbview_schema.enabled_services
		WHERE "createdSub" = $1
	`, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.GetServicesResponse{Services: *services}, nil
}

func (h *Handlers) GetServiceById(info ReqInfo, data *types.GetServiceByIdRequest) (*types.GetServiceByIdResponse, error) {
	service := util.BatchQueryRow[types.IService](info.Batch, `
		SELECT id, name, "formId", "surveyId", "createdOn", tiers
		FROM dbview_schema.enabled_services_ext
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.GetServiceByIdResponse{Service: *service}, nil
}

func (h *Handlers) DeleteService(info ReqInfo, data *types.DeleteServiceRequest) (*types.DeleteServiceResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.services
		WHERE id = ANY($1)
	`, pq.Array(strings.Split(data.GetIds(), ",")))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteServiceResponse{Success: true}, nil
}

func (h *Handlers) DisableService(info ReqInfo, data *types.DisableServiceRequest) (*types.DisableServiceResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.services
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = ANY($1)
	`, pq.Array(strings.Split(data.GetIds(), ",")), time.Now(), info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.DisableServiceResponse{Success: true}, nil
}
