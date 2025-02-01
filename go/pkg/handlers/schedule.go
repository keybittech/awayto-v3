package handlers

import (
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

func (h *Handlers) PostSchedule(w http.ResponseWriter, req *http.Request, data *types.PostScheduleRequest) (*types.PostScheduleResponse, error) {
	session := h.Redis.ReqSession(req)
	schedule := data.GetSchedule()

	tx, ongoing := h.Database.ReqTx(req)
	if tx == nil {
		return nil, util.ErrCheck(errors.New("bad post schedule tx"))
	}

	if !ongoing {
		defer tx.Rollback()
	}

	userSub := session.UserSub
	if reqSub := req.Context().Value("UserSub"); reqSub != nil {
		userSub = reqSub.(string)
	}

	var scheduleEndTime *string
	if schedule.GetEndTime() != "" {
		scheduleEndTime = &schedule.EndTime
	}

	var scheduleId string

	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.schedules (name, created_sub, slot_duration, schedule_time_unit_id, bracket_time_unit_id, slot_time_unit_id, start_time, end_time, timezone)
		VALUES ($1, $2::uuid, $3::integer, $4::uuid, $5::uuid, $6::uuid, $7, $8, $9)
		RETURNING id
	`, schedule.GetName(), userSub, schedule.GetSlotDuration(), schedule.GetScheduleTimeUnitId(), schedule.GetBracketTimeUnitId(), schedule.GetSlotTimeUnitId(), schedule.GetStartTime(), scheduleEndTime, schedule.GetTimezone()).Scan(&scheduleId)

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Constraint == "unique_enabled_name_created_sub" {
			return nil, util.ErrCheck(errors.New("You can only join a master schedule once. Instead, edit that schedule, then add another bracket to it."))
		}
		return nil, util.ErrCheck(err)
	}

	if !ongoing {
		tx.Commit()
	}

	return &types.PostScheduleResponse{Id: scheduleId}, nil
}

func (h *Handlers) PostScheduleBrackets(w http.ResponseWriter, req *http.Request, data *types.PostScheduleBracketsRequest) (*types.PostScheduleBracketsResponse, error) {
	session := h.Redis.ReqSession(req)

	if err := h.RemoveScheduleBrackets(req.Context(), data.GetScheduleId()); err != nil {
		return nil, util.ErrCheck(err)
	}

	brackets := data.GetBrackets()
	for _, b := range brackets {
		var bracket types.IScheduleBracket

		err := h.Database.Client().QueryRow(`
			INSERT INTO dbtable_schema.schedule_brackets (schedule_id, duration, multiplier, automatic, created_sub)
			VALUES ($1, $2, $3, $4, $5::uuid)
			RETURNING id
		`, data.GetScheduleId(), b.Duration, b.Multiplier, b.Automatic, session.UserSub).Scan(&bracket.Id)

		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, serv := range b.Services {
			_, err = h.Database.Client().Exec(`
				INSERT INTO dbtable_schema.schedule_bracket_services (schedule_bracket_id, service_id, created_sub)
				VALUES ($1, $2, $3::uuid)
			`, bracket.Id, serv.Id, session.UserSub)

			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}

		newSlots := make(map[string]*types.IScheduleBracketSlot)

		_, err = h.Database.Client().Exec(`
			DELETE FROM dbtable_schema.schedule_bracket_slots
			WHERE schedule_bracket_id = $1
		`, bracket.Id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, slot := range b.Slots {
			var slotId string

			err = h.Database.Client().QueryRow(`
				INSERT INTO dbtable_schema.schedule_bracket_slots (schedule_bracket_id, start_time, created_sub)
				VALUES ($1, $2::interval, $3::uuid)
				RETURNING id
			`, bracket.Id, slot.StartTime, session.UserSub).Scan(&slotId)
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

func (h *Handlers) PatchSchedule(w http.ResponseWriter, req *http.Request, data *types.PatchScheduleRequest) (*types.PatchScheduleResponse, error) {
	session := h.Redis.ReqSession(req)
	schedule := data.GetSchedule()
	// TODO add a what to expect page

	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.schedules
		SET name = $2, start_time = $3, end_time = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
	`, schedule.GetId(), schedule.GetName(), schedule.GetStartTime(), schedule.GetEndTime(), session.UserSub, time.Now().Local().UTC())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchScheduleResponse{Success: true}, nil
}

func (h *Handlers) GetSchedules(w http.ResponseWriter, req *http.Request, data *types.GetSchedulesRequest) (*types.GetSchedulesResponse, error) {
	session := h.Redis.ReqSession(req)
	var schedules []*types.ISchedule

	err := h.Database.QueryRows(&schedules, `
		SELECT es.* 
		FROM dbview_schema.enabled_schedules es
		JOIN dbtable_schema.schedules s ON s.id = es.id
		WHERE s.created_sub = $1
	`, session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetSchedulesResponse{Schedules: schedules}, nil
}

func (h *Handlers) GetScheduleById(w http.ResponseWriter, req *http.Request, data *types.GetScheduleByIdRequest) (*types.GetScheduleByIdResponse, error) {
	var schedules []*types.ISchedule

	err := h.Database.QueryRows(&schedules, `
		SELECT * FROM dbview_schema.enabled_schedules_ext
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(schedules) == 0 {
		err := errors.New("no schedules found for id " + data.GetId())
		return nil, util.ErrCheck(err)
	}

	return &types.GetScheduleByIdResponse{Schedule: schedules[0]}, nil
}

func (h *Handlers) DeleteSchedule(w http.ResponseWriter, req *http.Request, data *types.DeleteScheduleRequest) (*types.DeleteScheduleResponse, error) {
	session := h.Redis.ReqSession(req)

	for _, scheduleId := range strings.Split(data.GetIds(), ",") {

		if err := h.RemoveScheduleBrackets(req.Context(), scheduleId); err != nil {
			return nil, err
		}

		_, err := h.Database.Client().Exec(`
			UPDATE dbtable_schema.schedules
			SET enabled = false
			WHERE id = $1
		`, scheduleId)

		_, err = h.Database.Client().Exec(`
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

func (h *Handlers) DisableSchedule(w http.ResponseWriter, req *http.Request, data *types.DisableScheduleRequest) (*types.DisableScheduleResponse, error) {
	session := h.Redis.ReqSession(req)
	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.schedules
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableScheduleResponse{Success: true}, nil
}

func (h *Handlers) RemoveScheduleBrackets(ctx context.Context, scheduleId string) error {
	var parts []*types.ScheduledParts

	err := h.Database.QueryRows(&parts, `
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

	err = h.Database.Client().QueryRow(`
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
				h.Database.Client().Exec(`
					DELETE FROM dbtable_schema.schedule_bracket_slots
					WHERE schedule_bracket_id = $1 AND id <> ALL($2::uuid[])
				`, bracketId, scheduledSlots.GetIds())

				h.Database.Client().Exec(`
					UPDATE dbtable_schema.schedule_bracket_slots
					SET enabled = false
					WHERE schedule_bracket_id = $1 AND id = ANY($2::uuid[])
				`, bracketId, scheduledSlots.GetIds())
			}

			if hasServices {
				h.Database.Client().Exec(`
					DELETE FROM dbtable_schema.schedule_bracket_services
					WHERE schedule_bracket_id = $1 AND id <> ALL($2::uuid[])
				`, bracketId, scheduledServices.GetIds())

				h.Database.Client().Exec(`
					UPDATE dbtable_schema.schedule_bracket_services
					SET enabled = false
					WHERE schedule_bracket_id = $1 AND id = ANY($2::uuid[])
				`, bracketId, scheduledServices.GetIds())
			}

			h.Database.Client().Exec(`
				DELETE FROM dbtable_schema.schedule_brackets
				USING dbtable_schema.schedule_bracket_slots slot,
							dbtable_schema.schedule_bracket_services service
				WHERE dbtable_schema.schedule_brackets.id = $1
				AND dbtable_schema.schedule_brackets.id = slot.schedule_bracket_id
				AND slot.schedule_bracket_id = service.schedule_bracket_id
				AND slot.id <> ALL($2::uuid[])
				AND service.id <> ALL($3::uuid[])
			`, bracketId, scheduledSlots.GetIds(), scheduledServices.GetIds())

			h.Database.Client().Exec(`
				UPDATE dbtable_schema.schedule_brackets
				SET enabled = false
				FROM dbtable_schema.schedule_bracket_slots slot
				JOIN dbtable_schema.schedule_bracket_services service ON service.schedule_bracket_id = slot.schedule_bracket_id
				WHERE dbtable_schema.schedule_brackets.id = $1
				AND slot.schedule_bracket_id = dbtable_schema.schedule_brackets.id
				AND (slot.id = ANY($2::uuid[])
				OR service.id = ANY($3::uuid[]))
			`, bracketId, scheduledSlots.GetIds(), scheduledServices.GetIds())
		} else {
			h.Database.Client().Exec(`
				DELETE FROM dbtable_schema.schedule_brackets
				WHERE id = $1
			`, bracketId)
		}
	}

	return nil
}
