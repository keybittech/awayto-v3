package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostServiceAddon(info ReqInfo, data *types.PostServiceAddonRequest) (*types.PostServiceAddonResponse, error) {
	serviceAddonInsert := util.BatchQueryRow[types.IServiceAddon](info.Batch, `
		WITH input_rows(name, created_sub) as (VALUES ($1, $2::uuid)), ins AS (
			INSERT INTO dbtable_schema.service_addons (name, created_sub)
			SELECT name, created_sub FROM input_rows
			ON CONFLICT (name) DO NOTHING
			RETURNING id, name
		)
		SELECT id, name
		FROM ins
		UNION ALL
		SELECT sa.id, sa.name
		FROM input_rows
		JOIN dbtable_schema.service_addons sa USING (name);
	`, data.GetName(), info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/service_addons")

	return &types.PostServiceAddonResponse{Id: (*serviceAddonInsert).Id}, nil
}

func (h *Handlers) PatchServiceAddon(info ReqInfo, data *types.PatchServiceAddonRequest) (*types.PatchServiceAddonResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.service_addons
		SET name = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
	`, data.Id, data.Name, info.Session.GetUserSub(), time.Now())

	info.Batch.Send(info.Ctx)

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/service_addons")

	return &types.PatchServiceAddonResponse{Success: true}, nil
}

func (h *Handlers) GetServiceAddons(info ReqInfo, data *types.GetServiceAddonsRequest) (*types.GetServiceAddonsResponse, error) {
	serviceAddons := util.BatchQuery[types.IServiceAddon](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_service_addons
		WHERE "createdSub" = $1
	`, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.GetServiceAddonsResponse{ServiceAddons: *serviceAddons}, nil
}

func (h *Handlers) GetServiceAddonById(info ReqInfo, data *types.GetServiceAddonByIdRequest) (*types.GetServiceAddonByIdResponse, error) {
	serviceAddon := util.BatchQueryRow[types.IServiceAddon](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_service_addons
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.GetServiceAddonByIdResponse{ServiceAddon: *serviceAddon}, nil
}

func (h *Handlers) DeleteServiceAddon(info ReqInfo, data *types.DeleteServiceAddonRequest) (*types.DeleteServiceAddonResponse, error) {
	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.service_addons
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/service_addons")

	return &types.DeleteServiceAddonResponse{Success: true}, nil
}

func (h *Handlers) DisableServiceAddon(info ReqInfo, data *types.DisableServiceAddonRequest) (*types.DisableServiceAddonResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.service_addons
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.Id, time.Now(), info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/service_addons")

	return &types.DisableServiceAddonResponse{Success: true}, nil
}
