package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/google/uuid"
)

func (h *Handlers) PostGroup(w http.ResponseWriter, req *http.Request, data *types.PostGroupRequest, session *types.UserSession, tx *sql.Tx) (*types.PostGroupResponse, error) {
	var undos []func()

	defer func() {
		recover()
		if len(undos) > 0 {
			for _, undo := range undos {
				undo()
			}
		}
	}()

	check, err := h.CheckGroupName(w, req, &types.CheckGroupNameRequest{Name: data.GetName()}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if !check.GetIsValid() {
		return nil, util.ErrCheck(util.UserError("Group name in use."))
	}

	var kcGroupExternalId, kcAdminSubgroupExternalId, groupId, groupName string

	// Create group system user
	groupSub, err := uuid.NewV7()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	session.GroupSub = groupSub.String()

	err = h.Database.SetDbVar(tx, "user_sub", session.GroupSub)
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

	err = h.Database.SetDbVar(tx, "user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Create group in application db
	// All the repeated $1 (UserSub) at the start are just placeholders until later in the method
	// group "code" is generated via TRIGGER set_group_code
	err = tx.QueryRow(`
		INSERT INTO dbtable_schema.groups (external_id, code, admin_role_external_id, name, purpose, allowed_domains, created_sub, display_name, ai, sub)
		VALUES ($1::uuid, $1, $1::uuid, $2, $3, $4, $1::uuid, $5, $6, $7)
		RETURNING id, name
	`, session.UserSub, data.GetName(), data.GetPurpose(), data.GetAllowedDomains(), data.GetDisplayName(), data.GetAi(), session.GroupSub).Scan(&groupId, &groupName)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Create group resource in Keycloak
	kcGroup, err := h.Keycloak.CreateGroup(session.UserSub, data.GetName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if kcGroup.Id != "" {
		kcGroupExternalId = kcGroup.Id

		undos = append(undos, func() {
			err = h.Keycloak.DeleteGroup(session.UserSub, kcGroupExternalId)
			if err != nil {
				util.ErrCheck(err)
			}
		})
	} else {
		return nil, util.ErrCheck(errors.New("error creating keycloak group"))
	}

	// Create Admin role subgroup
	kcAdminSubgroup, err := h.Keycloak.CreateOrGetSubGroup(session.UserSub, kcGroupExternalId, "Admin")
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
	roles, err := h.Keycloak.GetGroupAdminRoles(session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddRolesToGroup(session.UserSub, kcAdminSubgroupExternalId, roles)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Attach the user to the admin subgroup
	err = h.Keycloak.AddUserToGroup(session.UserSub, session.UserSub, kcAdminSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Update the group with the keycloak reference id and get the group code
	var groupCode string
	err = tx.QueryRow(`
		UPDATE dbtable_schema.groups 
		SET external_id = $2, admin_role_external_id = $3, purpose = $4
		WHERE id = $1
		RETURNING code
	`, groupId, kcGroupExternalId, kcAdminSubgroupExternalId, data.GetPurpose()).Scan(&groupCode)
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

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.DeleteSession(req.Context(), session.UserSub)
	h.Redis.SetGroupSessionVersion(req.Context(), groupId)
	h.Socket.RoleCall(session.UserSub)

	undos = nil
	return &types.PostGroupResponse{Id: groupId, Code: groupCode}, nil
}

func (h *Handlers) PatchGroup(w http.ResponseWriter, req *http.Request, data *types.PatchGroupRequest, session *types.UserSession, tx *sql.Tx) (*types.PatchGroupResponse, error) {

	_, err := tx.Exec(`
		UPDATE dbtable_schema.groups
		SET name = $2, purpose = $3, display_name = $4, updated_sub = $5, updated_on = $6, ai = $7
		WHERE id = $1
	`, session.GroupId, data.GetName(), data.GetPurpose(), data.GetDisplayName(), session.UserSub, time.Now().Local().UTC(), data.GetAi())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.SetDbVar(tx, "user_sub", session.GroupSub)
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

	err = h.Database.SetDbVar(tx, "user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Keycloak.UpdateGroup(session.UserSub, session.GroupExternalId, data.GetName())

	h.Redis.DeleteSession(req.Context(), session.UserSub)
	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	h.Redis.SetGroupSessionVersion(req.Context(), session.GroupId)

	return &types.PatchGroupResponse{Success: true}, nil
}

func (h *Handlers) PatchGroupAssignments(w http.ResponseWriter, req *http.Request, data *types.PatchGroupAssignmentsRequest, session *types.UserSession, tx *sql.Tx) (*types.PatchGroupAssignmentsResponse, error) {
	assignmentsBytes, err := h.Redis.Client().Get(req.Context(), "group_role_assignments:"+session.GroupId).Bytes()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	groupRoleActions := make(map[string]*types.IGroupRoleAuthActions)

	json.Unmarshal(assignmentsBytes, &groupRoleActions)

	adminRoles, err := h.Keycloak.GetGroupAdminRoles(session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for sgPath, assignmentSet := range data.Assignments {

		assignmentNames := []string{}
		for _, assignSet := range assignmentSet.Actions {
			assignmentNames = append(assignmentNames, assignSet.Name)
		}

		if sgRoleActions, ok := groupRoleActions[sgPath]; ok {

			groupRoleActionNames := []string{}
			for _, groupRoleActionSet := range sgRoleActions.Actions {
				groupRoleActionNames = append(groupRoleActionNames, groupRoleActionSet.Name)
			}

			deletions := []*types.KeycloakRole{}

			for _, sgRoleActionSet := range sgRoleActions.Actions {
				if !slices.Contains(assignmentNames, sgRoleActionSet.Name) {
					deletions = append(deletions, &types.KeycloakRole{
						Id:   sgRoleActionSet.Id,
						Name: sgRoleActionSet.Name,
					})
				}
			}

			if len(deletions) > 0 {
				response, err := h.Keycloak.SendCommand(clients.DeleteRolesFromGroupKeycloakCommand, &types.AuthRequestParams{
					GroupId: sgRoleActions.Id,
					Roles:   deletions,
				})
				if err = clients.ChannelError(err, response.Error); err != nil {
					return nil, util.ErrCheck(err)
				}

			}

			additions := []*types.KeycloakRole{}

			for _, assignSet := range assignmentSet.Actions {
				if !slices.Contains(groupRoleActionNames, assignSet.Name) {
					var roleId string
					for _, appRole := range adminRoles {
						if appRole.Name == assignSet.Name {
							roleId = appRole.Id
							break
						}
					}
					if roleId != "" {
						additions = append(additions, &types.KeycloakRole{
							Id:   roleId,
							Name: assignSet.Name,
						})
					}
				}
			}

			if len(additions) > 0 {
				h.Keycloak.AddRolesToGroup(session.UserSub, sgRoleActions.Id, additions)
			}
		}
	}

	// h.Redis.Client().Del(req.Context(), session.UserSub+"group/assignments")

	h.Redis.SetGroupSessionVersion(req.Context(), session.GroupId)

	return &types.PatchGroupAssignmentsResponse{Success: true}, nil
}

func (h *Handlers) GetGroupAssignments(w http.ResponseWriter, req *http.Request, data *types.GetGroupAssignmentsRequest, session *types.UserSession, tx *sql.Tx) (*types.GetGroupAssignmentsResponse, error) {
	kcGroup, err := h.Keycloak.GetGroup(session.UserSub, session.GroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	assignments := make(map[string]*types.IGroupRoleAuthActions)
	assignmentsWithoutId := make(map[string]*types.IGroupRoleAuthActions)

	subgroups, err := h.Keycloak.GetGroupSubgroups(session.UserSub, kcGroup.Id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, sg := range subgroups {
		graaWI := &types.IGroupRoleAuthActions{
			Actions: []*types.IGroupRoleAuthAction{},
		}
		graa := &types.IGroupRoleAuthActions{
			Id:      sg.Id,
			Actions: []*types.IGroupRoleAuthAction{},
		}

		sgRoles, err := h.Keycloak.GetGroupSiteRoles(session.UserSub, sg.Id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, sgRole := range sgRoles {
			graaWI.Actions = append(graaWI.Actions, &types.IGroupRoleAuthAction{
				Name: sgRole.Name,
			})
			graa.Actions = append(graa.Actions, &types.IGroupRoleAuthAction{
				Id:   sgRole.Id,
				Name: sgRole.Name,
			})
		}

		assignmentsWithoutId[sg.Path] = graaWI
		assignments[sg.Path] = graa
	}

	assignmentsBytes, err := json.Marshal(assignments)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defaultDuration, _ := time.ParseDuration("1d")

	h.Redis.Client().Set(req.Context(), "group_role_assignments:"+session.GroupId, assignmentsBytes, defaultDuration)

	return &types.GetGroupAssignmentsResponse{Assignments: assignmentsWithoutId}, nil
}

func (h *Handlers) DeleteGroup(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupRequest, session *types.UserSession, tx *sql.Tx) (*types.DeleteGroupResponse, error) {
	for _, id := range data.GetIds() {
		var groupExternalId string

		err := tx.QueryRow(`
			SELECT external_id FROM dbtable_schema.groups WHERE id = $1
		`, id).Scan(&groupExternalId)
		if err != nil || groupExternalId == "" {
			return nil, util.ErrCheck(err)
		}

		err = h.Keycloak.DeleteGroup(session.UserSub, groupExternalId)
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
