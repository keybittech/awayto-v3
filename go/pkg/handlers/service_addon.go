package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostServiceAddon(w http.ResponseWriter, req *http.Request, data *types.PostServiceAddonRequest, session *types.UserSession, tx *sql.Tx) (*types.PostServiceAddonResponse, error) {
	var serviceAddons []*types.IServiceAddon

	err := h.Database.QueryRows(tx, &serviceAddons, `
		WITH input_rows(name, created_sub) as (VALUES ($1, $2::uuid)), ins AS (
			INSERT INTO dbtable_schema.service_addons (name, created_sub)
			SELECT * FROM input_rows
			ON CONFLICT (name) DO NOTHING
			RETURNING id, name
		)
		SELECT id, name
		FROM ins
		UNION ALL
		SELECT sa.id, sa.name
		FROM input_rows
		JOIN dbtable_schema.service_addons sa USING (name);
	`, data.GetName(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(serviceAddons) == 0 {
		return nil, util.ErrCheck(sql.ErrNoRows)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/service_addons")

	return &types.PostServiceAddonResponse{Id: serviceAddons[0].GetId()}, nil
}

func (h *Handlers) PatchServiceAddon(w http.ResponseWriter, req *http.Request, data *types.PatchServiceAddonRequest, session *types.UserSession, tx *sql.Tx) (*types.PatchServiceAddonResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.service_addons
		SET name = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
	`, data.GetId(), data.GetName(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/service_addons")

	return &types.PatchServiceAddonResponse{Success: true}, nil
}

func (h *Handlers) GetServiceAddons(w http.ResponseWriter, req *http.Request, data *types.GetServiceAddonsRequest, session *types.UserSession, tx *sql.Tx) (*types.GetServiceAddonsResponse, error) {
	var serviceAddons []*types.IServiceAddon

	err := h.Database.QueryRows(tx, &serviceAddons, `
		SELECT * FROM dbview_schema.enabled_service_addons
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetServiceAddonsResponse{ServiceAddons: serviceAddons}, nil
}

func (h *Handlers) GetServiceAddonById(w http.ResponseWriter, req *http.Request, data *types.GetServiceAddonByIdRequest, session *types.UserSession, tx *sql.Tx) (*types.GetServiceAddonByIdResponse, error) {
	var serviceAddons []*types.IServiceAddon

	err := h.Database.QueryRows(tx, &serviceAddons, `
		SELECT * FROM dbview_schema.enabled_service_addons
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(serviceAddons) == 0 {
		return nil, util.ErrCheck(sql.ErrNoRows)
	}

	return &types.GetServiceAddonByIdResponse{ServiceAddon: serviceAddons[0]}, nil
}

func (h *Handlers) DeleteServiceAddon(w http.ResponseWriter, req *http.Request, data *types.DeleteServiceAddonRequest, session *types.UserSession, tx *sql.Tx) (*types.DeleteServiceAddonResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.service_addons
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/service_addons")

	return &types.DeleteServiceAddonResponse{Success: true}, nil
}

func (h *Handlers) DisableServiceAddon(w http.ResponseWriter, req *http.Request, data *types.DisableServiceAddonRequest, session *types.UserSession, tx *sql.Tx) (*types.DisableServiceAddonResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.service_addons
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableServiceAddonResponse{Success: true}, nil
}
