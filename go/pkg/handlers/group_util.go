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
		SELECT COUNT(*) FROM dbtable_schema.groups WHERE name = $1
	`, data.GetName()).Scan(&count)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.CheckGroupNameResponse{IsValid: count == 0}, nil
}

func (h *Handlers) JoinGroup(info ReqInfo, data *types.JoinGroupRequest) (*types.JoinGroupResponse, error) {
	var userId, allowedDomains, defaultRoleId string

	err := info.Tx.QueryRow(info.Ctx, `
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&info.Session.GroupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Group not found."))
	}

	err = info.Tx.QueryRow(info.Ctx, `
		SELECT id
		FROM dbtable_schema.users
		WHERE sub = $1
	`, info.Session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if allowedDomains != "" {
		allowedDomainsSlice := strings.Split(allowedDomains, ",")
		if !slices.Contains(allowedDomainsSlice, strings.Split(info.Session.UserEmail, "@")[1]) {
			return nil, util.ErrCheck(util.UserError("Group access is restricted."))
		}
	}

	err = info.Tx.SetSession(info.Ctx, info.Session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var kcSubgroupExternalId string
	err = info.Tx.QueryRow(info.Ctx, `
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE role_id = $1
	`, defaultRoleId).Scan(&kcSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.group_users (user_id, group_id, external_id, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
	`, userId, info.Session.GroupId, kcSubgroupExternalId, info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Cache.SetGroupSessionVersion(info.Session.GroupId)

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

	err = info.Tx.QueryRow(info.Ctx, `SELECT id FROM dbtable_schema.users WHERE sub = $1`, info.Session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.group_users WHERE user_id = $1 AND group_id = $2
	`, userId, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.DeleteUserFromGroup(info.Ctx, info.Session.UserSub, info.Session.UserSub, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.LeaveGroupResponse{Success: true}, nil
}

// AttachUser
func (h *Handlers) AttachUser(info ReqInfo, data *types.AttachUserRequest) (*types.AttachUserResponse, error) {
	var kcRoleSubgroupExternalId, createdSub string

	err := info.Tx.QueryRow(info.Ctx, `
		SELECT g.id
		FROM dbtable_schema.groups g
		WHERE g.code = $1
	`, data.GetCode()).Scan(&info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, info.Session)

	row, done, err := ds.SessionBatchQueryRow(info.Ctx, `
		SELECT g.created_sub, gr.external_id
		FROM dbtable_schema.groups g
		JOIN dbtable_schema.group_roles gr ON gr.role_id = g.default_role_id
		WHERE g.id = $1
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer done()

	err = row.Scan(&createdSub, &kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddUserToGroup(info.Ctx, info.Session.UserSub, info.Session.UserSub, kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if err := h.Socket.RoleCall(info.Session.UserSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Cache.SetGroupSessionVersion(info.Session.GroupId)
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")
	h.Redis.Client().Del(info.Ctx, createdSub+"profile/details")

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
		StartTime:          schedule.StartTime,
		EndTime:            schedule.EndTime,
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
