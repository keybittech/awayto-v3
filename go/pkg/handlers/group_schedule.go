package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupSchedule(w http.ResponseWriter, req *http.Request, data *types.PostGroupScheduleRequest, session *types.UserSession, tx *sql.Tx) (*types.PostGroupScheduleResponse, error) {

	var groupScheduleId string
	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.group_schedules (group_id, schedule_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, session.GroupId, data.ScheduleId, session.GroupSub).Scan(&groupScheduleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")

	return &types.PostGroupScheduleResponse{Id: groupScheduleId}, nil
}

func (h *Handlers) PatchGroupSchedule(w http.ResponseWriter, req *http.Request, data *types.PatchGroupScheduleRequest, session *types.UserSession, tx *sql.Tx) (*types.PatchGroupScheduleResponse, error) {
	scheduleResp, err := h.PatchSchedule(w, req, &types.PatchScheduleRequest{Schedule: data.GetGroupSchedule().GetSchedule()}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules/master/"+data.GetGroupSchedule().GetSchedule().GetId())

	return &types.PatchGroupScheduleResponse{Success: scheduleResp.Success}, nil
}

func (h *Handlers) GetGroupSchedules(w http.ResponseWriter, req *http.Request, data *types.GetGroupSchedulesRequest, session *types.UserSession, tx *sql.Tx) (*types.GetGroupSchedulesResponse, error) {
	var groupSchedules []*types.IGroupSchedule
	err := h.Database.QueryRows(tx, &groupSchedules, `
		SELECT TO_JSONB(es) as schedule, es.name, egs.id, egs."groupId"
		FROM dbview_schema.enabled_schedules es
		JOIN dbview_schema.enabled_group_schedules egs ON egs."scheduleId" = es.id
		WHERE egs."groupId" = $1
	`, session.GroupId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupSchedulesResponse{GroupSchedules: groupSchedules}, nil
}

func (h *Handlers) GetGroupScheduleMasterById(w http.ResponseWriter, req *http.Request, data *types.GetGroupScheduleMasterByIdRequest, session *types.UserSession, tx *sql.Tx) (*types.GetGroupScheduleMasterByIdResponse, error) {
	var groupSchedules []*types.IGroupSchedule

	err := h.Database.QueryRows(tx, &groupSchedules, `
		SELECT TO_JSONB(ese) as schedule, ese.name, true as master, ese.id as "scheduleId"
		FROM dbview_schema.enabled_schedules_ext ese
		WHERE ese.id = $1
	`, data.GetGroupScheduleId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupSchedules) == 0 {
		return nil, util.ErrCheck(errors.New("schedule not found"))
	}

	return &types.GetGroupScheduleMasterByIdResponse{GroupSchedule: groupSchedules[0]}, nil
}

func (h *Handlers) GetGroupScheduleByDate(w http.ResponseWriter, req *http.Request, data *types.GetGroupScheduleByDateRequest, session *types.UserSession, tx *sql.Tx) (*types.GetGroupScheduleByDateResponse, error) {

	var scheduleTimeUnitName string
	err := tx.QueryRow(`
		SELECT tu.name
		FROM dbtable_schema.schedules s
		JOIN dbtable_schema.time_units tu ON tu.id = s.schedule_time_unit_id
		WHERE s.id = $1
	`, data.GroupScheduleId).Scan(&scheduleTimeUnitName)

	var groupScheduleDateSlots []*types.IGroupScheduleDateSlots

	if "week" == scheduleTimeUnitName {
		err = h.Database.QueryRows(tx, &groupScheduleDateSlots, `
			SELECT
        DISTINCT slot."startTime"::INTERVAL,
        slot.id as "scheduleBracketSlotId",
        TO_CHAR(DATE_TRUNC('week', week_start::DATE), 'YYYY-MM-DD')::TEXT as "weekStart",
        TO_CHAR(DATE_TRUNC('week', week_start::DATE) + slot."startTime"::INTERVAL, 'YYYY-MM-DD')::TEXT as "startDate"
      FROM generate_series($1::DATE, $1::DATE + INTERVAL '1 month', INTERVAL '1 week') AS week_start
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
      ORDER BY "startDate", "startTime"
		`, data.Date, data.GroupScheduleId, session.Timezone)
	} else {
		err = h.Database.QueryRows(tx, &groupScheduleDateSlots, `
			SELECT
				DISTINCT slot."startTime"::INTERVAL,
				slot.id as "scheduleBracketSlotId",
				TO_CHAR(cycle_start::DATE, 'YYYY-MM-DD')::TEXT as "weekStart",
				TO_CHAR(cycle_start::DATE + slot."startTime"::INTERVAL, 'YYYY-MM-DD')::TEXT as "startDate"
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
					(DATE_TRUNC('month', $1::DATE) - INTERVAL '14 days') AND
					(DATE_TRUNC('month', $1::DATE) + INTERVAL '45 days')
				AND (cycle_start::DATE + slot."startTime"::INTERVAL) AT TIME ZONE schedule.timezone AT TIME ZONE $3 > (NOW() AT TIME ZONE $3)
			ORDER BY "startDate", "startTime"
		`, data.Date, data.GroupScheduleId, session.Timezone)
	}
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupScheduleByDateResponse{GroupScheduleDateSlots: groupScheduleDateSlots}, nil
}

func (h *Handlers) DeleteGroupSchedule(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupScheduleRequest, session *types.UserSession, tx *sql.Tx) (*types.DeleteGroupScheduleResponse, error) {

	err := h.Database.SetDbVar(tx, "user_sub", session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, groupScheduleTableId := range strings.Split(data.GetGroupScheduleIds(), ",") {
		var groupScheduleId string
		err = tx.QueryRow(`
			DELETE FROM dbtable_schema.group_schedules
			WHERE id = $1
			RETURNING schedule_id
		`, groupScheduleTableId).Scan(&groupScheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		var userScheduleIds sql.NullString
		err = tx.QueryRow(`
			SELECT STRING_AGG(user_schedule_id::TEXT, ',')
			FROM dbtable_schema.group_user_schedules
			WHERE group_schedule_id = $1
		`, groupScheduleId).Scan(&userScheduleIds)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if userScheduleIds.Valid {
			_, err = h.DeleteSchedule(w, req, &types.DeleteScheduleRequest{Ids: userScheduleIds.String}, session, tx)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}

		var userScheduleCount int64
		err = tx.QueryRow(`
			SELECT COUNT(user_schedule_id)
			FROM dbtable_schema.group_user_schedules
			WHERE group_schedule_id = $1
		`, groupScheduleId).Scan(&userScheduleCount)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if userScheduleCount > 0 {
			_, err = tx.Exec(`
				UPDATE dbtable_schema.schedules
				SET enabled = false
				WHERE id = $1
			`, groupScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		} else {
			_, err = tx.Exec(`
			DELETE FROM dbtable_schema.schedules
			WHERE dbtable_schema.schedules.id = $1
		`, groupScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	err = h.Database.SetDbVar(tx, "user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")

	return &types.DeleteGroupScheduleResponse{Success: true}, nil
}
