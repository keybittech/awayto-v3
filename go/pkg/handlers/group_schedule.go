package handlers

import (
	"database/sql"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func (h *Handlers) PostGroupSchedule(info ReqInfo, data *types.PostGroupScheduleRequest) (*types.PostGroupScheduleResponse, error) {

	var groupScheduleId string
	err := info.Tx.QueryRow(info.Req.Context(), `
		INSERT INTO dbtable_schema.group_schedules (group_id, schedule_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, info.Session.GroupId, data.ScheduleId, info.Session.GroupSub).Scan(&groupScheduleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/schedules")

	return &types.PostGroupScheduleResponse{Id: groupScheduleId}, nil
}

func (h *Handlers) PatchGroupSchedule(info ReqInfo, data *types.PatchGroupScheduleRequest) (*types.PatchGroupScheduleResponse, error) {
	scheduleResp, err := h.PatchSchedule(info, &types.PatchScheduleRequest{Schedule: data.GetGroupSchedule().GetSchedule()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/schedules")
	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/schedules/master/"+data.GetGroupSchedule().GetSchedule().GetId())

	return &types.PatchGroupScheduleResponse{Success: scheduleResp.Success}, nil
}

func (h *Handlers) GetGroupSchedules(info ReqInfo, data *types.GetGroupSchedulesRequest) (*types.GetGroupSchedulesResponse, error) {
	var groupSchedules []*types.IGroupSchedule
	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &groupSchedules, `
		SELECT TO_JSONB(es) as schedule, es.name, egs.id, egs."groupId"
		FROM dbview_schema.enabled_schedules es
		JOIN dbview_schema.enabled_group_schedules egs ON egs."scheduleId" = es.id
		WHERE egs."groupId" = $1
	`, info.Session.GroupId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupSchedulesResponse{GroupSchedules: groupSchedules}, nil
}

func (h *Handlers) GetGroupScheduleMasterById(info ReqInfo, data *types.GetGroupScheduleMasterByIdRequest) (*types.GetGroupScheduleMasterByIdResponse, error) {
	groupSchedule := &types.IGroupSchedule{}
	var scheduleBytes []byte
	err := info.Tx.QueryRow(info.Req.Context(), `
		SELECT TO_JSONB(ese), ese.name
		FROM dbview_schema.enabled_schedules_ext ese
		WHERE ese.id = $1
	`, data.GetGroupScheduleId()).Scan(&scheduleBytes, &groupSchedule.Name)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	schedule := &types.ISchedule{}
	err = protojson.Unmarshal(scheduleBytes, schedule)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	groupSchedule.Master = true
	groupSchedule.ScheduleId = schedule.Id
	groupSchedule.Schedule = schedule

	return &types.GetGroupScheduleMasterByIdResponse{GroupSchedule: groupSchedule}, nil
}

func (h *Handlers) GetGroupScheduleByDate(info ReqInfo, data *types.GetGroupScheduleByDateRequest) (*types.GetGroupScheduleByDateResponse, error) {

	var scheduleTimeUnitName string
	err := info.Tx.QueryRow(info.Req.Context(), `
		SELECT tu.name
		FROM dbtable_schema.schedules s
		JOIN dbtable_schema.time_units tu ON tu.id = s.schedule_time_unit_id
		WHERE s.id = $1
	`, data.GroupScheduleId).Scan(&scheduleTimeUnitName)

	var rows pgx.Rows
	if "week" == scheduleTimeUnitName {
		rows, err = info.Tx.Query(info.Req.Context(), `
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
		`, data.Date, data.GroupScheduleId, info.Session.Timezone)
	} else {
		rows, err = info.Tx.Query(info.Req.Context(), `
			WITH times AS (
				SELECT
					DISTINCT slot."startTime",
					slot.id as "scheduleBracketSlotId",
					TO_CHAR(cycle_start::DATE, 'YYYY-MM-DD')::TEXT as "weekStart",
					TO_CHAR(cycle_start::DATE + slot."startTime"::INTERVAL, 'YYYY-MM-DD')::TEXT as "startDate",
					cycle_start::DATE + slot."startTime"::INTERVAL as real_time
				FROM (
					WITH schedule_info AS (
						SELECT start_time FROM dbtable_schema.schedules	WHERE id = $2::uuid
					)
					SELECT 
						si.start_time + (INTERVAL '28 days' * n) as cycle_start
					FROM schedule_info si,
					generate_series(
						FLOOR(EXTRACT(EPOCH FROM (DATE_TRUNC('month', $1::DATE) - INTERVAL '28 days' - si.start_time)) / EXTRACT(EPOCH FROM INTERVAL '28 days'))::INTEGER,
						CEIL(EXTRACT(EPOCH FROM (DATE_TRUNC('month', $1::DATE) + INTERVAL '2 months' - si.start_time)) / EXTRACT(EPOCH FROM INTERVAL '28 days'))::INTEGER,
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
		`, data.Date, data.GroupScheduleId, info.Session.Timezone)
	}
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

	groupPoolTx, _, err := h.Database.DatabaseClient.OpenPoolSessionGroupTx(info.Req.Context(), info.Session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer groupPoolTx.Rollback(info.Req.Context())

	for _, groupScheduleTableId := range strings.Split(data.GetGroupScheduleIds(), ",") {
		var groupScheduleId string
		err = groupPoolTx.QueryRow(info.Req.Context(), `
			DELETE FROM dbtable_schema.group_schedules
			WHERE id = $1
			RETURNING schedule_id
		`, groupScheduleTableId).Scan(&groupScheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		var userScheduleIds sql.NullString
		err = groupPoolTx.QueryRow(info.Req.Context(), `
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
		err = groupPoolTx.QueryRow(info.Req.Context(), `
			SELECT COUNT(user_schedule_id)
			FROM dbtable_schema.group_user_schedules
			WHERE group_schedule_id = $1
		`, groupScheduleId).Scan(&userScheduleCount)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if userScheduleCount > 0 {
			_, err = groupPoolTx.Exec(info.Req.Context(), `
				UPDATE dbtable_schema.schedules
				SET enabled = false
				WHERE id = $1
			`, groupScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		} else {
			_, err = groupPoolTx.Exec(info.Req.Context(), `
			DELETE FROM dbtable_schema.schedules
			WHERE dbtable_schema.schedules.id = $1
		`, groupScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	err = h.Database.DatabaseClient.ClosePoolSessionTx(info.Req.Context(), groupPoolTx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"schedules")
	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/schedules")

	return &types.DeleteGroupScheduleResponse{Success: true}, nil
}
