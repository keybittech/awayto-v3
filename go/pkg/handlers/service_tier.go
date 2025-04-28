package handlers

import (
	"errors"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostServiceTier(info ReqInfo, data *types.PostServiceTierRequest) (*types.PostServiceTierResponse, error) {
	var serviceTierId string

	err := info.Tx.QueryRow(info.Req.Context(), `
		INSERT INTO dbtable_schema.service_tiers (name, serviceId, multiplier, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
		RETURNING id
	`, data.GetName(), data.GetServiceId(), data.GetMultiplier(), info.Session.UserSub).Scan(&serviceTierId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if serviceTierId == "" {
		return nil, util.ErrCheck(errors.New("failed to insert service tier"))
	}

	return &types.PostServiceTierResponse{Id: serviceTierId}, nil
}

func (h *Handlers) PatchServiceTier(info ReqInfo, data *types.PatchServiceTierRequest) (*types.PatchServiceTierResponse, error) {
	var serviceTiers []*types.IServiceTier

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &serviceTiers, `
		UPDATE dbtable_schema.service_tiers
		SET name = $2, multiplier = $3, updated_sub = $4, updated_on = $5
		WHERE id = $1
	`, data.GetId(), data.GetName(), data.GetMultiplier(), info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchServiceTierResponse{Success: true}, nil
}

func (h *Handlers) GetServiceTiers(info ReqInfo, data *types.GetServiceTiersRequest) (*types.GetServiceTiersResponse, error) {
	var serviceTiers []*types.IServiceTier

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &serviceTiers, `
		SELECT * FROM dbview_schema.enabled_service_tiers
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetServiceTiersResponse{ServiceTiers: serviceTiers}, nil
}

func (h *Handlers) GetServiceTierById(info ReqInfo, data *types.GetServiceTierByIdRequest) (*types.GetServiceTierByIdResponse, error) {
	var serviceTiers []*types.IServiceTier

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &serviceTiers, `
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

func (h *Handlers) DeleteServiceTier(info ReqInfo, data *types.DeleteServiceTierRequest) (*types.DeleteServiceTierResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		DELETE FROM dbtable_schema.service_tiers
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteServiceTierResponse{Success: true}, nil
}

func (h *Handlers) DisableServiceTier(info ReqInfo, data *types.DisableServiceTierRequest) (*types.DisableServiceTierResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.service_tiers
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableServiceTierResponse{Success: true}, nil
}
