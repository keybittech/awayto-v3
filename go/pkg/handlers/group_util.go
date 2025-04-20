package handlers

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) CheckGroupName(w http.ResponseWriter, req *http.Request, data *types.CheckGroupNameRequest, session *types.UserSession, tx *clients.PoolTx) (*types.CheckGroupNameResponse, error) {
	var count int

	time.Sleep(time.Second)

	err := tx.QueryRow(`
		SELECT COUNT(*) FROM dbtable_schema.groups WHERE name = $1
	`, data.GetName()).Scan(&count)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.CheckGroupNameResponse{IsValid: count == 0}, nil
}

func (h *Handlers) JoinGroup(w http.ResponseWriter, req *http.Request, data *types.JoinGroupRequest, session *types.UserSession, tx *clients.PoolTx) (*types.JoinGroupResponse, error) {
	var userId, groupId, allowedDomains, defaultRoleId string

	err := tx.QueryRow(`
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Group not found."))
	}

	err = tx.QueryRow(`SELECT id FROM dbtable_schema.users WHERE sub = $1`, session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if allowedDomains != "" {
		allowedDomainsSlice := strings.Split(allowedDomains, ",")
		if !slices.Contains(allowedDomainsSlice, strings.Split(session.UserEmail, "@")[1]) {
			return nil, util.ErrCheck(util.UserError("Group access is restricted."))
		}
	}

	err = h.Database.SetDbVar("group_id", groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var kcSubgroupExternalId string
	err = tx.QueryRow(`
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE role_id = $1
	`, defaultRoleId).Scan(&kcSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		INSERT INTO dbtable_schema.group_users (user_id, group_id, external_id, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
	`, userId, groupId, kcSubgroupExternalId, session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.SetGroupSessionVersion(req.Context(), groupId)
	h.Redis.DeleteSession(req.Context(), session.UserSub)

	return &types.JoinGroupResponse{Success: true}, nil
}

func (h *Handlers) LeaveGroup(w http.ResponseWriter, req *http.Request, data *types.LeaveGroupRequest, session *types.UserSession, tx *clients.PoolTx) (*types.LeaveGroupResponse, error) {
	var userId, groupId, allowedDomains, defaultRoleId string

	err := tx.QueryRow(`
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Group not found."))
	}

	err = tx.QueryRow(`SELECT id FROM dbtable_schema.users WHERE sub = $1`, session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		DELETE FROM dbtable_schema.group_users WHERE user_id = $1 AND group_id = $2
	`, userId, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.DeleteUserFromGroup(session.UserSub, session.UserSub, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.LeaveGroupResponse{Success: true}, nil
}

// AttachUser
func (h *Handlers) AttachUser(w http.ResponseWriter, req *http.Request, data *types.AttachUserRequest, session *types.UserSession, tx *clients.PoolTx) (*types.AttachUserResponse, error) {
	var groupId, kcGroupExternalId, kcRoleSubgroupExternalId, defaultRoleId, createdSub string

	err := tx.QueryRow(`SELECT g.id FROM dbtable_schema.groups g WHERE g.code = $1`, data.GetCode()).Scan(&groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.SetDbVar("group_id", groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.QueryRow(`
		SELECT g.external_id, g.default_role_id, g.created_sub, gr.external_id FROM dbtable_schema.groups g
		JOIN dbtable_schema.group_roles gr ON gr.role_id = g.default_role_id
		WHERE g.id = $1
	`, groupId).Scan(&kcGroupExternalId, &defaultRoleId, &createdSub, &kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddUserToGroup(session.UserSub, session.UserSub, kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if err := h.Socket.RoleCall(session.UserSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.Client().Del(req.Context(), createdSub+"profile/details")
	h.Redis.DeleteSession(req.Context(), session.UserSub)
	h.Redis.SetGroupSessionVersion(req.Context(), groupId)

	return &types.AttachUserResponse{Success: true}, nil
}

func (h *Handlers) CompleteOnboarding(w http.ResponseWriter, req *http.Request, data *types.CompleteOnboardingRequest, session *types.UserSession, tx *clients.PoolTx) (*types.CompleteOnboardingResponse, error) {
	service := data.GetService()
	schedule := data.GetSchedule()

	postServiceReq := &types.PostServiceRequest{
		Service: service,
	}
	postServiceRes, err := h.PostService(w, req, postServiceReq, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	postGroupServiceReq := &types.PostGroupServiceRequest{
		ServiceId: postServiceRes.Id,
	}
	postGroupServiceRes, err := h.PostGroupService(w, req, postGroupServiceReq, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	postScheduleReq := &types.PostScheduleRequest{
		AsGroup:            true,
		Name:               schedule.Name,
		StartTime:          schedule.StartTime,
		EndTime:            schedule.EndTime,
		ScheduleTimeUnitId: schedule.ScheduleTimeUnitId,
		BracketTimeUnitId:  schedule.BracketTimeUnitId,
		SlotTimeUnitId:     schedule.SlotTimeUnitId,
		SlotDuration:       schedule.SlotDuration,
		Brackets:           map[string]*types.IScheduleBracket{},
	}
	postScheduleRes, err := h.PostSchedule(w, req, postScheduleReq, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	postGroupScheduleReq := &types.PostGroupScheduleRequest{
		ScheduleId: postScheduleRes.Id,
	}
	postGroupScheduleRes, err := h.PostGroupSchedule(w, req, postGroupScheduleReq, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.ActivateProfile(w, req, &types.ActivateProfileRequest{}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.CompleteOnboardingResponse{
		ServiceId:       postServiceRes.Id,
		GroupServiceId:  postGroupServiceRes.Id,
		ScheduleId:      postScheduleRes.Id,
		GroupScheduleId: postGroupScheduleRes.Id,
	}, nil
}
