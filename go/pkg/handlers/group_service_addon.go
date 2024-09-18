package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
)

func (h *Handlers) PostGroupServiceAddon(w http.ResponseWriter, req *http.Request, data *types.PostGroupServiceAddonRequest) (*types.PostGroupServiceAddonResponse, error) {
	session := h.Redis.ReqSession(req)

	// TODO fix uuid thing?
	_, err := h.Database.Client().Exec(`
		INSERT INTO dbtable_schema.uuid_service_addons (parent_uuid, service_addon_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		ON CONFLICT (parent_uuid, service_addon_id) DO NOTHING
	`, session.GroupId, data.GetServiceAddonId(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/service_addons")

	return &types.PostGroupServiceAddonResponse{}, nil
}

func (h *Handlers) GetGroupServiceAddons(w http.ResponseWriter, req *http.Request, data *types.GetGroupServiceAddonsRequest) (*types.GetGroupServiceAddonsResponse, error) {
	session := h.Redis.ReqSession(req)

	var groupServiceAddons []*types.IGroupServiceAddon

	err := h.Database.QueryRows(&groupServiceAddons, `
		SELECT eusa.id, eusa."parentUuid" as "groupId", TO_JSONB(esa.*) as "serviceAddon" 
		FROM dbview_schema.enabled_uuid_service_addons eusa
		LEFT JOIN dbview_schema.enabled_service_addons esa ON esa.id = eusa."serviceAddonId"
		WHERE eusa."parentUuid" = $1
	`, session.GroupId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupServiceAddonsResponse{GroupServiceAddons: groupServiceAddons}, nil
}

func (h *Handlers) DeleteGroupServiceAddon(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupServiceAddonRequest) (*types.DeleteGroupServiceAddonResponse, error) {
	session := h.Redis.ReqSession(req)

	_, err := h.Database.Client().Exec(`
		DELETE FROM dbtable_schema.uuid_service_addons
		WHERE parent_uuid = $1 AND service_addon_id = $2
	`, session.GroupId, data.GetGroupServiceAddonId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/service_addons")

	return &types.DeleteGroupServiceAddonResponse{Success: true}, nil
}
