package handlers

import (
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) GetLookups(w http.ResponseWriter, req *http.Request, data *types.GetLookupsRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetLookupsResponse, error) {
	var budgets []*types.ILookup
	var timelines []*types.ILookup
	var timeUnits []*types.ITimeUnit

	err := h.Database.QueryRows(tx, &budgets, `
		SELECT id, name FROM dbtable_schema.budgets
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.QueryRows(tx, &timelines, `
		SELECT id, name FROM dbtable_schema.timelines
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.QueryRows(tx, &timeUnits, `
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
