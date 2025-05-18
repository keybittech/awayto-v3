package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostServiceTier(info ReqInfo, data *types.PostServiceTierRequest) (*types.PostServiceTierResponse, error) {
	serviceTierInsert := util.BatchQueryRow[types.ILookup](info.Batch, `
		INSERT INTO dbtable_schema.service_tiers (name, serviceId, multiplier, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
		RETURNING id
	`, data.Name, data.ServiceId, data.Multiplier, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.PostServiceTierResponse{Id: (*serviceTierInsert).Id}, nil
}

func (h *Handlers) PatchServiceTier(info ReqInfo, data *types.PatchServiceTierRequest) (*types.PatchServiceTierResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.service_tiers
		SET name = $2, multiplier = $3, updated_sub = $4, updated_on = $5
		WHERE id = $1
	`, data.Id, data.Name, data.Multiplier, info.Session.GetUserSub(), time.Now())

	info.Batch.Send(info.Ctx)

	return &types.PatchServiceTierResponse{Success: true}, nil
}

func (h *Handlers) GetServiceTiers(info ReqInfo, data *types.GetServiceTiersRequest) (*types.GetServiceTiersResponse, error) {
	serviceTiers := util.BatchQuery[types.IServiceTier](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_service_tiers
		WHERE "createdSub" = $1
	`, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.GetServiceTiersResponse{ServiceTiers: *serviceTiers}, nil
}

func (h *Handlers) GetServiceTierById(info ReqInfo, data *types.GetServiceTierByIdRequest) (*types.GetServiceTierByIdResponse, error) {
	serviceTier := util.BatchQueryRow[types.IServiceTier](info.Batch, `
		SELECT id, name, multiplier, addons, "serviceId", "formId", "surveyId", "createdOn"
		FROM dbview_schema.enabled_service_tiers_ext
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.GetServiceTierByIdResponse{ServiceTier: *serviceTier}, nil
}

func (h *Handlers) DeleteServiceTier(info ReqInfo, data *types.DeleteServiceTierRequest) (*types.DeleteServiceTierResponse, error) {
	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.service_tiers
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.DeleteServiceTierResponse{Success: true}, nil
}

func (h *Handlers) DisableServiceTier(info ReqInfo, data *types.DisableServiceTierRequest) (*types.DisableServiceTierResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.service_tiers
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.Id, time.Now(), info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.DisableServiceTierResponse{Success: true}, nil
}
