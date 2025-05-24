package handlers

import (
	"slices"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) CheckGroupName(info ReqInfo, data *types.CheckGroupNameRequest) (*types.CheckGroupNameResponse, error) {
	var count int

	err := info.Tx.QueryRow(info.Ctx, `
		SELECT COUNT(id)
		FROM dbtable_schema.groups
		WHERE name = $1
	`, data.Name).Scan(&count)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.CheckGroupNameResponse{IsValid: count == 0}, nil
}

func (h *Handlers) JoinGroup(info ReqInfo, data *types.JoinGroupRequest) (*types.JoinGroupResponse, error) {
	userSub := info.Session.GetUserSub()

	// Get group information using the group code
	var groupId, allowedDomains, kcRoleSubGroupExternalId string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT g.id, g.allowed_domains, gr.external_id
		FROM dbtable_schema.groups g
		JOIN dbtable_schema.group_roles gr ON gr.role_id = g.default_role_id
		WHERE g.code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &kcRoleSubGroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Group not found."))
	}

	if allowedDomains != "" {
		allowedDomainsSlice := strings.Split(allowedDomains, ",")
		if !slices.Contains(allowedDomainsSlice, strings.Split(info.Session.GetUserEmail(), "@")[1]) {
			return nil, util.ErrCheck(util.UserError("Group access is restricted."))
		}
	}

	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.group_users (user_id, group_id, external_id, created_sub)
		SELECT id, $2, $3, $1::uuid
		FROM dbtable_schema.users
		WHERE created_sub = $1
	`, userSub, groupId, kcRoleSubGroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Skip add to group for registation joiners, keycloak must complete the registration
	if !data.GetRegistering() {
		// User sub twice for worker queue id + user id to add to role group
		err = h.Keycloak.AddUserToGroup(info.Ctx, userSub, userSub, kcRoleSubGroupExternalId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	_, err = h.ActivateProfile(info, nil)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.JoinGroupResponse{Success: true}, nil
}

func (h *Handlers) LeaveGroup(info ReqInfo, data *types.LeaveGroupRequest) (*types.LeaveGroupResponse, error) {
	var userId, groupId, allowedDomains, defaultRoleId string

	err := info.Tx.QueryRow(info.Ctx, `
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Group not found."))
	}

	err = info.Tx.QueryRow(info.Ctx, `SELECT id FROM dbtable_schema.users WHERE sub = $1`, info.Session.GetUserSub()).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.group_users WHERE user_id = $1 AND group_id = $2
	`, userId, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.DeleteUserFromGroup(info.Ctx, info.Session.GetUserSub(), info.Session.GetUserSub(), groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.LeaveGroupResponse{Success: true}, nil
}

// AttachUser
func (h *Handlers) AttachUser(info ReqInfo, data *types.AttachUserRequest) (*types.AttachUserResponse, error) {
	var kcRoleSubgroupExternalId string

	var groupId string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT g.id
		FROM dbtable_schema.groups g
		WHERE g.code = $1
	`, data.GetCode()).Scan(&groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	info.Session.SetGroupId(groupId)

	ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, info.Session)

	row, done, err := ds.SessionBatchQueryRow(info.Ctx, `
		SELECT gr.external_id
		FROM dbtable_schema.groups g
		JOIN dbtable_schema.group_roles gr ON gr.role_id = g.default_role_id
		WHERE g.id = $1
	`, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer done()

	err = row.Scan(&kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Sub passed twice here to act as worker pool key as well as data for the add
	err = h.Keycloak.AddUserToGroup(info.Ctx, info.Session.GetUserSub(), info.Session.GetUserSub(), kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if err := h.Socket.RoleCall(info.Session.GetUserSub()); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.AttachUserResponse{Success: true}, nil
}

func (h *Handlers) CompleteOnboarding(info ReqInfo, data *types.CompleteOnboardingRequest) (*types.CompleteOnboardingResponse, error) {
	service := data.GetService()
	schedule := data.GetSchedule()

	postServiceReq := &types.PostServiceRequest{
		Service: service,
	}
	postServiceRes, err := h.PostService(info, postServiceReq)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	postGroupServiceReq := &types.PostGroupServiceRequest{
		ServiceId: postServiceRes.Id,
	}
	postGroupServiceRes, err := h.PostGroupService(info, postGroupServiceReq)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	postScheduleReq := &types.PostScheduleRequest{
		AsGroup:            true,
		Name:               schedule.Name,
		StartDate:          schedule.StartDate,
		EndDate:            schedule.EndDate,
		ScheduleTimeUnitId: schedule.ScheduleTimeUnitId,
		BracketTimeUnitId:  schedule.BracketTimeUnitId,
		SlotTimeUnitId:     schedule.SlotTimeUnitId,
		SlotDuration:       schedule.SlotDuration,
		Brackets:           map[string]*types.IScheduleBracket{},
	}
	postScheduleRes, err := h.PostSchedule(info, postScheduleReq)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	postGroupScheduleRes, err := h.PostGroupSchedule(info, &types.PostGroupScheduleRequest{
		ScheduleId: postScheduleRes.Id,
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.ActivateProfile(info, nil)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	onboardingResponse := &types.CompleteOnboardingResponse{
		ServiceId:       postServiceRes.Id,
		GroupServiceId:  postGroupServiceRes.Id,
		ScheduleId:      postScheduleRes.Id,
		GroupScheduleId: postGroupScheduleRes.Id,
	}

	return onboardingResponse, nil
}
