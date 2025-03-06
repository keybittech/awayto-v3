package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

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

	existingBracketIds := make([]string, 0)
	existingBrackets := make(map[string]*types.IScheduleBracket)
	newBrackets := make(map[string]*types.IScheduleBracket)

	for key, bracket := range data.Brackets {
		if util.IsUUID(key) {
			existingBracketIds = append(existingBracketIds, key)
			existingBrackets[key] = bracket
		} else if util.IsEpoch(key) {
			newBrackets[key] = bracket
		} else {
			return nil, util.ErrCheck(fmt.Errorf("invalid bracket key: %s (must be UUID or epoch timestamp)", key))
		}
	}

	// 1. Handle existing bracket updates
	if len(existingBracketIds) > 0 {
		err := handleExistingBrackets(existingBracketIds, existingBrackets, tx, session)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	// 2. Insert new brackets and their components
	if len(newBrackets) > 0 {
		err := insertNewBrackets(data.ScheduleId, newBrackets, tx, session)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	// 3. Handle brackets that should be deleted or disabled
	err := handleDeletedBrackets(data.ScheduleId, existingBracketIds, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// // When submitting brackets, it's a mix of new and existing
	// // New brackets are just unix epoch timestamps while existing
	// // records are uuids. collect just the uuids here to pass when
	// // removing brackets, to ignore deleting existing brackets
	// existingBracketIds := make([]string, 0, len(data.Brackets))
	// for k := range data.Brackets {
	// 	_, err := strconv.Atoi(k)
	// 	if err != nil {
	// 		existingBracketIds = append(existingBracketIds, k)
	// 	}
	// }
	//
	// if err := h.RemoveScheduleBrackets(req.Context(), tx, data.ScheduleId, existingBracketIds); err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	//
	// for _, b := range data.Brackets {
	//
	// 	err := tx.QueryRow(`
	// 		INSERT INTO dbtable_schema.schedule_brackets (schedule_id, duration, multiplier, automatic, created_sub, group_id)
	// 		VALUES ($1, $2, $3, $4, $5::uuid, $6)
	// 		RETURNING id
	// 	`, data.GetScheduleId(), b.Duration, b.Multiplier, b.Automatic, session.UserSub, session.GroupId).Scan(&b.Id)
	//
	// 	if err != nil {
	// 		return nil, util.ErrCheck(err)
	// 	}
	//
	// 	for _, serv := range b.Services {
	// 		_, err = tx.Exec(`
	// 			INSERT INTO dbtable_schema.schedule_bracket_services (schedule_bracket_id, service_id, created_sub, group_id)
	// 			VALUES ($1, $2, $3::uuid, $4::uuid)
	// 		`, b.Id, serv.Id, session.UserSub, session.GroupId)
	// 		if err != nil {
	// 			return nil, util.ErrCheck(err)
	// 		}
	// 	}
	//
	// 	newSlots := make(map[string]*types.IScheduleBracketSlot)
	//
	// 	_, err = tx.Exec(`
	// 		DELETE FROM dbtable_schema.schedule_bracket_slots
	// 		WHERE schedule_bracket_id = $1
	// 	`, b.Id)
	// 	if err != nil {
	// 		return nil, util.ErrCheck(err)
	// 	}
	//
	// 	for _, slot := range b.Slots {
	// 		var slotId string
	//
	// 		err = tx.QueryRow(`
	// 			INSERT INTO dbtable_schema.schedule_bracket_slots (schedule_bracket_id, start_time, created_sub, group_id)
	// 			VALUES ($1, $2::interval, $3::uuid, $4::uuid)
	// 			RETURNING id
	// 		`, b.Id, slot.StartTime, session.UserSub, session.GroupId).Scan(&slotId)
	// 		if err != nil {
	// 			return nil, util.ErrCheck(err)
	// 		}
	//
	// 		slot.Id = slotId
	// 		slot.ScheduleBracketId = b.Id
	// 		newSlots[slotId] = slot
	// 	}
	//
	// 	b.Slots = newSlots
	// }

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules/"+data.GetScheduleId())
	h.Redis.Client().Del(req.Context(), session.UserSub+"schedules")

	return &types.PostScheduleBracketsResponse{Success: true}, nil
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

	err := tx.QueryRows(&schedules, `
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

		if err := h.RemoveScheduleBrackets(req.Context(), tx, scheduleId); err != nil {
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

// The method for a user to update their available times to work is
// 1. get_scheduled_parts takes a scheduleId and returns rows for
// bracket_slots and bracket_services if any are associated to
// the schedule. this means that someone has requested services from
// the schedule, and the slot/service are the when/what of the data
// 2. next we get all the schedule_bracket ids using schedule_id and
// loop through them, deleting

func (h *Handlers) RemoveScheduleBrackets(ctx context.Context, tx clients.IDatabaseTx, scheduleId string, persistingBracketIds ...[]string) error {

	// var serviceDel, slotDel, bracketDel, bracketUp, slotUp, serviceUp int64
	// err := tx.QueryRow(`
	// 	WITH scheduled_records AS ( -- use schedule id to find records attached to a quote
	// 		SELECT brackets.id as bracket_id, slots.id as slot_id, services.id as service_id
	// 		FROM dbtable_schema.schedules schedules
	// 		JOIN dbtable_schema.schedule_brackets brackets ON brackets.schedule_id = schedules.id
	// 		JOIN dbtable_schema.schedule_bracket_slots slots ON slots.schedule_bracket_id = brackets.id
	// 		JOIN dbtable_schema.schedule_bracket_services services ON services.schedule_bracket_id = brackets.id
	// 		LEFT JOIN dbtable_schema.quotes quotes ON quotes.schedule_bracket_slot_id = slots.id
	// 		WHERE schedules.id = $1::uuid AND quotes.id IS NOT NULL
	// 	),
	// 	schedule_bracket_ids AS ( -- get all schedule_brackets for the schedule, except ones that are persisting
	// 		SELECT id FROM dbtable_schema.schedule_brackets
	// 		WHERE schedule_id = $1::uuid
	// 		AND id <> ANY($2::uuid[])
	// 	),
	// 	service_deletions AS ( -- delete schedule brackets that are not attached to a quote, and are not persisting
	// 		DELETE FROM dbtable_schema.schedule_bracket_services
	// 		WHERE id NOT IN (SELECT service_id FROM scheduled_records)
	// 		AND schedule_bracket_id IN (SELECT id FROM schedule_bracket_ids)
	// 		RETURNING id
	// 	),
	// 	slot_deletions AS (
	// 		DELETE FROM dbtable_schema.schedule_bracket_slots
	// 		WHERE id NOT IN (SELECT slot_id FROM scheduled_records)
	// 		AND schedule_bracket_id IN (SELECT id FROM schedule_bracket_ids)
	// 		RETURNING id
	// 	),
	// 	bracket_deletions AS (
	// 		DELETE FROM dbtable_schema.schedule_brackets schedule_bracket
	// 		WHERE id NOT IN (SELECT bracket_id FROM scheduled_records)
	// 		AND id IN (SELECT id FROM schedule_bracket_ids)
	// 		RETURNING id
	// 	),
	// 	bracket_updates AS ( -- disable schedule_bracket related records that are attached to a quote but no longer attached to the user's schedule (they will need to be rescheduled later with the schedule stubs flow, disabling hides them in the ui)
	// 		UPDATE dbtable_schema.schedule_brackets
	// 		SET enabled = false
	// 		FROM scheduled_records WHERE id = bracket_id
	// 		AND id <> ANY($2::uuid[]) -- don't modify persisting brackets
	// 		RETURNING id
	// 	),
	// 	slot_updates AS (
	// 		UPDATE dbtable_schema.schedule_bracket_slots
	// 		SET enabled = false
	// 		FROM scheduled_records WHERE id = slot_id
	// 		AND schedule_bracket_id <> ANY($2::uuid[])
	// 		RETURNING id
	// 	),
	// 	service_updates AS (
	// 		UPDATE dbtable_schema.schedule_bracket_services
	// 		SET enabled = false
	// 		FROM scheduled_records WHERE id = scheduled_records.service_id
	// 		AND schedule_bracket_id <> ANY($2::uuid[])
	// 		RETURNING id
	// 	)
	// 	SELECT
	// 		(SELECT COUNT(*) FROM service_deletions),
	// 		(SELECT COUNT(*) FROM slot_deletions),
	// 		(SELECT COUNT(*) FROM bracket_deletions),
	// 		(SELECT COUNT(*) FROM bracket_updates),
	// 		(SELECT COUNT(*) FROM slot_updates),
	// 		(SELECT COUNT(*) FROM service_updates)
	//
	// 	-- SELECT
	// 	-- 	COUNT(bracket_id),
	// 	-- 	COUNT(bracket_id),
	// 	-- 	COUNT(slot_id),
	// 	-- 	COUNT(slot_id),
	// 	-- 	COUNT(service_id),
	// 	-- 	COUNT(service_id) FROM scheduled_records
	// `, scheduleId, pq.Array(persistingBracketIds)).Scan(&serviceDel, &slotDel, &bracketDel, &bracketUp, &slotUp, &serviceUp)
	// if err != nil {
	// 	return util.ErrCheck(err)
	// }
	//
	// println(fmt.Sprintf("completed update with services: d%d u%d, slots: d%d u%d, brackets: d%d u%d", serviceDel, serviceUp, slotDel, slotUp, bracketDel, bracketUp))
	//
	// return nil
	//
	// var parts []*types.ScheduledParts
	//
	// err := tx.QueryRows(&parts, `
	//    SELECT * FROM dbfunc_schema.get_scheduled_parts($1);
	//  `, scheduleId)
	// if err != nil {
	// 	return util.ErrCheck(err)
	// }
	//
	// scheduledSlots := &types.ScheduledParts{}
	// hasSlots := false
	// scheduledServices := &types.ScheduledParts{}
	// hasServices := false
	//
	// for _, p := range parts {
	// 	if len(p.GetIds()) > 0 {
	// 		if p.GetParttype() == "slot" {
	// 			scheduledSlots = p
	// 			hasSlots = true
	// 		} else {
	// 			scheduledServices = p
	// 			hasServices = true
	// 		}
	// 	}
	// }
	//
	// rows, err := tx.Query(`
	//    SELECT sb.id
	// 	FROM dbtable_schema.schedule_brackets sb
	//    WHERE sb.schedule_id = $1
	//  `, scheduleId)
	// if err != nil {
	// 	return util.ErrCheck(err)
	// }
	//
	// defer rows.Close()
	//
	// var ids []string
	// for rows.Next() {
	// 	var idstr string
	// 	rows.Scan(&idstr)
	// 	ids = append(ids, idstr)
	// }
	//
	// for _, bracketId := range ids {
	//
	// 	if hasSlots || hasServices {
	// 		var bracketDelCount, slotDelCount, serviceDelCount int64
	//
	// 		if hasSlots {
	// 			err = tx.QueryRow(`
	// 				WITH deletions AS (
	// 					DELETE FROM dbtable_schema.schedule_bracket_slots
	// 					WHERE schedule_bracket_id = $1 AND id <> ALL($2::uuid[])
	// 					RETURNING id
	// 				)
	// 				SELECT COUNT(*) FROM deletions
	// 			`, bracketId, pq.Array(scheduledSlots.GetIds())).Scan(&slotDelCount)
	// 			if err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 				return util.ErrCheck(err)
	// 			}
	// 			_, err = tx.Exec(`
	// 				UPDATE dbtable_schema.schedule_bracket_slots
	// 				SET enabled = false
	// 				WHERE schedule_bracket_id = $1 AND id = ANY($2::uuid[])
	// 			`, bracketId, pq.Array(scheduledSlots.GetIds()))
	// 			if err != nil {
	// 				return util.ErrCheck(err)
	// 			}
	// 		}
	//
	// 		if hasServices {
	// 			err = tx.QueryRow(`
	// 				WITH deletions AS (
	// 					DELETE FROM dbtable_schema.schedule_bracket_services
	// 					WHERE schedule_bracket_id = $1 AND id <> ALL($2::uuid[])
	// 					RETURNING id
	// 				)
	// 				SELECT COUNT(*) FROM deletions
	// 			`, bracketId, pq.Array(scheduledServices.GetIds())).Scan(&serviceDelCount)
	// 			if err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 				return util.ErrCheck(err)
	// 			}
	//
	// 			_, err = tx.Exec(`
	// 				UPDATE dbtable_schema.schedule_bracket_services
	// 				SET enabled = false
	// 				WHERE schedule_bracket_id = $1 AND id = ANY($2::uuid[])
	// 			`, bracketId, pq.Array(scheduledServices.GetIds()))
	// 			if err != nil {
	// 				return util.ErrCheck(err)
	// 			}
	// 		}
	//
	// 		err = tx.QueryRow(`
	// 			WITH scheduled_ids AS (
	// 				SELECT $2::uuid[] as slot_ids, $3::uuid[] as service_ids
	// 			),
	// 			deletions AS (
	// 				DELETE FROM dbtable_schema.schedule_brackets schedule_bracket
	// 				WHERE schedule_bracket.id = $1::uuid
	// 				AND NOT EXISTS(
	// 					SELECT 1 FROM dbtable_schema.schedule_bracket_slots slots
	// 					JOIN scheduled_ids sids ON slots.id = ANY(sids.slot_ids)
	// 					WHERE slots.schedule_bracket_id = schedule_bracket.id
	// 				)
	// 				AND NOT EXISTS(
	// 					SELECT 1 FROM dbtable_schema.schedule_bracket_services services
	// 					JOIN scheduled_ids sids ON services.id = ANY(sids.service_ids)
	// 					WHERE services.schedule_bracket_id = schedule_bracket.id
	// 				)
	// 				RETURNING schedule_bracket.id
	// 			)
	// 			SELECT COUNT(*) FROM deletions
	// 		`, bracketId, pq.Array(scheduledSlots.GetIds()), pq.Array(scheduledServices.GetIds())).Scan(&bracketDelCount)
	// 		if err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 			return util.ErrCheck(err)
	// 		}
	//
	// 		_, err = tx.Exec(`
	// 			UPDATE dbtable_schema.schedule_brackets
	// 			SET enabled = false
	// 			WHERE id = $1
	// 		`, bracketId)
	// 		if err != nil {
	// 			return util.ErrCheck(err)
	// 		}
	// 	} else {
	// 		println(3)
	// 		_, err = tx.Exec(`
	// 			DELETE FROM dbtable_schema.schedule_brackets
	// 			WHERE id = $1
	// 		`, bracketId)
	// 		if err != nil {
	// 			return util.ErrCheck(err)
	// 		}
	// 	}
	// }
	//
	return nil
}

func handleExistingBrackets(existingIds []string, brackets map[string]*types.IScheduleBracket, tx clients.IDatabaseTx, session *clients.UserSession) error {

	// Step 1. Get all existing slots and services ids

	// First, get all existing slots for these brackets
	rows, err := tx.Query(`
		SELECT id, schedule_bracket_id
		FROM dbtable_schema.schedule_bracket_slots
		WHERE schedule_bracket_id = ANY($1)
	`, pq.Array(existingIds))
	if err != nil {
		return fmt.Errorf("failed to query existing slots: %w", err)
	}
	defer rows.Close()

	existingSlotIds := make(map[string][]string)
	allExistingSlotIds := make([]string, 0)

	for rows.Next() {
		var id, bracketId string
		if err := rows.Scan(&id, &bracketId); err != nil {
			return fmt.Errorf("failed to scan slot row: %w", err)
		}
		if _, ok := existingSlotIds[bracketId]; !ok {
			existingSlotIds[bracketId] = make([]string, 0)
		}
		existingSlotIds[bracketId] = append(existingSlotIds[bracketId], id)
		allExistingSlotIds = append(allExistingSlotIds, id)
	}

	// Get existing services for these brackets
	rows, err = tx.Query(`
		SELECT schedule_bracket_id, service_id
		FROM dbtable_schema.schedule_bracket_services
		WHERE schedule_bracket_id = ANY($1)
	`, pq.Array(existingIds))
	if err != nil {
		return fmt.Errorf("failed to query existing services: %w", err)
	}
	defer rows.Close()

	// Map of bracket ID -> service IDs
	existingServices := make(map[string][]string)
	for rows.Next() {
		var bracketId, serviceId string
		if err := rows.Scan(&bracketId, &serviceId); err != nil {
			return fmt.Errorf("failed to scan service row: %w", err)
		}
		if _, ok := existingServices[bracketId]; !ok {
			existingServices[bracketId] = make([]string, 0)
		}
		existingServices[bracketId] = append(existingServices[bracketId], serviceId)
	}

	// Step 2. Figure out what should be preserved or inserted

	type slotToInsert struct {
		BracketID string
		StartTime string
	}

	type serviceKey struct {
		BracketID string
		ServiceID string
	}

	slotsToKeep := make([]string, 0)

	var copySlots, copyServices bool

	slotStmt, err := tx.Prepare(pq.CopyInSchema("dbtable_schema", "schedule_bracket_slots",
		"schedule_bracket_id", "start_time", "created_sub", "group_id"))
	if err != nil {
		return fmt.Errorf("failed to prepare CopyIn statement: %w", err)
	}

	serviceStmt, err := tx.Prepare(pq.CopyInSchema("dbtable_schema", "schedule_bracket_services",
		"schedule_bracket_id", "service_id", "created_sub", "group_id"))
	if err != nil {
		return fmt.Errorf("failed to prepare CopyIn statement for services: %w", err)
	}

	for bracketId, bracket := range brackets {

		for slotId, slot := range bracket.Slots {
			if util.IsUUID(slotId) {
				// Existing slot - mark to keep
				slotsToKeep = append(slotsToKeep, slotId)
			} else if util.IsEpoch(slotId) {
				copySlots = true
				_, err = slotStmt.Exec(slot.ScheduleBracketId, slot.StartTime, session.UserSub, session.GroupId)
				if err != nil {
					slotStmt.Close()
					return fmt.Errorf("failed to execute CopyIn: %w", err)
				}
			}
		}

		for serviceId := range bracket.Services {
			// Check if this service already exists for this bracket
			found := false
			if services, ok := existingServices[bracketId]; ok {
				for _, existing := range services {
					if existing == serviceId {
						found = true
						break
					}
				}
			}

			if !found {
				copyServices = true
				_, err = serviceStmt.Exec(bracketId, serviceId, session.UserSub, session.GroupId)
				if err != nil {
					serviceStmt.Close()
					return fmt.Errorf("failed to execute CopyIn for services: %w", err)
				}
			}
		}
	}

	if copySlots {
		_, err = slotStmt.Exec()
		if err != nil {
			return fmt.Errorf("failed to complete CopyIn: %w", err)
		}
	}

	err = slotStmt.Close()
	if err != nil {
		return fmt.Errorf("failed to close CopyIn statement: %w", err)
	}

	if copyServices {
		_, err = serviceStmt.Exec()
		if err != nil {
			return fmt.Errorf("failed to complete CopyIn for services: %w", err)
		}
	}

	err = serviceStmt.Close()
	if err != nil {
		return fmt.Errorf("failed to close CopyIn statement for services: %w", err)
	}

	// Step 3. Handle deletions or disables of existing slots and services

	if len(allExistingSlotIds) > 0 && len(slotsToKeep) > 0 {
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
			rows, err := tx.Query(`
				SELECT schedule_bracket_slot_id
				FROM dbtable_schema.quotes
				WHERE schedule_bracket_slot_id = ANY($1) AND enabled = true
			`, pq.Array(slotsToCheck))
			if err != nil {
				return fmt.Errorf("failed to check slots with quotes: %w", err)
			}

			slotsWithQuotes := make([]string, 0)
			for rows.Next() {
				var slotId string
				if err := rows.Scan(&slotId); err != nil {
					rows.Close()
					return fmt.Errorf("failed to scan slot with quote: %w", err)
				}
				slotsWithQuotes = append(slotsWithQuotes, slotId)
			}
			rows.Close()

			// Disable slots with quotes
			if len(slotsWithQuotes) > 0 {
				_, err = tx.Exec(`
					UPDATE dbtable_schema.schedule_bracket_slots
					SET enabled = false
					WHERE id = ANY($1)
				`, pq.Array(slotsWithQuotes))
				if err != nil {
					return fmt.Errorf("failed to disable slots with quotes: %w", err)
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
				_, err = tx.Exec(`
					DELETE FROM dbtable_schema.schedule_bracket_slots
					WHERE id = ANY($1)
				`, pq.Array(slotsToDelete))
				if err != nil {
					return fmt.Errorf("failed to delete slots without quotes: %w", err)
				}
			}
		}
	}

	// Find services to disable or delete
	for bracketId, existingServiceIds := range existingServices {

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
			err := tx.QueryRow(`
				SELECT COUNT(*) FROM dbtable_schema.quotes q
				JOIN dbtable_schema.schedule_bracket_slots sbs ON q.schedule_bracket_slot_id = sbs.id
				WHERE sbs.schedule_bracket_id = $1 AND q.enabled = true
			`, bracketId).Scan(&quoteCount)
			if err != nil {
				return fmt.Errorf("failed to check quotes for bracket %s: %w", bracketId, err)
			}

			for _, serviceId := range servicesToHandle {
				if quoteCount > 0 {
					// Disable service
					_, err := tx.Exec(`
						UPDATE dbtable_schema.schedule_bracket_services
						SET enabled = false
						WHERE schedule_bracket_id = $1 AND service_id = $2
					`, bracketId, serviceId)
					if err != nil {
						return fmt.Errorf("failed to disable service %s for bracket %s: %w",
							serviceId, bracketId, err)
					}
				} else {
					// Delete service
					_, err := tx.Exec(`
						DELETE FROM dbtable_schema.schedule_bracket_services
						WHERE schedule_bracket_id = $1 AND service_id = $2
					`, bracketId, serviceId)
					if err != nil {
						return fmt.Errorf("failed to delete service %s for bracket %s: %w",
							serviceId, bracketId, err)
					}
				}
			}
		}
	}

	return nil
}

// insertNewBrackets inserts new brackets with their slots and services
func insertNewBrackets(scheduleId string, newBrackets map[string]*types.IScheduleBracket, tx clients.IDatabaseTx, session *clients.UserSession) error {
	if len(newBrackets) == 0 {
		return nil
	}

	for _, bracket := range newBrackets {
		// Insert bracket and get its new ID
		err := tx.QueryRow(`
			INSERT INTO dbtable_schema.schedule_brackets (schedule_id, duration, multiplier, automatic, created_sub, group_id)
			VALUES ($1, $2, $3, $4, $5::uuid, $6)
			RETURNING id
		`, scheduleId, bracket.Duration, bracket.Multiplier, bracket.Automatic, session.UserSub, session.GroupId).Scan(&bracket.Id)
		if err != nil {
			return util.ErrCheck(err)
		}

		// Batch insert slots
		if len(bracket.Slots) > 0 {
			stmt, err := tx.Prepare(pq.CopyInSchema("dbtable_schema", "schedule_bracket_slots",
				"schedule_bracket_id", "start_time", "enabled"))
			if err != nil {
				return fmt.Errorf("failed to prepare CopyIn for slots: %w", err)
			}

			for _, slot := range bracket.Slots {
				_, err = stmt.Exec(bracket.Id, slot.StartTime, true)
				if err != nil {
					stmt.Close()
					return fmt.Errorf("failed to execute CopyIn for slots: %w", err)
				}
			}

			_, err = stmt.Exec()
			if err != nil {
				return fmt.Errorf("failed to complete CopyIn for slots: %w", err)
			}

			err = stmt.Close()
			if err != nil {
				return fmt.Errorf("failed to close CopyIn statement for slots: %w", err)
			}
		}

		// Batch insert services
		if len(bracket.Services) > 0 {
			stmt, err := tx.Prepare(pq.CopyInSchema("dbtable_schema", "schedule_bracket_services",
				"schedule_bracket_id", "service_id", "enabled"))
			if err != nil {
				return fmt.Errorf("failed to prepare CopyIn for bracket services: %w", err)
			}

			for serviceId := range bracket.Services {
				_, err = stmt.Exec(bracket.Id, serviceId, true)
				if err != nil {
					stmt.Close()
					return fmt.Errorf("failed to execute CopyIn for bracket services: %w", err)
				}
			}

			_, err = stmt.Exec()
			if err != nil {
				return fmt.Errorf("failed to complete CopyIn for bracket services: %w", err)
			}

			err = stmt.Close()
			if err != nil {
				return fmt.Errorf("failed to close CopyIn statement for bracket services: %w", err)
			}
		}
	}

	return nil
}

// handleDeletedBrackets handles brackets that should be deleted or disabled
func handleDeletedBrackets(scheduleId string, existingBracketIds []string, tx clients.IDatabaseTx) error {
	if len(existingBracketIds) == 0 {
		// If no existing IDs are provided, we still need to get brackets that might have quotes
		rows, err := tx.Query(`
			SELECT DISTINCT sb.id
			FROM dbtable_schema.schedule_brackets sb
			JOIN dbtable_schema.schedule_bracket_slots sbs ON sb.id = sbs.schedule_bracket_id
			JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
			WHERE sb.schedule_id = $1 AND q.enabled = true
		`, scheduleId)
		if err != nil {
			return fmt.Errorf("failed to query brackets with quotes: %w", err)
		}
		defer rows.Close()

		var bracketsToDisable []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return fmt.Errorf("failed to scan bracket row: %w", err)
			}
			bracketsToDisable = append(bracketsToDisable, id)
		}

		if len(bracketsToDisable) > 0 {
			// Disable these brackets
			_, err = tx.Exec(`
				UPDATE dbtable_schema.schedule_brackets
				SET enabled = false
				WHERE id = ANY($1)
			`, pq.Array(bracketsToDisable))
			if err != nil {
				return fmt.Errorf("failed to disable brackets with quotes: %w", err)
			}

			// Disable their slots and services
			_, err = tx.Exec(`
				UPDATE dbtable_schema.schedule_bracket_slots
				SET enabled = false
				WHERE schedule_bracket_id = ANY($1)
			`, pq.Array(bracketsToDisable))
			if err != nil {
				return fmt.Errorf("failed to disable slots for brackets with quotes: %w", err)
			}

			_, err = tx.Exec(`
				UPDATE dbtable_schema.schedule_bracket_services
				SET enabled = false
				WHERE schedule_bracket_id = ANY($1)
			`, pq.Array(bracketsToDisable))
			if err != nil {
				return fmt.Errorf("failed to disable services for brackets with quotes: %w", err)
			}
		}

		// Delete brackets without quotes as no existing ids were given
		_, err = tx.Exec(`
			DELETE FROM dbtable_schema.schedule_brackets
			WHERE schedule_id = $1
			AND NOT EXISTS (
				SELECT 1 FROM dbtable_schema.schedule_bracket_slots sbs
				JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
				WHERE sbs.schedule_bracket_id = dbtable_schema.schedule_brackets.id
				AND q.enabled = true
			)
		`, scheduleId)
		if err != nil {
			return fmt.Errorf("failed to delete brackets without quotes: %w", err)
		}

		return nil
	}

	// If we have existing IDs, only delete/disable brackets not in this list
	_, err := tx.Exec(`
		WITH brackets_with_quotes AS (
			SELECT DISTINCT sb.id
			FROM dbtable_schema.schedule_brackets sb
			JOIN dbtable_schema.schedule_bracket_slots sbs ON sb.id = sbs.schedule_bracket_id
			JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
			WHERE sb.schedule_id = $1 
			AND sb.id NOT IN (SELECT unnest($2::uuid[]))
			AND q.enabled = true
		)
		UPDATE dbtable_schema.schedule_brackets
		SET enabled = false
		WHERE id IN (SELECT id FROM brackets_with_quotes)
	`, scheduleId, pq.Array(existingBracketIds))
	if err != nil {
		return fmt.Errorf("failed to disable brackets with quotes: %w", err)
	}

	// // Disable slots and services for these brackets
	// _, err = tx.Exec(`
	// 	UPDATE dbtable_schema.schedule_bracket_slots
	// 	SET enabled = false
	// 	WHERE schedule_bracket_id IN (
	// 		SELECT DISTINCT sb.id
	// 		FROM dbtable_schema.schedule_brackets sb
	// 		JOIN dbtable_schema.schedule_bracket_slots sbs ON sb.id = sbs.schedule_bracket_id
	// 		JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
	// 		WHERE sb.schedule_id = $1
	// 		AND sb.id NOT IN (SELECT unnest($2::uuid[]))
	// 		AND q.enabled = true
	// 	)
	// `, scheduleId, pq.Array(existingBracketIds))
	// if err != nil {
	// 	return fmt.Errorf("failed to disable slots for brackets with quotes: %w", err)
	// }
	//
	// _, err = tx.Exec(`
	// 	UPDATE dbtable_schema.schedule_bracket_services
	// 	SET enabled = false
	// 	WHERE schedule_bracket_id IN (
	// 		SELECT DISTINCT sb.id
	// 		FROM dbtable_schema.schedule_brackets sb
	// 		JOIN dbtable_schema.schedule_bracket_slots sbs ON sb.id = sbs.schedule_bracket_id
	// 		JOIN dbtable_schema.quotes q ON sbs.id = q.schedule_bracket_slot_id
	// 		WHERE sb.schedule_id = $1
	// 		AND sb.id NOT IN (SELECT unnest($2::uuid[]))
	// 		AND q.enabled = true
	// 	)
	// `, scheduleId, pq.Array(existingBracketIds))
	// if err != nil {
	// 	return fmt.Errorf("failed to disable services for brackets with quotes: %w", err)
	// }

	// Delete brackets without quotes
	_, err = tx.Exec(`
		DELETE FROM dbtable_schema.schedule_brackets
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
		return fmt.Errorf("failed to delete brackets without quotes: %w", err)
	}

	return nil
}
