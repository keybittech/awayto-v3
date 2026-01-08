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
		INSERT INTO dbtable_schema.services (name, cost, created_sub)
		VALUES ($1, $2::integer, $3::uuid)
		ON CONFLICT (name, created_sub) DO UPDATE
		SET enabled = true, cost = $2::integer
		RETURNING id
	`, service.GetName(), service.GetCost(), info.Session.GetUserSub()).Scan(&service.Id) // , service.FormId, service.SurveyId
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
	serviceId := service.GetId()
	userSub := info.Session.GetUserSub()

	// update service forms
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.service_forms
		WHERE service_id = $1::uuid
	`, serviceId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, formId := range service.GetIntakeIds() {
		_, err := info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.service_forms (service_id, form_id, stage, created_sub)
			VALUES ($1::uuid, $2::uuid, $3, $4::uuid)
		`, serviceId, formId, "intake", userSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	for _, formId := range service.GetSurveyIds() {
		_, err := info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.service_forms (service_id, form_id, stage, created_sub)
			VALUES ($1::uuid, $2::uuid, $3, $4::uuid)
		`, serviceId, formId, "survey", userSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	// insert new tiers, re-enabling if conflicting
	insertedTierIds := make([]string, 0)
	for _, tier := range service.GetTiers() {
		var tierId string

		err = info.Tx.QueryRow(info.Ctx, `
			WITH input_rows(name, service_id, multiplier, created_sub) as (VALUES ($1, $2::uuid, $3::decimal, $4::uuid)), ins AS (
				INSERT INTO dbtable_schema.service_tiers (name, service_id, multiplier, created_sub)
				SELECT name, service_id, multiplier, created_sub FROM input_rows
				ON CONFLICT (name, service_id) DO UPDATE
				SET enabled = true, multiplier = $3::decimal, updated_sub = $4::uuid, updated_on = $5
				RETURNING id
			)
			SELECT id
			FROM ins
			UNION ALL
			SELECT st.id
			FROM input_rows
			JOIN dbtable_schema.service_tiers st USING (name, service_id)
		`, tier.GetName(), serviceId, tier.GetMultiplier(), userSub, time.Now()).Scan(&tierId)
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
			`, addon.GetId(), tierId, userSub)
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

		// update tier forms
		_, err := info.Tx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.service_tier_forms
			WHERE service_tier_id = $1::uuid
		`, tierId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, formId := range tier.GetIntakeIds() {
			_, err := info.Tx.Exec(info.Ctx, `
				INSERT INTO dbtable_schema.service_tier_forms (service_tier_id, form_id, stage, created_sub)
				VALUES ($1::uuid, $2::uuid, $3, $4::uuid)
			`, tierId, formId, "intake", userSub)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}

		for _, formId := range tier.GetSurveyIds() {
			_, err := info.Tx.Exec(info.Ctx, `
				INSERT INTO dbtable_schema.service_tier_forms (service_tier_id, form_id, stage, created_sub)
				VALUES ($1::uuid, $2::uuid, $3, $4::uuid)
			`, tierId, formId, "survey", userSub)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	// disable tiers that were not inserted or re-enabled
	// may be referenced by a quote, which must show accurate info at the time of the request
	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.service_tiers
		SET enabled = false
		WHERE id IN (
			SELECT id
			FROM dbtable_schema.service_tiers
			WHERE service_id = $1
			AND id NOT IN (SELECT unnest($2::uuid[]))
		)
	`, serviceId, pq.Array(insertedTierIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// update service
	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.services
		SET name = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
	`, serviceId, service.GetName(), userSub, time.Now())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchServiceResponse{Success: true}, nil
}

func (h *Handlers) GetServices(info ReqInfo, data *types.GetServicesRequest) (*types.GetServicesResponse, error) {
	services := util.BatchQuery[types.IService](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_services
		WHERE "createdSub" = $1
	`, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.GetServicesResponse{Services: *services}, nil
}

func (h *Handlers) GetServiceById(info ReqInfo, data *types.GetServiceByIdRequest) (*types.GetServiceByIdResponse, error) {
	service := util.BatchQueryRow[types.IService](info.Batch, `
		SELECT id, name, "intakeIds", "surveyIds", "createdOn", tiers
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
