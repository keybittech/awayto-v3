package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"errors"
	"net/http"
	"time"
)

func (h *Handlers) PostServiceTier(w http.ResponseWriter, req *http.Request, data *types.PostServiceTierRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostServiceTierResponse, error) {
	var serviceTierId string

	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.service_tiers (name, serviceId, multiplier, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
		RETURNING id
	`, data.GetName(), data.GetServiceId(), data.GetMultiplier(), session.UserSub).Scan(&serviceTierId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if serviceTierId == "" {
		return nil, util.ErrCheck(errors.New("failed to insert service tier"))
	}

	return &types.PostServiceTierResponse{Id: serviceTierId}, nil
}

func (h *Handlers) PatchServiceTier(w http.ResponseWriter, req *http.Request, data *types.PatchServiceTierRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchServiceTierResponse, error) {
	var serviceTiers []*types.IServiceTier

	err := h.Database.QueryRows(&serviceTiers, `
		UPDATE dbtable_schema.service_tiers
		SET name = $2, multiplier = $3, updated_sub = $4, updated_on = $5
		WHERE id = $1
	`, data.GetId(), data.GetName(), data.GetMultiplier(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchServiceTierResponse{Success: true}, nil
}

func (h *Handlers) GetServiceTiers(w http.ResponseWriter, req *http.Request, data *types.GetServiceTiersRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetServiceTiersResponse, error) {
	var serviceTiers []*types.IServiceTier

	err := h.Database.QueryRows(&serviceTiers, `
		SELECT * FROM dbview_schema.enabled_service_tiers
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetServiceTiersResponse{ServiceTiers: serviceTiers}, nil
}

func (h *Handlers) GetServiceTierById(w http.ResponseWriter, req *http.Request, data *types.GetServiceTierByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetServiceTierByIdResponse, error) {
	var serviceTiers []*types.IServiceTier

	err := h.Database.QueryRows(&serviceTiers, `
		SELECT * FROM dbview_schema.enabled_service_tiers_ext
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(serviceTiers) == 0 {
		return nil, util.ErrCheck(errors.New("service tier not found"))
	}

	serviceTier := serviceTiers[0]

	return &types.GetServiceTierByIdResponse{ServiceTier: serviceTier}, nil
}

func (h *Handlers) DeleteServiceTier(w http.ResponseWriter, req *http.Request, data *types.DeleteServiceTierRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteServiceTierResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.service_tiers
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteServiceTierResponse{Success: true}, nil
}

func (h *Handlers) DisableServiceTier(w http.ResponseWriter, req *http.Request, data *types.DisableServiceTierRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableServiceTierResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.service_tiers
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableServiceTierResponse{Success: true}, nil
}
