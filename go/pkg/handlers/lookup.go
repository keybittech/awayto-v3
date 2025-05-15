package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) GetLookups(info ReqInfo, data *types.GetLookupsRequest) (*types.GetLookupsResponse, error) {
	budgets := util.BatchQuery[types.ILookup](info.Batch, `SELECT id, name FROM dbtable_schema.budgets`)
	timelines := util.BatchQuery[types.ILookup](info.Batch, `SELECT id, name FROM dbtable_schema.timelines`)
	timeUnits := util.BatchQuery[types.ITimeUnit](info.Batch, `SELECT id, name FROM dbtable_schema.time_units`)

	info.Batch.Send(info.Ctx)

	return &types.GetLookupsResponse{
		Budgets:   *budgets,
		Timelines: *timelines,
		TimeUnits: *timeUnits,
	}, nil
}
