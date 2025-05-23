package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupServiceAddon(info ReqInfo, data *types.PostGroupServiceAddonRequest) (*types.PostGroupServiceAddonResponse, error) {
	util.BatchExec(info.Batch, `
		INSERT INTO dbtable_schema.group_service_addons (group_id, service_addon_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		ON CONFLICT (group_id, service_addon_id) DO NOTHING
	`, info.Session.GetGroupId(), data.ServiceAddonId, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.PostGroupServiceAddonResponse{}, nil
}

func (h *Handlers) GetGroupServiceAddons(info ReqInfo, data *types.GetGroupServiceAddonsRequest) (*types.GetGroupServiceAddonsResponse, error) {
	groupServiceAddons := util.BatchQuery[types.IGroupServiceAddon](info.Batch, `
		SELECT
			egsa.id,
			egsa."groupId",
			JSON_BUILD_OBJECT(
				'id', esa.id,
				'name', esa.name,
				'createdOn', esa."createdOn"
			) as "serviceAddon"
		FROM dbview_schema.enabled_group_service_addons egsa
		LEFT JOIN dbview_schema.enabled_service_addons esa ON esa.id = egsa."serviceAddonId"
		WHERE egsa."groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupServiceAddonsResponse{GroupServiceAddons: *groupServiceAddons}, nil
}

func (h *Handlers) DeleteGroupServiceAddon(info ReqInfo, data *types.DeleteGroupServiceAddonRequest) (*types.DeleteGroupServiceAddonResponse, error) {

	res := util.BatchExec(info.Batch, `
		WITH existing_service_addon AS (
			SELECT sa.id
			FROM dbtable_schema.group_services gs
			LEFT JOIN dbtable_schema.service_tiers st ON st.service_id = gs.service_id
			LEFT JOIN dbtable_schema.service_tier_addons sta ON sta.service_tier_id = st.id
			LEFT JOIN dbtable_schema.service_addons sa ON sa.id = sta.service_addon_id
			WHERE gs.group_id = $1 AND sa.id = $2
			LIMIT 1
		)
		DELETE FROM dbtable_schema.group_service_addons gsa
		WHERE gsa.group_id = $1 AND gsa.service_addon_id = $2
		AND (SELECT COUNT(id) FROM existing_service_addon) = 0
	`, info.Session.GetGroupId(), data.ServiceAddonId)

	info.Batch.Send(info.Ctx)

	if (*res).RowsAffected() == 0 {
		return nil, util.ErrCheck(util.UserError("Deletion was skipped because the record is still associated with a group service."))
	}

	return &types.DeleteGroupServiceAddonResponse{Success: true}, nil
}
