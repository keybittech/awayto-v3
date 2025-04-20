package handlers

import (
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupServiceAddon(w http.ResponseWriter, req *http.Request, data *types.PostGroupServiceAddonRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PostGroupServiceAddonResponse, error) {
	// TODO potentially undo the global uuid nature of uuid_service_addons table
	_, err := tx.Exec(`
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

func (h *Handlers) GetGroupServiceAddons(w http.ResponseWriter, req *http.Request, data *types.GetGroupServiceAddonsRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetGroupServiceAddonsResponse, error) {
	var groupServiceAddons []*types.IGroupServiceAddon

	err := h.Database.QueryRows(tx, &groupServiceAddons, `
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

func (h *Handlers) DeleteGroupServiceAddon(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupServiceAddonRequest, session *types.UserSession, tx *clients.PoolTx) (*types.DeleteGroupServiceAddonResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.uuid_service_addons
		WHERE parent_uuid = $1 AND service_addon_id = $2
	`, session.GroupId, data.GetGroupServiceAddonId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/service_addons")

	return &types.DeleteGroupServiceAddonResponse{Success: true}, nil
}
