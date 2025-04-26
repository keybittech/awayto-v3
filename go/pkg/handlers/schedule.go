package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/lib/pq"
)

// PostSchedule handles group schedules (via PostGroupSchedule) and user schedules.
// Group schedules are owned by the group db user sub, while user schedules are owned by the user.
// User schedules must have a set of "brackets", which keep track of the available times on the schedule.
// Brackets also allow for supporting different services at different times, or where cost is available
// brackets can have multipliers so that the cost of services is different based on the bracket:
// for example standard cost from 9am-5pm, then an evening bracket where the cost is 1.5x

// When schedules are created, all records are normally inserted as usual. But when modifications
// are made to existing schedules/brackets, special attention must be paid to schedule/bracket/slot records
// which have been associated with quote requests. If a user goes to request service (create a quote) for
// 9:30AM on Tuesday Mar 11, 2025, this time is specifically recorded on the quote record using the MM-DD-YYYY
// slot_date 03-11-2025, and a schedule_bracket_slot_id pointing to an ISO 8601 duration like P1DT9H30M.
// The durations stored in the DB are always relative to the week start of the schedule in question, so it's
// expected that when displaying this quote's scheduled time on the front end, that dayjs or something else
// is used to figure out the week start for 03-11-2025, and then add P1DT9H30M, in this case it would
// work out to be 03-10-2025 (monday) + P1DT9H30M, resulting in 9:30AM on Tuesday Mar 11th, 2025. A limitation
// for now is that week start uses Monday for the 0 day, which is a limitation needing consideration when
// dealing with various locales. Schedules are also attached to a specific timezone for converting time on the client.

// Therefore when modifying schedules/brackets, any related quotes must be identified and handled such that if
// anything which would cause a schedule_bracket_slot to be removed, then it instead must be disabled, as it
// related to an existing quote. This comes into play when deleting both brackets or slots. Any records disabled
// in the modification process and which are still in the future will then appear in the user interface for
// group admins to see so that they may perform manual rescheduling of the slot. This allows staff to create
// schedules, users to request services from those schedules, while allowing for situations like the staff
// member is no longer part of the group, or their schedule changes at some point, etc., existing upcoming
// appointments will be gracefully disabled so that they may be handled by admins. These records needing
// handling are referred to elsewhere as user_schedule_stubs and represent disabled schedule_bracket_slots.
// For these reasons, the bracket modification process is spread across different focused functions, to help
// ease the task of debugging and general understanding.

func (h *Handlers) PostSchedule(w http.ResponseWriter, req *http.Request, data *types.PostScheduleRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PostScheduleResponse, error) {
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

	var insertScheduleQuery = `
		INSERT INTO dbtable_schema.schedules (name, created_sub, slot_duration, schedule_time_unit_id, bracket_time_unit_id, slot_time_unit_id, start_time, end_time, timezone)
		VALUES ($1, $2::uuid, $3::integer, $4::uuid, $5::uuid, $6::uuid, $7, $8, $9)
		RETURNING id
	`
	var insertScheduleParams = []any{data.GetName(), session.UserSub, data.GetSlotDuration(), data.GetScheduleTimeUnitId(), data.GetBracketTimeUnitId(), data.GetSlotTimeUnitId(), &startTime, &endTime, session.Timezone}

	var err error
	var row pgx.Row
	var done func()

	if data.AsGroup {
		// check to see APP_GROUP_SCHEDULE role is on session.Roles
		if session.RoleBits&int32(types.SiteRoles_APP_GROUP_SCHEDULES) == 0 {
			return nil, util.ErrCheck(errors.New("does not have APP_GROUP_SCHEDULES role"))
		}

		ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, session)
		insertScheduleParams[1] = session.GroupSub
		row, done, err = ds.SessionBatchQueryRow(req.Context(), insertScheduleQuery, insertScheduleParams...)
	} else {
		row = tx.QueryRow(req.Context(), insertScheduleQuery, insertScheduleParams...)
	}
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Constraint == "unique_enabled_name_created_sub" {
			return nil, util.ErrCheck(util.UserError("You can only join a master schedule once. Instead, edit that schedule, then add another bracket to it."))
		}
		return nil, util.ErrCheck(err)
	}
	if done != nil {
		defer done()
	}

	err = row.Scan(&scheduleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Runs when users post their brackets to an existing master schedule
	if len(data.Brackets) > 0 && data.GroupScheduleId != "" {
		err := h.InsertNewBrackets(req.Context(), scheduleId, data.Brackets, tx, session)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = h.PostGroupUserSchedule(w, req, &types.PostGroupUserScheduleRequest{
			UserScheduleId:  scheduleId,
			GroupScheduleId: data.GroupScheduleId,
		}, session, tx)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.PostScheduleResponse{Id: scheduleId}, nil
}

func (h *Handlers) PostScheduleBrackets(w http.ResponseWriter, req *http.Request, data *types.PostScheduleBracketsRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PostScheduleBracketsResponse, error) {

	existingBracketIds := make([]string, 0)
	existingBrackets := make(map[string]*types.IScheduleBracket)
	newBrackets := make(map[string]*types.IScheduleBracket)

	for key, bracket := range data.Brackets {
		if util.IsUUID(key) {
			existingBracketIds = append(existingBracketIds, key)
			existingBrackets[key] = bracket
		} else if util.IsEpoch(key) {
			newBrackets[key] = bracket
		}
	}

	err := handleDeletedBrackets(req.Context(), data.UserScheduleId, existingBracketIds, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(existingBracketIds) > 0 {
		err := h.HandleExistingBrackets(req.Context(), existingBracketIds, existingBrackets, tx, session)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	if len(newBrackets) > 0 && data.UserScheduleId != "" {
		err := h.InsertNewBrackets(req.Context(), data.UserScheduleId, newBrackets, tx, session)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules")
	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules/"+data.UserScheduleId)
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/user_schedules/"+data.GroupScheduleId)

	return &types.PostScheduleBracketsResponse{Success: true}, nil
}

func (h *Handlers) PatchSchedule(w http.ResponseWriter, req *http.Request, data *types.PatchScheduleRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PatchScheduleResponse, error) {
	schedule := data.GetSchedule()

	var startTime, endTime *string
	if schedule.StartTime != nil {
		st := schedule.StartTime.AsTime().Format(time.RFC3339Nano)
		startTime = &st
	}
	if schedule.EndTime != nil {
		et := schedule.EndTime.AsTime().Format(time.RFC3339Nano)
		endTime = &et
	}

	_, err := tx.Exec(req.Context(), `
		UPDATE dbtable_schema.schedules
		SET name = $2, start_time = $3, end_time = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
		`, schedule.GetId(), schedule.GetName(), &startTime, &endTime, session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchScheduleResponse{Success: true}, nil
}

func (h *Handlers) GetSchedules(w http.ResponseWriter, req *http.Request, data *types.GetSchedulesRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetSchedulesResponse, error) {
	var schedules []*types.ISchedule

	err := h.Database.QueryRows(req.Context(), tx, &schedules, `
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

func (h *Handlers) GetScheduleById(w http.ResponseWriter, req *http.Request, data *types.GetScheduleByIdRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetScheduleByIdResponse, error) {
	var schedules []*types.ISchedule

	err := h.Database.QueryRows(req.Context(), tx, &schedules, `
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

func (h *Handlers) DeleteSchedule(w http.ResponseWriter, req *http.Request, data *types.DeleteScheduleRequest, session *types.UserSession, tx *clients.PoolTx) (*types.DeleteScheduleResponse, error) {
	for _, scheduleId := range strings.Split(data.GetIds(), ",") {
		err := handleDeletedBrackets(req.Context(), scheduleId, []string{}, tx)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = tx.Exec(req.Context(), `
			UPDATE dbtable_schema.schedules
			SET enabled = false
			WHERE id = $1
		`, scheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = tx.Exec(req.Context(), `
			DELETE FROM dbtable_schema.schedules
			WHERE dbtable_schema.schedules.id = $1
			AND NOT EXISTS (
				SELECT 1
				FROM dbtable_schema.schedule_brackets bracket
				WHERE bracket.schedule_id = $1
			)
		`, scheduleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = tx.Exec(req.Context(), `
			DELETE FROM dbtable_schema.group_user_schedules
			WHERE user_schedule_id = $1
			AND NOT EXISTS (
				SELECT 1
				FROM dbtable_schema.schedule_brackets bracket
				WHERE bracket.schedule_id = $1
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

func (h *Handlers) DisableSchedule(w http.ResponseWriter, req *http.Request, data *types.DisableScheduleRequest, session *types.UserSession, tx *clients.PoolTx) (*types.DisableScheduleResponse, error) {
	_, err := tx.Exec(req.Context(), `
		UPDATE dbtable_schema.schedules
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableScheduleResponse{Success: true}, nil
}

func (h *Handlers) HandleExistingBrackets(ctx context.Context, existingBracketIds []string, brackets map[string]*types.IScheduleBracket, tx *clients.PoolTx, session *types.UserSession) error {

	// Step 1. Get all existing slots and services ids

	rows, err := tx.Query(ctx, `
		SELECT id, schedule_bracket_id
		FROM dbtable_schema.schedule_bracket_slots
		WHERE schedule_bracket_id = ANY($1)
	`, pq.Array(existingBracketIds))
	if err != nil {
		return fmt.Errorf("failed to query existing slots: %w", err)
	}
	defer rows.Close()

	existingBracketSlotIds := make(map[string][]string)
	allExistingSlotIds := make([]string, 0)

	for rows.Next() {
		var id, bracketId string
		if err := rows.Scan(&id, &bracketId); err != nil {
			return util.ErrCheck(fmt.Errorf("failed to scan slot row: %w", err))
		}
		if _, ok := existingBracketSlotIds[bracketId]; !ok {
			existingBracketSlotIds[bracketId] = make([]string, 0)
		}
		existingBracketSlotIds[bracketId] = append(existingBracketSlotIds[bracketId], id)
		allExistingSlotIds = append(allExistingSlotIds, id)
	}

	rows, err = tx.Query(ctx, `
		SELECT schedule_bracket_id, service_id
		FROM dbtable_schema.schedule_bracket_services
		WHERE schedule_bracket_id = ANY($1)
	`, pq.Array(existingBracketIds))
	if err != nil {
		return util.ErrCheck(fmt.Errorf("failed to query existing services: %w", err))
	}
	defer rows.Close()

	existingBracketServiceIds := make(map[string][]string)
	for rows.Next() {
		var bracketId, serviceId string
		if err := rows.Scan(&bracketId, &serviceId); err != nil {
			return util.ErrCheck(fmt.Errorf("failed to scan service row: %w", err))
		}
		if _, ok := existingBracketServiceIds[bracketId]; !ok {
			existingBracketServiceIds[bracketId] = make([]string, 0)
		}
		existingBracketServiceIds[bracketId] = append(existingBracketServiceIds[bracketId], serviceId)
	}

	// Step 2. Figure out what should be preserved or inserted

	slotsToKeep := make([]string, 0)

	var slotsQuery strings.Builder
	slotsQuery.WriteString(`
		INSERT INTO dbtable_schema.schedule_bracket_slots
		(schedule_bracket_id, start_time, created_sub, group_id)
		VALUES
	`)
	slotsValues := []interface{}{}

	var servicesQuery strings.Builder
	servicesQuery.WriteString(`
		INSERT INTO dbtable_schema.schedule_bracket_services
		(schedule_bracket_id, service_id, created_sub, group_id)
		VALUES
	`)
	servicesValues := []interface{}{}

	for bracketId, bracket := range brackets {

		var slotsLen int
		for slotId, slot := range bracket.Slots {
			if util.IsUUID(slotId) {
				// Existing slot - mark to keep
				slotsToKeep = append(slotsToKeep, slotId)
			} else if util.IsEpoch(slotId) {
				h.Database.BuildInserts(&slotsQuery, 4, slotsLen)
				slotsValues = append(slotsValues, bracketId, slot.StartTime, session.UserSub, session.GroupId)
				slotsLen += 4
			}
		}

		var servicesLen int
		for serviceId := range bracket.Services {
			// Check if this service already exists for this bracket
			found := false
			if services, ok := existingBracketServiceIds[bracketId]; ok {
				for _, existing := range services {
					if existing == serviceId {
						found = true
						break
					}
				}
			}

			if !found {
				h.Database.BuildInserts(&servicesQuery, 4, servicesLen)
				servicesValues = append(servicesValues, bracketId, serviceId, session.UserSub, session.GroupId)
				servicesLen += 4
			}
		}
	}

	// Step 3. Handle deletions or disables of no longer required slots and services

	if len(allExistingSlotIds) > 0 {
		// Find slots to potentially delete
		slotsToCheck := make([]string, 0)
		for _, id := range allExistingSlotIds {
			keep := false
			for _, keepId := range slotsToKeep {
				if id == keepId {
					keep = true
					break
				}
			}
			if !keep {
				slotsToCheck = append(slotsToCheck, id)
			}
		}

		if len(slotsToCheck) > 0 {
			// Check which slots have active quotes
			rows, err := tx.Query(ctx, `
				SELECT schedule_bracket_slot_id
				FROM dbtable_schema.quotes
				WHERE schedule_bracket_slot_id = ANY($1) AND enabled = true
			`, pq.Array(slotsToCheck))
			if err != nil {
				return util.ErrCheck(fmt.Errorf("failed to check slots with quotes: %w", err))
			}

			slotsWithQuotes := make([]string, 0)
			for rows.Next() {
				var slotId string
				if err := rows.Scan(&slotId); err != nil {
					rows.Close()
					return util.ErrCheck(fmt.Errorf("failed to scan slot with quote: %w", err))
				}
				slotsWithQuotes = append(slotsWithQuotes, slotId)
			}
			rows.Close()

			// Disable slots with quotes
			if len(slotsWithQuotes) > 0 {
				_, err = tx.Exec(ctx, `
					UPDATE dbtable_schema.schedule_bracket_slots
					SET enabled = false
					WHERE id = ANY($1)
				`, pq.Array(slotsWithQuotes))
				if err != nil {
					return util.ErrCheck(fmt.Errorf("failed to disable slots with quotes: %w", err))
				}
			}

			// Delete slots without quotes
			slotsToDelete := make([]string, 0)
			for _, id := range slotsToCheck {
				hasQuote := false
				for _, quoteSlotId := range slotsWithQuotes {
					if id == quoteSlotId {
						hasQuote = true
						break
					}
				}
				if !hasQuote {
					slotsToDelete = append(slotsToDelete, id)
				}
			}

			if len(slotsToDelete) > 0 {
				_, err = tx.Exec(ctx, `
					DELETE FROM dbtable_schema.schedule_bracket_slots
					WHERE id = ANY($1)
				`, pq.Array(slotsToDelete))
				if err != nil {
					return util.ErrCheck(fmt.Errorf("failed to delete slots without quotes: %w", err))
				}
			}
		}
	}

	// Find services to disable or delete
	for bracketId, existingServiceIds := range existingBracketServiceIds {

		// See if existingServices differs from the incoming brackets
		bracketServices := make(map[string]bool)
		if bracket, ok := brackets[bracketId]; ok {
			for serviceId := range bracket.Services {
				bracketServices[serviceId] = true
			}
		}

		// Prepare to handle services that are not represented in the incoming bracket
		servicesToHandle := make([]string, 0)
		for _, serviceId := range existingServiceIds {
			if !bracketServices[serviceId] {
				servicesToHandle = append(servicesToHandle, serviceId)
			}
		}

		if len(servicesToHandle) > 0 {
			// Check if bracket has any slots with quotes
			var quoteCount int
			err := tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM dbtable_schema.quotes q
				JOIN dbtable_schema.schedule_bracket_slots sbs ON q.schedule_bracket_slot_id = sbs.id
				WHERE sbs.schedule_bracket_id = $1 AND q.enabled = true
			`, bracketId).Scan(&quoteCount)
			if err != nil {
				return util.ErrCheck(fmt.Errorf("failed to check quotes for bracket %s: %w", bracketId, err))
			}

			for _, serviceId := range servicesToHandle {
				if quoteCount > 0 {
					_, err := tx.Exec(ctx, `
						UPDATE dbtable_schema.schedule_bracket_services
						SET enabled = false
						WHERE schedule_bracket_id = $1 AND service_id = $2
					`, bracketId, serviceId)
					if err != nil {
						return util.ErrCheck(fmt.Errorf("failed to disable service %s for bracket %s: %w", serviceId, bracketId, err))
					}
				} else {
					_, err := tx.Exec(ctx, `
						DELETE FROM dbtable_schema.schedule_bracket_services
						WHERE schedule_bracket_id = $1 AND service_id = $2
					`, bracketId, serviceId)
					if err != nil {
						return util.ErrCheck(fmt.Errorf("failed to delete service %s for bracket %s: %w", serviceId, bracketId, err))
					}
				}
			}
		}
	}

	// Step 4. Perform data inserts

	if len(slotsValues) > 0 {
		_, err := tx.Exec(ctx, strings.TrimSuffix(slotsQuery.String(), ","), slotsValues...)
		if err != nil {
			return util.ErrCheck(fmt.Errorf("failed to execute slot inserts: %w", err))
		}
	}

	if len(servicesValues) > 0 {
		_, err := tx.Exec(ctx, strings.TrimSuffix(servicesQuery.String(), ","), servicesValues...)
		if err != nil {
			return util.ErrCheck(fmt.Errorf("failed to execute service inserts: %w", err))
		}
	}

	return nil
}

func (h *Handlers) InsertNewBrackets(ctx context.Context, scheduleId string, newBrackets map[string]*types.IScheduleBracket, tx *clients.PoolTx, session *types.UserSession) error {
	var slotsQuery strings.Builder
	slotsQuery.WriteString(`
		INSERT INTO dbtable_schema.schedule_bracket_slots
		(schedule_bracket_id, start_time, created_sub, group_id)
		VALUES
	`)
	slotsValues := []interface{}{}

	var servicesQuery strings.Builder
	servicesQuery.WriteString(`
		INSERT INTO dbtable_schema.schedule_bracket_services
		(schedule_bracket_id, service_id, created_sub, group_id)
		VALUES
	`)
	servicesValues := []interface{}{}

	for _, bracket := range newBrackets {
		err := tx.QueryRow(ctx, `
			INSERT INTO dbtable_schema.schedule_brackets (schedule_id, duration, multiplier, automatic, created_sub, group_id)
			VALUES ($1, $2, $3, $4, $5::uuid, $6)
			RETURNING id
		`, scheduleId, bracket.Duration, bracket.Multiplier, bracket.Automatic, session.UserSub, session.GroupId).Scan(&bracket.Id)
		if err != nil {
			return util.ErrCheck(fmt.Errorf("failed to insert bracket new bracket record: %w", err))
		}

		var slotsLen int
		for _, slot := range bracket.Slots {
			h.Database.BuildInserts(&slotsQuery, 4, slotsLen)
			slotsValues = append(slotsValues, bracket.Id, slot.StartTime, session.UserSub, session.GroupId)
			slotsLen += 4
		}

		var servicesLen int
		for _, service := range bracket.Services {
			h.Database.BuildInserts(&servicesQuery, 4, servicesLen)
			servicesValues = append(servicesValues, bracket.Id, service.Id, session.UserSub, session.GroupId)
			servicesLen += 4
		}
	}

	if len(slotsValues) > 0 {
		_, err := tx.Exec(ctx, strings.TrimSuffix(slotsQuery.String(), ","), slotsValues...)
		if err != nil {
			return util.ErrCheck(fmt.Errorf("failed to execute slot inserts: %w", err))
		}
	}

	if len(servicesValues) > 0 {
		_, err := tx.Exec(ctx, strings.TrimSuffix(servicesQuery.String(), ","), servicesValues...)
		if err != nil {
			return util.ErrCheck(fmt.Errorf("failed to execute service inserts: %w", err))
		}
	}

	return nil
}

// handleDeletedBrackets handles brackets that should be deleted or disabled
func handleDeletedBrackets(ctx context.Context, scheduleId string, existingBracketIds []string, tx *clients.PoolTx) error {
	if len(existingBracketIds) == 0 {
		// If no existing IDs are provided, we still need to get brackets that might have quotes
		rows, err := tx.Query(ctx, `
			SELECT DISTINCT sb.id
			FROM dbtable_schema.schedule_brackets sb
			JOIN dbtable_schema.schedule_bracket_slots sbs ON sb.id = sbs.schedule_bracket_id
			JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
			WHERE sb.schedule_id = $1 AND q.enabled = true
		`, scheduleId)
		if err != nil {
			return util.ErrCheck(fmt.Errorf("failed to query brackets with quotes: %w", err))
		}
		defer rows.Close()

		var bracketsToDisable []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return util.ErrCheck(fmt.Errorf("failed to scan bracket row: %w", err))
			}
			bracketsToDisable = append(bracketsToDisable, id)
		}

		// Delete brackets without quotes as no existing ids were given
		rows, err = tx.Query(ctx, `
			SELECT id FROM dbtable_schema.schedule_brackets
			WHERE schedule_id = $1
			AND NOT EXISTS (
				SELECT 1 FROM dbtable_schema.schedule_bracket_slots sbs
				JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
				WHERE sbs.schedule_bracket_id = dbtable_schema.schedule_brackets.id
				AND q.enabled = true
			)
		`, scheduleId)
		if err != nil {
			return util.ErrCheck(err)
		}

		defer rows.Close()

		var bracketsToDelete []string
		for rows.Next() {
			var bracketId string
			rows.Scan(&bracketId)
			bracketsToDelete = append(bracketsToDelete, bracketId)
		}

		return disableAndDeleteBrackets(ctx, bracketsToDisable, bracketsToDelete, tx)
	}

	// If we have existing IDs, only delete/disable brackets not in this list

	// First disable brackets with quotes
	rows, err := tx.Query(ctx, `
		SELECT id FROM dbtable_schema.schedule_brackets
		WHERE id IN (SELECT DISTINCT sb.id
			FROM dbtable_schema.schedule_brackets sb
			JOIN dbtable_schema.schedule_bracket_slots sbs ON sb.id = sbs.schedule_bracket_id
			JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
			WHERE sb.schedule_id = $1 
			AND sb.id NOT IN (SELECT unnest($2::uuid[]))
			AND q.enabled = true
		)
	`, scheduleId, pq.Array(existingBracketIds))
	if err != nil {
		return util.ErrCheck(fmt.Errorf("failed to disable brackets with quotes: %w", err))
	}
	defer rows.Close()

	var bracketsToDisable []string
	for rows.Next() {
		var bracketId string
		rows.Scan(&bracketId)
		bracketsToDisable = append(bracketsToDisable, bracketId)
	}

	// Then delete brackets without quotes
	rows, err = tx.Query(ctx, `
		SELECT id FROM dbtable_schema.schedule_brackets
		WHERE schedule_id = $1
		AND id NOT IN (SELECT unnest($2::uuid[]))
		AND NOT EXISTS (
			SELECT 1 FROM dbtable_schema.schedule_bracket_slots sbs
			JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
			WHERE sbs.schedule_bracket_id = dbtable_schema.schedule_brackets.id
			AND q.enabled = true
		)
	`, scheduleId, pq.Array(existingBracketIds))
	if err != nil {
		return util.ErrCheck(fmt.Errorf("failed to delete brackets without quotes: %w", err))
	}

	defer rows.Close()

	var bracketsToDelete []string
	for rows.Next() {
		var bracketId string
		rows.Scan(&bracketId)
		bracketsToDelete = append(bracketsToDelete, bracketId)
	}

	return disableAndDeleteBrackets(ctx, bracketsToDisable, bracketsToDelete, tx)
}

func disableAndDeleteBrackets(ctx context.Context, bracketsToDisable, bracketsToDelete []string, tx *clients.PoolTx) error {
	var err error
	if len(bracketsToDisable) > 0 {
		_, err = tx.Exec(ctx, `
			UPDATE dbtable_schema.schedule_bracket_slots
			SET enabled = false
			WHERE schedule_bracket_id = ANY($1)
		`, pq.Array(bracketsToDisable))
		if err != nil {
			return util.ErrCheck(err)
		}

		_, err = tx.Exec(ctx, `
			UPDATE dbtable_schema.schedule_bracket_services
			SET enabled = false
			WHERE schedule_bracket_id = ANY($1)
		`, pq.Array(bracketsToDisable))
		if err != nil {
			return util.ErrCheck(err)
		}

		_, err = tx.Exec(ctx, `
			UPDATE dbtable_schema.schedule_brackets
			SET enabled = false
			WHERE id = ANY($1)
		`, pq.Array(bracketsToDisable))
		if err != nil {
			return util.ErrCheck(err)
		}
	}

	if len(bracketsToDelete) > 0 {
		_, err = tx.Exec(ctx, `
			DELETE FROM dbtable_schema.schedule_bracket_slots
			WHERE schedule_bracket_id = ANY($1)
		`, pq.Array(bracketsToDelete))
		if err != nil {
			return util.ErrCheck(err)
		}

		_, err = tx.Exec(ctx, `
			DELETE FROM dbtable_schema.schedule_bracket_services
			WHERE schedule_bracket_id = ANY($1)
		`, pq.Array(bracketsToDelete))
		if err != nil {
			return util.ErrCheck(err)
		}

		_, err = tx.Exec(ctx, `
			DELETE FROM dbtable_schema.schedule_brackets
			WHERE id = ANY($1)
		`, pq.Array(bracketsToDelete))
		if err != nil {
			return util.ErrCheck(err)
		}
	}

	return nil
}
