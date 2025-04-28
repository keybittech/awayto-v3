package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupServiceAddon(info ReqInfo, data *types.PostGroupServiceAddonRequest) (*types.PostGroupServiceAddonResponse, error) {
	// TODO potentially undo the global uuid nature of uuid_service_addons table
	_, err := info.Tx.Exec(info.Req.Context(), `
		INSERT INTO dbtable_schema.uuid_service_addons (parent_uuid, service_addon_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		ON CONFLICT (parent_uuid, service_addon_id) DO NOTHING
	`, info.Session.GroupId, data.GetServiceAddonId(), info.Session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/service_addons")

	return &types.PostGroupServiceAddonResponse{}, nil
}

func (h *Handlers) GetGroupServiceAddons(info ReqInfo, data *types.GetGroupServiceAddonsRequest) (*types.GetGroupServiceAddonsResponse, error) {
	var groupServiceAddons []*types.IGroupServiceAddon

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &groupServiceAddons, `
		SELECT eusa.id, eusa."parentUuid" as "groupId", TO_JSONB(esa.*) as "serviceAddon" 
		FROM dbview_schema.enabled_uuid_service_addons eusa
		LEFT JOIN dbview_schema.enabled_service_addons esa ON esa.id = eusa."serviceAddonId"
		WHERE eusa."parentUuid" = $1
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupServiceAddonsResponse{GroupServiceAddons: groupServiceAddons}, nil
}

func (h *Handlers) DeleteGroupServiceAddon(info ReqInfo, data *types.DeleteGroupServiceAddonRequest) (*types.DeleteGroupServiceAddonResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		DELETE FROM dbtable_schema.uuid_service_addons
		WHERE parent_uuid = $1 AND service_addon_id = $2
	`, info.Session.GroupId, data.GetGroupServiceAddonId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/service_addons")

	return &types.DeleteGroupServiceAddonResponse{Success: true}, nil
}
