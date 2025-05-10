package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupServiceAddon(info ReqInfo, data *types.PostGroupServiceAddonRequest) (*types.PostGroupServiceAddonResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.group_service_addons (group_id, service_addon_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		ON CONFLICT (group_id, service_addon_id) DO NOTHING
	`, info.Session.GroupId, data.ServiceAddonId, info.Session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/service_addons")

	return &types.PostGroupServiceAddonResponse{}, nil
}

func (h *Handlers) GetGroupServiceAddons(info ReqInfo, data *types.GetGroupServiceAddonsRequest) (*types.GetGroupServiceAddonsResponse, error) {
	groupServiceAddons, err := clients.QueryProtos[types.IGroupServiceAddon](info.Ctx, info.Tx, `
		SELECT egsa.id, egsa."groupId", TO_JSONB(esa.*) as "serviceAddon" 
		FROM dbview_schema.enabled_group_service_addons egsa
		LEFT JOIN dbview_schema.enabled_service_addons esa ON esa.id = egsa."serviceAddonId"
		WHERE egsa."groupId" = $1
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupServiceAddonsResponse{GroupServiceAddons: groupServiceAddons}, nil
}

func (h *Handlers) DeleteGroupServiceAddon(info ReqInfo, data *types.DeleteGroupServiceAddonRequest) (*types.DeleteGroupServiceAddonResponse, error) {

	res, err := info.Tx.Exec(info.Ctx, `
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
	`, info.Session.GroupId, data.ServiceAddonId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if res.RowsAffected() == 0 {
		return nil, util.ErrCheck(util.UserError("Deletion was skipped because the record is still associated with a group service."))
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/service_addons")

	return &types.DeleteGroupServiceAddonResponse{Success: true}, nil
}
