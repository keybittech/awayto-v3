package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (h *Handlers) PostSchedule(w http.ResponseWriter, req *http.Request, data *types.PostScheduleRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostScheduleResponse, error) {
	var scheduleId string

	var startTime, endTime *string
	if data.StartTime != nil {
		st := data.StartTime.AsTime().Format(time.RFC3339Nano)
		startTime = &st
	}
	if data.EndTime != nil {
		et := data.EndTime.AsTime().Format(time.RFC3339Nano)
		endTime = &et
	}

	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.schedules (name, created_sub, slot_duration, schedule_time_unit_id, bracket_time_unit_id, slot_time_unit_id, start_time, end_time, timezone)
		VALUES ($1, $2::uuid, $3::integer, $4::uuid, $5::uuid, $6::uuid, $7, $8, $9)
		RETURNING id
	`, data.GetName(), session.UserSub, data.GetSlotDuration(), data.GetScheduleTimeUnitId(), data.GetBracketTimeUnitId(), data.GetSlotTimeUnitId(), &startTime, &endTime, session.Timezone).Scan(&scheduleId)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Constraint == "unique_enabled_name_created_sub" {
			return nil, util.ErrCheck(util.UserError("You can only join a master schedule once. Instead, edit that schedule, then add another bracket to it."))
		}
		return nil, util.ErrCheck(err)
	}

	return &types.PostScheduleResponse{Id: scheduleId}, nil
}

func (h *Handlers) PostScheduleBrackets(w http.ResponseWriter, req *http.Request, data *types.PostScheduleBracketsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostScheduleBracketsResponse, error) {
	if err := h.RemoveScheduleBrackets(req.Context(), data.GetScheduleId(), tx); err != nil {
		return nil, util.ErrCheck(err)
	}

	brackets := data.GetBrackets()
	for _, b := range brackets {
		var bracket types.IScheduleBracket

		err := tx.QueryRow(`
			INSERT INTO dbtable_schema.schedule_brackets (schedule_id, duration, multiplier, automatic, created_sub, group_id)
			VALUES ($1, $2, $3, $4, $5::uuid, $6)
			RETURNING id
		`, data.GetScheduleId(), b.Duration, b.Multiplier, b.Automatic, session.UserSub, session.GroupId).Scan(&bracket.Id)

		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, servId := range data.ServiceIds {
			_, err = tx.Exec(`
				INSERT INTO dbtable_schema.schedule_bracket_services (schedule_bracket_id, service_id, created_sub, group_id)
				VALUES ($1, $2, $3::uuid, $4::uuid)
			`, bracket.Id, servId, session.UserSub, session.GroupId)

			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}

		newSlots := make(map[string]*types.IScheduleBracketSlot)

		_, err = tx.Exec(`
			DELETE FROM dbtable_schema.schedule_bracket_slots
			WHERE schedule_bracket_id = $1
		`, bracket.Id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, slot := range b.Slots {
			var slotId string

			err = tx.QueryRow(`
				INSERT INTO dbtable_schema.schedule_bracket_slots (schedule_bracket_id, start_time, created_sub, group_id)
				VALUES ($1, $2::interval, $3::uuid, $4::uuid)
				RETURNING id
			`, bracket.Id, slot.StartTime, session.UserSub, session.GroupId).Scan(&slotId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			slot.Id = slotId
			slot.ScheduleBracketId = bracket.Id
			newSlots[slotId] = slot
		}

		b.Slots = newSlots
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules/"+data.GetScheduleId())
	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules")

	return &types.PostScheduleBracketsResponse{Id: data.GetScheduleId(), Brackets: brackets}, nil
}

func (h *Handlers) PatchSchedule(w http.ResponseWriter, req *http.Request, data *types.PatchScheduleRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchScheduleResponse, error) {
	schedule := data.GetSchedule()
	// TODO add a what to expect page

	var startTime, endTime *string
	if schedule.StartTime != nil {
		st := schedule.StartTime.AsTime().Format(time.RFC3339Nano)
		startTime = &st
	}
	if schedule.EndTime != nil {
		et := schedule.EndTime.AsTime().Format(time.RFC3339Nano)
		endTime = &et
	}

	_, err := tx.Exec(`
		UPDATE dbtable_schema.schedules
		SET name = $2, start_time = $3, end_time = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
		`, schedule.GetId(), schedule.GetName(), &startTime, &endTime, session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchScheduleResponse{Success: true}, nil
}

func (h *Handlers) GetSchedules(w http.ResponseWriter, req *http.Request, data *types.GetSchedulesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetSchedulesResponse, error) {
	var schedules []*types.ISchedule

	err := tx.QueryRows(&schedules, `SELECT es.* 
		FROM dbview_schema.enabled_schedules es
		JOIN dbtable_schema.schedules s ON s.id = es.id
		WHERE s.created_sub = $1`, session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetSchedulesResponse{Schedules: schedules}, nil
}

func (h *Handlers) GetScheduleById(w http.ResponseWriter, req *http.Request, data *types.GetScheduleByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetScheduleByIdResponse, error) {
	var schedules []*types.ISchedule

	err := tx.QueryRows(&schedules, `
		SELECT * FROM dbview_schema.enabled_schedules_ext
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(schedules) == 0 {
		return nil, util.ErrCheck(util.UserError("No schedules found."))
	}

	return &types.GetScheduleByIdResponse{Schedule: schedules[0]}, nil
}

func (h *Handlers) DeleteSchedule(w http.ResponseWriter, req *http.Request, data *types.DeleteScheduleRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteScheduleResponse, error) {
	for _, scheduleId := range strings.Split(data.GetIds(), ",") {

		if err := h.RemoveScheduleBrackets(req.Context(), scheduleId, tx); err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err := tx.Exec(`
			UPDATE dbtable_schema.schedules
			SET enabled = false
			WHERE id = $1
		`, scheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = tx.Exec(`
			DELETE FROM dbtable_schema.schedules
			WHERE dbtable_schema.schedules.id = $1
			AND NOT EXISTS (
				SELECT 1
				FROM dbtable_schema.schedule_brackets bracket
				WHERE bracket.schedule_id = dbtable_schema.schedules.id
			)
		`, scheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules")
	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	return &types.DeleteScheduleResponse{Success: true}, nil
}

func (h *Handlers) DisableSchedule(w http.ResponseWriter, req *http.Request, data *types.DisableScheduleRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableScheduleResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.schedules
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableScheduleResponse{Success: true}, nil
}

func (h *Handlers) RemoveScheduleBrackets(ctx context.Context, scheduleId string, tx clients.IDatabaseTx) error {
	var parts []*types.ScheduledParts

	err := tx.QueryRows(&parts, `
    SELECT * FROM dbfunc_schema.get_scheduled_parts($1);
  `, scheduleId)
	if err != nil {
		return util.ErrCheck(err)
	}

	scheduledSlots := &types.ScheduledParts{}
	hasSlots := false
	scheduledServices := &types.ScheduledParts{}
	hasServices := false

	for _, p := range parts {
		if len(p.GetIds()) > 0 {
			if p.GetParttype() == "slot" {
				scheduledSlots = p
				hasSlots = true
			} else {
				scheduledServices = p
				hasServices = true
			}
		}
	}

	var idsString sql.NullString

	err = tx.QueryRow(`
    SELECT JSONB_AGG(sb.id)
		FROM dbtable_schema.schedule_brackets sb
    WHERE sb.schedule_id = $1
  `, scheduleId).Scan(&idsString)
	if err != nil {
		return util.ErrCheck(err)
	}

	var ids []string
	if idsString.Valid {
		err := json.Unmarshal([]byte(idsString.String), &ids)
		if err != nil {
			return util.ErrCheck(err)
		}
	}

	for _, bracketId := range ids {

		if hasSlots || hasServices {

			if hasSlots {
				_, err = tx.Exec(`
					DELETE FROM dbtable_schema.schedule_bracket_slots
					WHERE schedule_bracket_id = $1 AND id <> ALL($2::uuid[])
				`, bracketId, pq.Array(scheduledSlots.GetIds()))
				if err != nil {
					return util.ErrCheck(err)
				}
				_, err = tx.Exec(`
					UPDATE dbtable_schema.schedule_bracket_slots
					SET enabled = false
					WHERE schedule_bracket_id = $1 AND id = ANY($2::uuid[])
				`, bracketId, scheduledSlots.GetIds())
				if err != nil {
					return util.ErrCheck(err)
				}
			}

			if hasServices {
				_, err = tx.Exec(`
					DELETE FROM dbtable_schema.schedule_bracket_services
					WHERE schedule_bracket_id = $1 AND id <> ALL($2::uuid[])
				`, bracketId, scheduledServices.GetIds())
				if err != nil {
					return util.ErrCheck(err)
				}

				_, err = tx.Exec(`
					UPDATE dbtable_schema.schedule_bracket_services
					SET enabled = false
					WHERE schedule_bracket_id = $1 AND id = ANY($2::uuid[])
				`, bracketId, scheduledServices.GetIds())
				if err != nil {
					return util.ErrCheck(err)
				}
			}

			_, err = tx.Exec(`
				DELETE FROM dbtable_schema.schedule_brackets
				USING dbtable_schema.schedule_bracket_slots slot,
							dbtable_schema.schedule_bracket_services service
				WHERE dbtable_schema.schedule_brackets.id = $1
				AND dbtable_schema.schedule_brackets.id = slot.schedule_bracket_id
				AND slot.schedule_bracket_id = service.schedule_bracket_id
				AND slot.id <> ALL($2::uuid[])
				AND service.id <> ALL($3::uuid[])
			`, bracketId, scheduledSlots.GetIds(), scheduledServices.GetIds())
			if err != nil {
				return util.ErrCheck(err)
			}

			_, err = tx.Exec(`
				UPDATE dbtable_schema.schedule_brackets
				SET enabled = false
				FROM dbtable_schema.schedule_bracket_slots slot
				JOIN dbtable_schema.schedule_bracket_services service ON service.schedule_bracket_id = slot.schedule_bracket_id
				WHERE dbtable_schema.schedule_brackets.id = $1
				AND slot.schedule_bracket_id = dbtable_schema.schedule_brackets.id
				AND (slot.id = ANY($2::uuid[])
				OR service.id = ANY($3::uuid[]))
			`, bracketId, scheduledSlots.GetIds(), scheduledServices.GetIds())
			if err != nil {
				return util.ErrCheck(err)
			}
		} else {
			_, err = tx.Exec(`
				DELETE FROM dbtable_schema.schedule_brackets
				WHERE id = $1
			`, bracketId)
			if err != nil {
				return util.ErrCheck(err)
			}
		}
	}

	return nil
}
