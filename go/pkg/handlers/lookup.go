package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) GetLookups(info ReqInfo, data *types.GetLookupsRequest) (*types.GetLookupsResponse, error) {
	var budgets []*types.ILookup
	var timelines []*types.ILookup
	var timeUnits []*types.ITimeUnit

	err := h.Database.QueryRows(info.Ctx, info.Tx, &budgets, `
		SELECT id, name FROM dbtable_schema.budgets
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.QueryRows(info.Ctx, info.Tx, &timelines, `
		SELECT id, name FROM dbtable_schema.timelines
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.QueryRows(info.Ctx, info.Tx, &timeUnits, `
		SELECT id, name FROM dbtable_schema.time_units
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetLookupsResponse{
		Budgets:   budgets,
		Timelines: timelines,
		TimeUnits: timeUnits,
	}, nil
}
