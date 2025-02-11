package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (h *Handlers) PostGroup(w http.ResponseWriter, req *http.Request, data *types.PostGroupRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostGroupResponse, error) {
	var undos []func()

	defer func() {
		recover()
		if len(undos) > 0 {
			for _, undo := range undos {
				undo()
			}
		}
	}()

	var kcGroupExternalId, kcAdminSubgroupExternalId, groupId, groupName string

	// Create group system user
	groupSub, err := uuid.NewV7()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	session.GroupSub = groupSub.String()

	err = tx.SetDbVar("user_sub", session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		INSERT INTO dbtable_schema.users (sub, username, created_on, created_sub)
		VALUES ($1::uuid, $2, $3, $1::uuid)
	`, session.GroupSub, data.GetName(), time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.SetDbVar("user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Create group in application db
	// All the repeated $1 (UserSub) at the start are just placeholders until later in the method
	err = tx.QueryRow(`
		INSERT INTO dbtable_schema.groups (external_id, code, admin_role_external_id, name, purpose, allowed_domains, created_sub, display_name, ai, sub)
		VALUES ($1::uuid, $1, $1::uuid, $2, $3, $4, $1::uuid, $5, $6, $7)
		RETURNING id, name
	`, session.UserSub, data.GetName(), data.GetPurpose(), data.GetAllowedDomains(), data.GetDisplayName(), data.GetAi(), session.GroupSub).Scan(&groupId, &groupName)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Create group resource in Keycloak
	kcGroup, err := h.Keycloak.CreateGroup(data.GetName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if kcGroup.Id != "" {
		kcGroupExternalId = kcGroup.Id

		undos = append(undos, func() {
			err = h.Keycloak.DeleteGroup(kcGroupExternalId)
			if err != nil {
				util.ErrCheck(err)
			}
		})
	} else {
		return nil, util.ErrCheck(errors.New("error creating keycloak group"))
	}

	// Create Admin role subgroup
	kcAdminSubgroup, err := h.Keycloak.CreateOrGetSubGroup(kcGroupExternalId, "Admin")
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	kcAdminSubgroupExternalId = kcAdminSubgroup.Id

	// Add admin subgroup/role to the app db
	_, err = tx.Exec(`
		INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5::uuid)
		ON CONFLICT (group_id, role_id) DO NOTHING
	`, groupId, h.Database.AdminRoleId(), kcAdminSubgroupExternalId, time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Add admin roles to the admin subgroup
	err = h.Keycloak.AddRolesToGroup(kcAdminSubgroupExternalId, h.Keycloak.GetGroupAdminRoles())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Attach the user to the admin subgroup
	err = h.Keycloak.AddUserToGroup(session.UserSub, kcAdminSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Update the group with the keycloak reference id
	_, err = tx.Exec(`
		UPDATE dbtable_schema.groups 
		SET external_id = $2, admin_role_external_id = $3, purpose = $4
		WHERE id = $1
	`, groupId, kcGroupExternalId, kcAdminSubgroupExternalId, data.GetPurpose())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var adminUserId string
	err = tx.QueryRow(`
		SELECT id FROM dbview_schema.enabled_users WHERE sub = $1
	`, session.UserSub).Scan(&adminUserId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		INSERT INTO dbtable_schema.group_users (user_id, group_id, external_id, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
		ON CONFLICT (user_id, group_id) DO NOTHING
	`, adminUserId, groupId, kcAdminSubgroupExternalId, session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Keycloak.RoleCall(http.MethodPost, session.UserSub)
	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.DeleteSession(req.Context(), session.UserSub)
	h.Redis.SetGroupSessionVersion(req.Context(), groupId)

	undos = nil
	return &types.PostGroupResponse{Id: groupId}, nil
}

func (h *Handlers) PatchGroup(w http.ResponseWriter, req *http.Request, data *types.PatchGroupRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchGroupResponse, error) {

	_, err := tx.Exec(`
		UPDATE dbtable_schema.groups
		SET name = $2, purpose = $3, display_name = $4, updated_sub = $5, updated_on = $6, ai = $7
		WHERE id = $1
	`, session.GroupId, data.GetName(), data.GetPurpose(), data.GetDisplayName(), session.UserSub, time.Now().Local().UTC(), data.GetAi())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.SetDbVar("user_sub", session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		UPDATE dbtable_schema.users
		SET username = $2, updated_sub = $3, updated_on = $4
		WHERE sub = $1
	`, session.GroupSub, data.GetName(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.SetDbVar("user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Keycloak.UpdateGroup(session.GroupExternalId, data.GetName())

	h.Redis.DeleteSession(req.Context(), session.UserSub)
	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	h.Redis.SetGroupSessionVersion(req.Context(), session.GroupId)

	return &types.PatchGroupResponse{Success: true}, nil
}

func (h *Handlers) PatchGroupAssignments(w http.ResponseWriter, req *http.Request, data *types.PatchGroupAssignmentsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchGroupAssignmentsResponse, error) {
	assignmentsBytes, err := h.Redis.Client().Get(req.Context(), "group_role_assignments:"+session.GroupId).Bytes()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	groupRoleActions := make(map[string]*types.IGroupRoleAuthActions)

	json.Unmarshal(assignmentsBytes, &groupRoleActions)

	for sgPath, assignmentSet := range data.GetAssignments() {

		assignmentNames := []string{}
		for _, assignSet := range assignmentSet.Actions {
			assignmentNames = append(assignmentNames, assignSet.GetName())
		}

		if sgRoleActions, ok := groupRoleActions[sgPath]; ok {

			groupRoleActionNames := []string{}
			for _, groupRoleActionSet := range sgRoleActions.Actions {
				groupRoleActionNames = append(groupRoleActionNames, groupRoleActionSet.GetName())
			}

			deletions := []clients.KeycloakRole{}

			for _, sgRoleActionSet := range sgRoleActions.Actions {
				if !slices.Contains(assignmentNames, sgRoleActionSet.GetName()) {
					deletions = append(deletions, clients.KeycloakRole{
						Id:   sgRoleActionSet.GetId(),
						Name: sgRoleActionSet.GetName(),
					})
				}
			}

			if len(deletions) > 0 {
				err = h.Keycloak.DeleteRolesFromGroup(sgRoleActions.GetId(), deletions)
				if err != nil {
					return nil, util.ErrCheck(err)
				}
			}

			additions := []clients.KeycloakRole{}

			for _, assignSet := range assignmentSet.Actions {
				if !slices.Contains(groupRoleActionNames, assignSet.GetName()) {
					var roleId string
					for _, appRole := range h.Keycloak.GetGroupAdminRoles() {
						if appRole.Name == assignSet.GetName() {
							roleId = appRole.Id
							break
						}
					}
					if roleId != "" {
						additions = append(additions, clients.KeycloakRole{
							Id:   roleId,
							Name: assignSet.GetName(),
						})
					}
				}
			}

			if len(additions) > 0 {
				h.Keycloak.AddRolesToGroup(sgRoleActions.GetId(), additions)
			}
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/assignments")

	h.Redis.SetGroupSessionVersion(req.Context(), session.GroupId)

	return &types.PatchGroupAssignmentsResponse{Success: true}, nil
}

func (h *Handlers) GetGroupAssignments(w http.ResponseWriter, req *http.Request, data *types.GetGroupAssignmentsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetGroupAssignmentsResponse, error) {
	kcGroup, err := h.Keycloak.GetGroup(session.GroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	assignments := make(map[string]*types.IGroupRoleAuthActions)

	subgroups, err := h.Keycloak.GetGroupSubgroups(kcGroup.Id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, sg := range *subgroups {
		graa := &types.IGroupRoleAuthActions{
			Id:      sg.Id,
			Fetch:   false,
			Actions: []*types.IGroupRoleAuthAction{},
		}

		sgRoles := h.Keycloak.GetGroupSiteRoles(sg.Id)

		for _, sgRole := range sgRoles {
			roleAction := &types.IGroupRoleAuthAction{
				Id:   sgRole.Id,
				Name: sgRole.Name,
			}
			graa.Actions = append(graa.Actions, roleAction)
		}

		assignments[sg.Path] = graa
	}

	assignmentsBytes, err := json.Marshal(assignments)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defaultDuration, _ := time.ParseDuration("1d")

	h.Redis.Client().Set(req.Context(), "group_role_assignments:"+session.GroupId, assignmentsBytes, defaultDuration)

	return &types.GetGroupAssignmentsResponse{Assignments: assignments}, nil
}

func (h *Handlers) DeleteGroup(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteGroupResponse, error) {
	for _, id := range data.GetIds() {
		var groupExternalId string

		err := tx.QueryRow(`
			SELECT external_id FROM dbtable_schema.groups WHERE id = $1
		`, id).Scan(&groupExternalId)
		if err != nil || groupExternalId == "" {
			return nil, util.ErrCheck(err)
		}

		err = h.Keycloak.DeleteGroup(groupExternalId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = tx.Exec(`
			DELETE FROM dbtable_schema.group_roles WHERE group_id = $1;
			DELETE FROM dbtable_schema.groups WHERE id = $1;
		`, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.DeleteGroupResponse{Success: true}, nil
}

func (h *Handlers) CheckGroupName(w http.ResponseWriter, req *http.Request, data *types.CheckGroupNameRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.CheckGroupNameResponse, error) {
	var count int

	err := tx.QueryRow(`
		SELECT COUNT(*) FROM dbtable_schema.groups WHERE name = $1
	`, data.GetName()).Scan(&count)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if count > 0 {
		return nil, util.ErrCheck(util.UserError("Group name in use."))
	}

	return &types.CheckGroupNameResponse{IsValid: true}, nil
}

func (h *Handlers) JoinGroup(w http.ResponseWriter, req *http.Request, data *types.JoinGroupRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.JoinGroupResponse, error) {
	var userId, groupId, allowedDomains, defaultRoleId string

	err := tx.QueryRow(`
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(errors.New("Group not found."))
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

func (h *Handlers) LeaveGroup(w http.ResponseWriter, req *http.Request, data *types.LeaveGroupRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.LeaveGroupResponse, error) {
	var userId, groupId, allowedDomains, defaultRoleId string

	err := tx.QueryRow(`
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(errors.New("Group not found."))
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

	err = h.Keycloak.DeleteUserFromGroup(session.UserSub, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.LeaveGroupResponse{Success: true}, nil
}

// AttachUser
func (h *Handlers) AttachUser(w http.ResponseWriter, req *http.Request, data *types.AttachUserRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.AttachUserResponse, error) {
	var groupId, kcGroupExternalId, kcRoleSubgroupExternalId, defaultRoleId, createdSub string

	err := tx.QueryRow(`
		SELECT g.id, g.external_id, g.default_role_id, g.created_sub, gr.external_id FROM dbtable_schema.groups g
		JOIN dbtable_schema.group_roles gr ON gr.role_id = g.default_role_id
		WHERE g.code = $1
	`, data.GetCode()).Scan(&groupId, &kcGroupExternalId, &defaultRoleId, &createdSub, &kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddUserToGroup(session.UserSub, kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if err := h.Keycloak.RoleCall(http.MethodPost, session.UserSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.Client().Del(req.Context(), createdSub+"profile/details")
	h.Redis.DeleteSession(req.Context(), session.UserSub)
	h.Redis.SetGroupSessionVersion(req.Context(), groupId)

	return &types.AttachUserResponse{Success: true}, nil
}

func (h *Handlers) CompleteOnboarding(w http.ResponseWriter, req *http.Request, data *types.CompleteOnboardingRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.CompleteOnboardingResponse, error) {
	service := data.GetService()
	schedule := data.GetSchedule()

	postServiceRes, err := h.PostService(w, req, &types.PostServiceRequest{Service: service}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	service.Id = postServiceRes.GetId()

	_, err = h.PatchService(w, req, &types.PatchServiceRequest{Service: service}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.PostGroupService(w, req, &types.PostGroupServiceRequest{ServiceId: postServiceRes.GetId()}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.PostGroupSchedule(w, req, &types.PostGroupScheduleRequest{GroupSchedule: &types.IGroupSchedule{Schedule: schedule}}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.ActivateProfile(w, req, &types.ActivateProfileRequest{}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.CompleteOnboardingResponse{Success: true}, nil
}
