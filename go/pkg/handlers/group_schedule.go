package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

func (h *Handlers) PostGroupSchedule(w http.ResponseWriter, req *http.Request, data *types.PostGroupScheduleRequest) (*types.PostGroupScheduleResponse, error) {
	session := h.Redis.ReqSession(req)
	tx, ongoing := h.Database.ReqTx(req)
	if tx == nil {
		return nil, util.ErrCheck(errors.New("bad post schedule tx"))
	}

	if !ongoing {
		defer tx.Rollback()
	}

	ctx := context.WithValue(req.Context(), "UserSub", session.GroupSub)

	req = req.WithContext(ctx)

	scheduleResp, err := h.PostSchedule(w, req, &types.PostScheduleRequest{Schedule: data.GetGroupSchedule().GetSchedule()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	scheduleId := scheduleResp.GetId()

	_, err = tx.Exec(`
		INSERT INTO dbtable_schema.group_schedules (group_id, schedule_id, created_sub)
		VALUES ($1, $2, $3::uuid)
	`, session.GroupId, scheduleId, session.GroupSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")

	if !ongoing {
		tx.Commit()
	}

	return &types.PostGroupScheduleResponse{Id: scheduleId}, nil
}

func (h *Handlers) PatchGroupSchedule(w http.ResponseWriter, req *http.Request, data *types.PatchGroupScheduleRequest) (*types.PatchGroupScheduleResponse, error) {
	session := h.Redis.ReqSession(req)

	scheduleResp, err := h.PatchSchedule(w, req, &types.PatchScheduleRequest{
		Schedule: data.GetGroupSchedule().GetSchedule(),
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules/master/"+data.GetGroupSchedule().GetScheduleId())

	return &types.PatchGroupScheduleResponse{Success: scheduleResp.Success}, nil
}

func (h *Handlers) GetGroupSchedules(w http.ResponseWriter, req *http.Request, data *types.GetGroupSchedulesRequest) (*types.GetGroupSchedulesResponse, error) {
	session := h.Redis.ReqSession(req)

	var groupSchedules []*types.IGroupSchedule
	err := h.Database.QueryRows(&groupSchedules, `
		SELECT TO_JSONB(es) as schedule, es.name, egs.id, egs."groupId"
		FROM dbview_schema.enabled_group_schedules egs
		LEFT JOIN dbview_schema.enabled_schedules es ON es.id = egs."scheduleId"
		WHERE egs."groupId" = $1
	`, session.GroupId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupSchedulesResponse{GroupSchedules: groupSchedules}, nil
}

func (h *Handlers) GetGroupScheduleMasterById(w http.ResponseWriter, req *http.Request, data *types.GetGroupScheduleMasterByIdRequest) (*types.GetGroupScheduleMasterByIdResponse, error) {
	session := h.Redis.ReqSession(req)

	var groupSchedules []*types.IGroupSchedule

	err := h.Database.QueryRows(&groupSchedules, `
		SELECT TO_JSONB(ese) as schedule, ese.name, true as master, ese.id as "scheduleId"
		FROM dbview_schema.enabled_schedules_ext ese
		JOIN dbtable_schema.schedules s ON s.id = ese.id
		JOIN dbtable_schema.users u ON u.sub = s.created_sub
		WHERE ese.id = $1 AND u.username = $2
	`, data.GetGroupScheduleId(), "system_group_"+session.GroupId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupSchedules) == 0 {
		return nil, util.ErrCheck(errors.New("schedule not found"))
	}

	return &types.GetGroupScheduleMasterByIdResponse{GroupSchedule: groupSchedules[0]}, nil
}

func (h *Handlers) GetGroupScheduleByDate(w http.ResponseWriter, req *http.Request, data *types.GetGroupScheduleByDateRequest) (*types.GetGroupScheduleByDateResponse, error) {

	// TODO limit to group only query
	tzString, err := base64.StdEncoding.DecodeString(data.GetTimezone())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var groupScheduleDateSlots []*types.IGroupScheduleDateSlots
	err = h.Database.QueryRows(&groupScheduleDateSlots, `
		SELECT * FROM dbfunc_schema.get_group_schedules($1, $2, $3)
	`, data.GetDate(), data.GetGroupScheduleId(), tzString)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupScheduleByDateResponse{GroupScheduleDateSlots: groupScheduleDateSlots}, nil
}

func (h *Handlers) DeleteGroupSchedule(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupScheduleRequest) (*types.DeleteGroupScheduleResponse, error) {
	session := h.Redis.ReqSession(req)

	// TODO tx for deletes

	for _, scheduleId := range strings.Split(data.GetIds(), ",") {
		_, err := h.Database.Client().Exec(`
			DELETE FROM dbtable_schema.group_schedules
			WHERE group_id = $1 AND schedule_id = $2
		`, session.GroupId, scheduleId)

		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/schedules")

	return &types.DeleteGroupScheduleResponse{Success: true}, nil
}
