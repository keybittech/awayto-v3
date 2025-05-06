package handlers

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupUserSchedule(info ReqInfo, data *types.PostGroupUserScheduleRequest) (*types.PostGroupUserScheduleResponse, error) {
	var groupUserScheduleId string

	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_user_schedules (group_schedule_id, user_schedule_id, created_sub, group_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid)
		ON CONFLICT (group_schedule_id, user_schedule_id) DO UPDATE SET updated_on = NOW()
		RETURNING id
	`, data.GetGroupScheduleId(), data.GetUserScheduleId(), info.Session.UserSub, info.Session.GroupId).Scan(&groupUserScheduleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/schedules")
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/user_schedules/"+data.GetGroupScheduleId())
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/user_schedules_stubs")

	return &types.PostGroupUserScheduleResponse{Id: groupUserScheduleId}, nil
}

func (h *Handlers) GetGroupUserSchedules(info ReqInfo, data *types.GetGroupUserSchedulesRequest) (*types.GetGroupUserSchedulesResponse, error) {
	var groupUserSchedules []*types.IGroupUserSchedule

	err := h.Database.QueryRows(info.Ctx, info.Tx, &groupUserSchedules, `
		SELECT egus.*
		FROM dbview_schema.enabled_group_user_schedules_ext egus
		WHERE egus."groupScheduleId" = $1
	`, data.GetGroupScheduleId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUserSchedulesResponse{GroupUserSchedules: groupUserSchedules}, nil
}

func (h *Handlers) GetGroupUserScheduleStubs(info ReqInfo, data *types.GetGroupUserScheduleStubsRequest) (*types.GetGroupUserScheduleStubsResponse, error) {

	ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, info.Session)

	rows, done, err := ds.SessionBatchQuery(info.Ctx, `
		SELECT guss.*, gus.group_schedule_id as "groupScheduleId"
		FROM dbview_schema.group_user_schedule_stubs guss
		JOIN dbtable_schema.group_user_schedules gus ON gus.user_schedule_id = guss."userScheduleId"
		JOIN dbtable_schema.schedules schedule ON schedule.id = gus.group_schedule_id
		JOIN dbview_schema.enabled_users eu ON eu.sub = schedule.created_sub
		JOIN dbtable_schema.users u ON u.id = eu.id
		WHERE u.sub = $1
	`, info.Session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer done()

	groupUserScheduleStubs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[types.IGroupUserScheduleStub])
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUserScheduleStubsResponse{GroupUserScheduleStubs: groupUserScheduleStubs}, nil
}

func (h *Handlers) GetGroupUserScheduleStubReplacement(info ReqInfo, data *types.GetGroupUserScheduleStubReplacementRequest) (*types.GetGroupUserScheduleStubReplacementResponse, error) {
	var stubs []*types.IGroupUserScheduleStub

	err := h.Database.QueryRows(info.Ctx, info.Tx, &stubs, `
		SELECT replacement FROM dbfunc_schema.get_peer_schedule_replacement($1::UUID[], $2::DATE, $3::INTERVAL, $4::TEXT)
	`, data.GetUserScheduleId(), data.GetSlotDate(), data.GetStartTime(), data.GetTierName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUserScheduleStubReplacementResponse{GroupUserScheduleStubs: stubs}, nil
}

func (h *Handlers) PatchGroupUserScheduleStubReplacement(info ReqInfo, data *types.PatchGroupUserScheduleStubReplacementRequest) (*types.PatchGroupUserScheduleStubReplacementResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.quotes
		SET slot_date = $2, schedule_bracket_slot_id = $3, service_tier_id = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
	`, data.GetQuoteId(), data.GetSlotDate(), data.GetScheduleBracketSlotId(), data.GetServiceTierId(), info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchGroupUserScheduleStubReplacementResponse{Success: true}, nil
}

func (h *Handlers) DeleteGroupUserScheduleByUserScheduleId(info ReqInfo, data *types.DeleteGroupUserScheduleByUserScheduleIdRequest) (*types.DeleteGroupUserScheduleByUserScheduleIdResponse, error) {
	idsSplit := strings.Split(data.GetIds(), ",")

	for _, userScheduleId := range idsSplit {
		var parts []*types.ScheduledParts

		err := h.Database.QueryRows(info.Ctx, info.Tx, &parts, `SELECT * FROM dbfunc_schema.get_scheduled_parts($1);`, userScheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		hasParts := false
		for _, part := range parts {
			if len(part.GetIds()) > 0 {
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

		var groupScheduleId string

		err = info.Tx.QueryRow(info.Ctx, `
			SELECT group_schedule_id as "groupScheduleId"
			FROM dbtable_schema.group_user_schedules
			WHERE user_schedule_id = $1
		`, userScheduleId).Scan(&groupScheduleId)

		h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/user_schedules/"+groupScheduleId)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/schedules")
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/user_schedules_stubs")

	return &types.DeleteGroupUserScheduleByUserScheduleIdResponse{Success: true}, nil
}
