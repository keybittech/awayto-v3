package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
)

func (h *Handlers) GetLookups(w http.ResponseWriter, req *http.Request, data *types.GetLookupsRequest) (*types.GetLookupsResponse, error) {
	var budgets []*types.ILookup
	var timelines []*types.ILookup
	var timeUnits []*types.ITimeUnit

	err := h.Database.QueryRows(&budgets, `
		SELECT id, name FROM dbtable_schema.budgets
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.QueryRows(&timelines, `
		SELECT id, name FROM dbtable_schema.timelines
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.QueryRows(&timeUnits, `
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
