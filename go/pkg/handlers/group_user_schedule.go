package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupUserSchedule(w http.ResponseWriter, req *http.Request, data *types.PostGroupUserScheduleRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PostGroupUserScheduleResponse, error) {
	var groupUserScheduleId string

	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.group_user_schedules (group_schedule_id, user_schedule_id, created_sub, group_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid)
		ON CONFLICT (group_schedule_id, user_schedule_id) DO UPDATE SET updated_on = NOW()
		RETURNING id
	`, data.GetGroupScheduleId(), data.GetUserScheduleId(), session.UserSub, session.GroupId).Scan(&groupUserScheduleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/user_schedules/"+data.GetGroupScheduleId())
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/user_schedules_stubs")

	return &types.PostGroupUserScheduleResponse{Id: groupUserScheduleId}, nil
}

func (h *Handlers) GetGroupUserSchedules(w http.ResponseWriter, req *http.Request, data *types.GetGroupUserSchedulesRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.GetGroupUserSchedulesResponse, error) {
	var groupUserSchedules []*types.IGroupUserSchedule

	err := tx.QueryRows(&groupUserSchedules, `
		SELECT egus.*
		FROM dbview_schema.enabled_group_user_schedules_ext egus
		WHERE egus."groupScheduleId" = $1
	`, data.GetGroupScheduleId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUserSchedulesResponse{GroupUserSchedules: groupUserSchedules}, nil
}

func (h *Handlers) GetGroupUserScheduleStubs(w http.ResponseWriter, req *http.Request, data *types.GetGroupUserScheduleStubsRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.GetGroupUserScheduleStubsResponse, error) {
	var groupUserScheduleStubs []*types.IGroupUserScheduleStub

	err := tx.SetDbVar("user_sub", session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.QueryRows(&groupUserScheduleStubs, `
		SELECT guss.*, gus.group_schedule_id as "groupScheduleId"
		FROM dbview_schema.group_user_schedule_stubs guss
		JOIN dbtable_schema.group_user_schedules gus ON gus.user_schedule_id = guss."userScheduleId"
		JOIN dbtable_schema.schedules schedule ON schedule.id = gus.group_schedule_id
		JOIN dbview_schema.enabled_users eu ON eu.sub = schedule.created_sub
		JOIN dbtable_schema.users u ON u.id = eu.id
		WHERE u.sub = $1
	`, session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.SetDbVar("user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUserScheduleStubsResponse{GroupUserScheduleStubs: groupUserScheduleStubs}, nil
}

func (h *Handlers) GetGroupUserScheduleStubReplacement(w http.ResponseWriter, req *http.Request, data *types.GetGroupUserScheduleStubReplacementRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.GetGroupUserScheduleStubReplacementResponse, error) {
	var stubs []*types.IGroupUserScheduleStub

	err := tx.QueryRows(&stubs, `
		SELECT replacement FROM dbfunc_schema.get_peer_schedule_replacement($1::UUID[], $2::DATE, $3::INTERVAL, $4::TEXT)
	`, data.GetUserScheduleId(), data.GetSlotDate(), data.GetStartTime(), data.GetTierName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUserScheduleStubReplacementResponse{GroupUserScheduleStubs: stubs}, nil
}

func (h *Handlers) PatchGroupUserScheduleStubReplacement(w http.ResponseWriter, req *http.Request, data *types.PatchGroupUserScheduleStubReplacementRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PatchGroupUserScheduleStubReplacementResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.quotes
		SET slot_date = $2, schedule_bracket_slot_id = $3, service_tier_id = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
	`, data.GetQuoteId(), data.GetSlotDate(), data.GetScheduleBracketSlotId(), data.GetServiceTierId(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchGroupUserScheduleStubReplacementResponse{Success: true}, nil
}

func (h *Handlers) DeleteGroupUserScheduleByUserScheduleId(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupUserScheduleByUserScheduleIdRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.DeleteGroupUserScheduleByUserScheduleIdResponse, error) {
	idsSplit := strings.Split(data.GetIds(), ",")

	for _, userScheduleId := range idsSplit {
		var parts []*types.ScheduledParts

		err := tx.QueryRows(&parts, `SELECT * FROM dbfunc_schema.get_scheduled_parts($1);`, userScheduleId)
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
			_, err = tx.Exec(`
				DELETE FROM dbtable_schema.group_user_schedules
				WHERE user_schedule_id = $1
			`, userScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		} else {
			_, err = tx.Exec(`
				UPDATE dbtable_schema.group_user_schedules
				SET enabled = false
				WHERE user_schedule_id = $1
			`, userScheduleId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}

		var groupScheduleId string

		err = tx.QueryRow(`
			SELECT group_schedule_id as "groupScheduleId"
			FROM dbtable_schema.group_user_schedules
			WHERE user_schedule_id = $1
		`, userScheduleId).Scan(&groupScheduleId)

		h.Redis.Client().Del(req.Context(), session.UserSub+"group/user_schedules/"+groupScheduleId)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/user_schedules_stubs")

	return &types.DeleteGroupUserScheduleByUserScheduleIdResponse{Success: true}, nil
}
