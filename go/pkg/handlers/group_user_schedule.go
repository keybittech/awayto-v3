package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupUserSchedule(info ReqInfo, data *types.PostGroupUserScheduleRequest) (*types.PostGroupUserScheduleResponse, error) {
	var groupScheduleId string
	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_user_schedules (group_schedule_id, user_schedule_id, created_sub, group_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid)
		ON CONFLICT (group_schedule_id, user_schedule_id) DO UPDATE SET updated_on = NOW()
		RETURNING id
	`, data.GroupScheduleId, data.UserScheduleId, info.Session.GetUserSub(), info.Session.GetGroupId()).Scan(&groupScheduleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostGroupUserScheduleResponse{Id: groupScheduleId}, nil
}

func (h *Handlers) GetGroupUserSchedules(info ReqInfo, data *types.GetGroupUserSchedulesRequest) (*types.GetGroupUserSchedulesResponse, error) {
	groupUserSchedules := util.BatchQuery[types.IGroupUserSchedule](info.Batch, `
		SELECT egus.id, egus."groupScheduleId", egus."userScheduleId", brackets
		FROM dbview_schema.enabled_group_user_schedules_ext egus
		WHERE egus."groupScheduleId" = $1
	`, data.GroupScheduleId)

	info.Batch.Send(info.Ctx)

	return &types.GetGroupUserSchedulesResponse{GroupUserSchedules: *groupUserSchedules}, nil
}

func (h *Handlers) GetGroupUserScheduleStubs(info ReqInfo, data *types.GetGroupUserScheduleStubsRequest) (*types.GetGroupUserScheduleStubsResponse, error) {
	info.Batch.Sub = fmt.Sprint(info.Session.GetGroupSub())
	info.Batch.Reset()

	groupUserScheduleStubs := util.BatchQuery[types.IGroupUserScheduleStub](info.Batch, `
		SELECT "userScheduleId", "quoteId", "slotDate", "startTime", "serviceName", "tierName", replacement
		FROM dbview_schema.group_user_schedule_stubs
		WHERE "groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupUserScheduleStubsResponse{GroupUserScheduleStubs: *groupUserScheduleStubs}, nil
}

func (h *Handlers) GetGroupUserScheduleStubReplacement(info ReqInfo, data *types.GetGroupUserScheduleStubReplacementRequest) (*types.GetGroupUserScheduleStubReplacementResponse, error) {
	replacement := util.BatchQuery[types.IGroupUserScheduleStub](info.Batch, `
		SELECT replacement
		FROM dbfunc_schema.get_peer_schedule_replacement($1::UUID[], $2::DATE, $3::INTERVAL, $4::TEXT)
	`, data.UserScheduleId, data.SlotDate, data.StartTime, data.TierName)

	info.Batch.Send(info.Ctx)

	return &types.GetGroupUserScheduleStubReplacementResponse{GroupUserScheduleStubs: *replacement}, nil
}

func (h *Handlers) PatchGroupUserScheduleStubReplacement(info ReqInfo, data *types.PatchGroupUserScheduleStubReplacementRequest) (*types.PatchGroupUserScheduleStubReplacementResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.quotes
		SET slot_date = $2, schedule_bracket_slot_id = $3, service_tier_id = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
	`, data.QuoteId, data.SlotDate, data.ScheduleBracketSlotId, data.ServiceTierId, info.Session.GetUserSub(), time.Now())

	info.Batch.Send(info.Ctx)

	return &types.PatchGroupUserScheduleStubReplacementResponse{Success: true}, nil
}

func (h *Handlers) DeleteGroupUserScheduleByUserScheduleId(info ReqInfo, data *types.DeleteGroupUserScheduleByUserScheduleIdRequest) (*types.DeleteGroupUserScheduleByUserScheduleIdResponse, error) {
	for userScheduleId := range strings.SplitSeq(data.Ids, ",") {
		rows, err := info.Tx.Query(info.Ctx, `
			SELECT "partType", ids
			FROM dbfunc_schema.get_scheduled_parts($1)
		`, userScheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		parts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[types.ScheduledParts])
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		hasParts := false
		for _, part := range parts {
			if len(part.Ids) > 0 {
				hasParts = true
				break
			}
		}

		if !hasParts {
			_, err = info.Tx.Exec(info.Ctx, `
				DELETE FROM dbtable_schema.group_user_schedules
				WHERE user_schedule_id = $1
			`, userScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		} else {
			_, err = info.Tx.Exec(info.Ctx, `
				UPDATE dbtable_schema.group_user_schedules
				SET enabled = false
				WHERE user_schedule_id = $1
			`, userScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	return &types.DeleteGroupUserScheduleByUserScheduleIdResponse{Success: true}, nil
}
