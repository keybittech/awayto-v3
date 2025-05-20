package handlers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupSchedule(info ReqInfo, data *types.PostGroupScheduleRequest) (*types.PostGroupScheduleResponse, error) {
	var groupScheduleId string
	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_schedules (group_id, schedule_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, info.Session.GetGroupId(), data.ScheduleId, info.Session.GetGroupSub()).Scan(&groupScheduleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/schedules")

	return &types.PostGroupScheduleResponse{Id: groupScheduleId}, nil
}

func (h *Handlers) PatchGroupSchedule(info ReqInfo, data *types.PatchGroupScheduleRequest) (*types.PatchGroupScheduleResponse, error) {
	scheduleResp, err := h.PatchSchedule(info, &types.PatchScheduleRequest{Schedule: data.GetGroupSchedule().GetSchedule()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/schedules")
	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/schedules/master/"+data.GetGroupSchedule().GetSchedule().GetId())

	return &types.PatchGroupScheduleResponse{Success: scheduleResp.Success}, nil
}

func (h *Handlers) GetGroupSchedules(info ReqInfo, data *types.GetGroupSchedulesRequest) (*types.GetGroupSchedulesResponse, error) {
	groupSchedules := util.BatchQuery[types.IGroupSchedule](info.Batch, `
		SELECT TO_JSONB(es) as schedule, es.name, egs.id, egs."groupId"
		FROM dbview_schema.enabled_schedules es
		JOIN dbview_schema.enabled_group_schedules egs ON egs."scheduleId" = es.id
		WHERE egs."groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupSchedulesResponse{GroupSchedules: *groupSchedules}, nil
}

func (h *Handlers) GetGroupScheduleMasterById(info ReqInfo, data *types.GetGroupScheduleMasterByIdRequest) (*types.GetGroupScheduleMasterByIdResponse, error) {
	// The schedule master is the root ISchedule, not an IGroupSchedule
	schedule := util.BatchQueryRow[types.ISchedule](info.Batch, `
		SELECT id, name, timezone, "startDate", "endDate", "scheduleTimeUnitId", "bracketTimeUnitId", "slotTimeUnitId", "slotDuration", "createdOn", brackets
		FROM dbview_schema.enabled_schedules_ext
		WHERE id = $1
	`, data.GroupScheduleId)

	info.Batch.Send(info.Ctx)

	s := *schedule

	info.Batch.Reset()

	return &types.GetGroupScheduleMasterByIdResponse{GroupSchedule: &types.IGroupSchedule{Master: true, ScheduleId: s.Id, Schedule: s}}, nil
}

func (h *Handlers) GetGroupScheduleByDate(info ReqInfo, data *types.GetGroupScheduleByDateRequest) (*types.GetGroupScheduleByDateResponse, error) {
	println(fmt.Sprint(info.Session.GetProtoClone()))
	var scheduleTimeUnitName string

	err := info.Tx.QueryRow(info.Ctx, `
		SELECT tu.name
		FROM dbtable_schema.schedules s
		JOIN dbtable_schema.time_units tu ON tu.id = s.schedule_time_unit_id
		WHERE s.id = $1
	`, data.GroupScheduleId).Scan(&scheduleTimeUnitName)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var query string
	if "week" == scheduleTimeUnitName {
		query = `
			WITH times AS (
				SELECT
					DISTINCT slot."startTime",
					slot.id as "scheduleBracketSlotId",
					TO_CHAR(DATE_TRUNC('week', week_start::DATE), 'YYYY-MM-DD')::TEXT as "weekStart",
					TO_CHAR(DATE_TRUNC('week', week_start::DATE) + slot."startTime"::INTERVAL, 'YYYY-MM-DD')::TEXT as "startDate",
					DATE_TRUNC('week', week_start::DATE) + slot."startTime"::INTERVAL as real_time
				FROM generate_series($1::DATE, $1::DATE + INTERVAL '5 weeks', INTERVAL '1 week') AS week_start
				CROSS JOIN dbview_schema.enabled_schedule_bracket_slots slot
				LEFT JOIN dbtable_schema.bookings booking ON booking.schedule_bracket_slot_id = slot.id
					AND DATE_TRUNC('week', booking.slot_date) = DATE_TRUNC('week', week_start)
				LEFT JOIN dbtable_schema.schedule_bracket_slot_exclusions exclusion ON exclusion.schedule_bracket_slot_id = slot.id 
					AND DATE_TRUNC('week', exclusion.exclusion_date) = DATE_TRUNC('week', week_start)
				JOIN dbtable_schema.schedule_brackets bracket ON bracket.id = slot."scheduleBracketId"
				JOIN dbtable_schema.group_user_schedules gus ON gus.user_schedule_id = bracket.schedule_id
				JOIN dbtable_schema.schedules schedule ON schedule.id = gus.group_schedule_id
				WHERE
					booking.id IS NULL
					AND exclusion.id IS NULL
					AND schedule.id = $2::uuid
					AND (DATE_TRUNC('week', week_start::DATE) + slot."startTime"::INTERVAL) AT TIME ZONE schedule.timezone AT TIME ZONE $3 > (NOW() AT TIME ZONE $3)
				ORDER BY real_time
			)
			SELECT "startTime", "scheduleBracketSlotId", "weekStart", "startDate"
			FROM times
		`
	} else {
		query = ` 
			WITH times AS (
				SELECT
					DISTINCT slot."startTime",
					slot.id as "scheduleBracketSlotId",
					TO_CHAR(cycle_start::DATE, 'YYYY-MM-DD')::TEXT as "weekStart",
					TO_CHAR(cycle_start::DATE + slot."startTime"::INTERVAL, 'YYYY-MM-DD')::TEXT as "startDate",
					cycle_start::DATE + slot."startTime"::INTERVAL as real_time
				FROM (
					WITH schedule_info AS (
						SELECT start_date FROM dbtable_schema.schedules WHERE id = $2::uuid
					)
					SELECT 
						si.start_date + (INTERVAL '28 days' * n) as cycle_start
					FROM schedule_info si,
					generate_series(
						FLOOR(EXTRACT(EPOCH FROM (DATE_TRUNC('month', $1::DATE) - INTERVAL '28 days' - si.start_date)) / EXTRACT(EPOCH FROM INTERVAL '28 days'))::INTEGER,
						CEIL(EXTRACT(EPOCH FROM (DATE_TRUNC('month', $1::DATE) + INTERVAL '2 months' - si.start_date)) / EXTRACT(EPOCH FROM INTERVAL '28 days'))::INTEGER,
						1
					) as n
				) cycle_dates
				CROSS JOIN dbview_schema.enabled_schedule_bracket_slots slot
				LEFT JOIN dbtable_schema.bookings booking ON booking.schedule_bracket_slot_id = slot.id
					AND DATE_TRUNC('day', booking.slot_date) = DATE_TRUNC('day', cycle_start + slot."startTime"::INTERVAL)
				LEFT JOIN dbtable_schema.schedule_bracket_slot_exclusions exclusion ON exclusion.schedule_bracket_slot_id = slot.id 
					AND DATE_TRUNC('day', exclusion.exclusion_date) = DATE_TRUNC('day', cycle_start + slot."startTime"::INTERVAL)
				JOIN dbtable_schema.schedule_brackets bracket ON bracket.id = slot."scheduleBracketId"
				JOIN dbtable_schema.group_user_schedules gus ON gus.user_schedule_id = bracket.schedule_id
				JOIN dbtable_schema.schedules schedule ON schedule.id = gus.group_schedule_id
				WHERE
					booking.id IS NULL
					AND exclusion.id IS NULL
					AND schedule.id = $2::uuid
					AND (cycle_start::DATE + slot."startTime"::INTERVAL) BETWEEN 
						(DATE_TRUNC('month', $1::DATE) - INTERVAL '14 days') AND (DATE_TRUNC('month', $1::DATE) + INTERVAL '45 days')
					AND (cycle_start::DATE + slot."startTime"::INTERVAL) AT TIME ZONE schedule.timezone AT TIME ZONE $3 > (NOW() AT TIME ZONE $3)
				ORDER BY real_time
			)
			SELECT "startTime", "scheduleBracketSlotId", "weekStart", "startDate"
			FROM times
		`
	}

	rows, err := info.Tx.Query(info.Ctx, query, data.Date, data.GroupScheduleId, info.Session.GetTimezone())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	groupScheduleDateSlots, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[types.IGroupScheduleDateSlots])
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupScheduleByDateResponse{GroupScheduleDateSlots: groupScheduleDateSlots}, nil
}

func (h *Handlers) DeleteGroupSchedule(info ReqInfo, data *types.DeleteGroupScheduleRequest) (*types.DeleteGroupScheduleResponse, error) {
	groupPoolTx, _, err := h.Database.DatabaseClient.OpenPoolSessionGroupTx(info.Ctx, info.Session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer groupPoolTx.Rollback(info.Ctx)

	for groupScheduleTableId := range strings.SplitSeq(data.GetGroupScheduleIds(), ",") {
		var groupScheduleId string
		err = groupPoolTx.QueryRow(info.Ctx, `
			DELETE FROM dbtable_schema.group_schedules
			WHERE id = $1
			RETURNING schedule_id
		`, groupScheduleTableId).Scan(&groupScheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		var userScheduleIds sql.NullString
		err = groupPoolTx.QueryRow(info.Ctx, `
			SELECT STRING_AGG(user_schedule_id::TEXT, ',')
			FROM dbtable_schema.group_user_schedules
			WHERE group_schedule_id = $1
		`, groupScheduleId).Scan(&userScheduleIds)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if userScheduleIds.Valid {
			_, err = h.DeleteSchedule(info, &types.DeleteScheduleRequest{Ids: userScheduleIds.String})
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}

		var userScheduleCount int64
		err = groupPoolTx.QueryRow(info.Ctx, `
			SELECT COUNT(user_schedule_id)
			FROM dbtable_schema.group_user_schedules
			WHERE group_schedule_id = $1
		`, groupScheduleId).Scan(&userScheduleCount)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if userScheduleCount > 0 {
			_, err = groupPoolTx.Exec(info.Ctx, `
				UPDATE dbtable_schema.schedules
				SET enabled = false
				WHERE id = $1
			`, groupScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		} else {
			_, err = groupPoolTx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.schedules
			WHERE dbtable_schema.schedules.id = $1
		`, groupScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	err = h.Database.DatabaseClient.ClosePoolSessionTx(info.Ctx, groupPoolTx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"schedules")
	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/schedules")

	return &types.DeleteGroupScheduleResponse{Success: true}, nil
}
