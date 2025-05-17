package handlers

import (
	json "encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
)

var groupNameInUseError error = util.UserError("Group name in use.")

func (h *Handlers) PostGroup(info ReqInfo, data *types.PostGroupRequest) (*types.PostGroupResponse, error) {
	var undos []func()

	defer func() {
		if undos != nil && len(undos) > 0 {
			for _, undo := range undos {
				undo()
			}
		}
	}()

	check, err := h.CheckGroupName(info, &types.CheckGroupNameRequest{Name: data.Name})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if !check.IsValid {
		return nil, util.ErrCheck(groupNameInUseError)
	}

	var kcGroupExternalId, kcAdminSubGroupExternalId, groupId, groupName string

	// Create group system user
	groupSub, err := uuid.NewV7()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	info.Session.GroupSub = groupSub.String()

	ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, info.Session)

	_, err = ds.SessionBatchExec(info.Ctx, `
		INSERT INTO dbtable_schema.users (sub, created_sub, username, created_on)
		VALUES ($1::uuid, $1::uuid, $2, $3)
	`, info.Session.GroupSub, data.Name, time.Now())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Create group in application db
	// All the repeated $1 (UserSub) at the start are just placeholders until later in the method
	// group "code" is generated via TRIGGER set_group_code
	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.groups (external_id, code, admin_role_external_id, created_sub, name, purpose, allowed_domains, display_name, ai, sub)
		VALUES ($1::uuid, $1, $1::uuid, $1::uuid, $2, $3, $4, $5, $6, $7)
		RETURNING id, name
	`, info.Session.UserSub, data.Name, data.Purpose, data.AllowedDomains, data.DisplayName, data.Ai, info.Session.GroupSub).Scan(&groupId, &groupName)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Create group resource in Keycloak
	kcGroupIdOnly, err := h.Keycloak.CreateGroup(info.Ctx, info.Session.UserSub, data.GetName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if kcGroupIdOnly.Id != "" {
		kcGroupExternalId = kcGroupIdOnly.Id

		undos = append(undos, func() {
			err = h.Keycloak.DeleteGroup(info.Ctx, info.Session.UserSub, kcGroupExternalId)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
			}
		})
	} else {
		return nil, util.ErrCheck(errors.New("error creating keycloak group"))
	}

	// Create Admin role subgroup
	kcAdminSubGroup, err := h.Keycloak.CreateOrGetSubGroup(info.Ctx, info.Session.UserSub, kcGroupExternalId, "Admin")
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	kcAdminSubGroupExternalId = kcAdminSubGroup.Id

	// Add admin subgroup/role to the app db
	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5::uuid)
		ON CONFLICT (group_id, role_id) DO NOTHING
	`, groupId, h.Database.AdminRoleId(), kcAdminSubGroupExternalId, time.Now(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Add admin roles to the admin subgroup
	roles, err := h.Keycloak.GetGroupAdminRoles(info.Ctx, info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddRolesToGroup(info.Ctx, info.Session.UserSub, kcAdminSubGroupExternalId, roles)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Attach the user to the admin subgroup
	err = h.Keycloak.AddUserToGroup(info.Ctx, info.Session.UserSub, info.Session.UserSub, kcAdminSubGroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Update the group with the keycloak reference id and get the group code
	var groupCode string
	err = info.Tx.QueryRow(info.Ctx, `
		UPDATE dbtable_schema.groups 
		SET external_id = $2, admin_role_external_id = $3, purpose = $4
		WHERE id = $1
		RETURNING code
	`, groupId, kcGroupExternalId, kcAdminSubGroupExternalId, data.Purpose).Scan(&groupCode)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var adminUserId string
	err = info.Tx.QueryRow(info.Ctx, `
		SELECT id FROM dbview_schema.enabled_users WHERE sub = $1
	`, info.Session.UserSub).Scan(&adminUserId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.group_users (user_id, group_id, external_id, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
		ON CONFLICT (user_id, group_id) DO NOTHING
	`, adminUserId, groupId, kcAdminSubGroupExternalId, info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	newGroupPath := "/" + data.Name

	h.Cache.SetGroupSessionVersion(groupId)
	h.Cache.SetCachedGroup(newGroupPath, groupId, kcGroupExternalId, groupSub.String(), data.Name, data.Ai)
	h.Cache.SetCachedSubGroup(kcAdminSubGroup.Path, kcAdminSubGroup.Id, kcAdminSubGroup.Name, newGroupPath)
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")
	_ = h.Socket.RoleCall(info.Session.UserSub)

	undos = nil
	return &types.PostGroupResponse{Code: groupCode}, nil
}

func (h *Handlers) PatchGroup(info ReqInfo, data *types.PatchGroupRequest) (*types.PatchGroupResponse, error) {

	cachedGroup := h.Cache.GetCachedGroup(info.Session.GroupPath)
	nameChanged := cachedGroup.Name != data.Name

	if nameChanged {
		check, err := h.CheckGroupName(info, &types.CheckGroupNameRequest{Name: data.Name})
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if !check.GetIsValid() {
			return nil, util.ErrCheck(groupNameInUseError)
		}
	}

	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.groups
		SET name = $2, purpose = $3, display_name = $4, updated_sub = $5, updated_on = $6, ai = $7
		WHERE id = $1
	`, info.Session.GroupId, data.Name, data.Purpose, data.DisplayName, info.Session.UserSub, time.Now(), data.Ai)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, info.Session)

	_, err = ds.SessionBatchExec(info.Ctx, `
		UPDATE dbtable_schema.users
		SET username = $2, updated_sub = $3, updated_on = $4
		WHERE sub = $1
	`, info.Session.GroupSub, data.Name, info.Session.UserSub, time.Now())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// If the group name changed, make a new group cache entry
	if nameChanged {
		newGroupPath := strings.Replace(info.Session.GroupPath, cachedGroup.Name, data.Name, 1)
		h.Cache.SetCachedGroup(newGroupPath, cachedGroup.Id, cachedGroup.ExternalId, cachedGroup.Sub, data.Name, data.Ai)

		err = h.Keycloak.UpdateGroup(info.Ctx, info.Session.UserSub, info.Session.GroupExternalId, data.Name)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		h.Cache.UnsetCachedGroup(info.Session.GroupPath)
		info.Session.GroupPath = newGroupPath
		h.Cache.SetSessionToken(info.Req.Header.Get("Authorization"), info.Session)
	}

	h.Cache.SetGroupSessionVersion(info.Session.GroupId)
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")

	return &types.PatchGroupResponse{Success: true}, nil
}

func (h *Handlers) PatchGroupAssignments(info ReqInfo, data *types.PatchGroupAssignmentsRequest) (*types.PatchGroupAssignmentsResponse, error) {
	assignmentsBytes, err := h.Redis.Client().Get(info.Ctx, "group_role_assignments:"+info.Session.GroupId).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, util.ErrCheck(err)
	}

	groupRoleActions := make(map[string]*types.IGroupRoleAuthActions)

	err = json.Unmarshal(assignmentsBytes, &groupRoleActions)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	adminRoles, err := h.Keycloak.GetGroupAdminRoles(info.Ctx, info.Session.UserSub)
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
				_, err := h.Keycloak.SendCommand(info.Ctx, clients.DeleteRolesFromGroupKeycloakCommand, &types.AuthRequestParams{
					UserSub: info.Session.UserSub,
					GroupId: sgRoleActions.Id,
					Roles:   deletions,
				})
				if err != nil {
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
				err = h.Keycloak.AddRolesToGroup(info.Ctx, info.Session.UserSub, sgRoleActions.Id, additions)
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
				}
			}
		}
	}

	// h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/assignments")

	h.Cache.SetGroupSessionVersion(info.Session.GroupId)

	return &types.PatchGroupAssignmentsResponse{Success: true}, nil
}

func (h *Handlers) GetGroupAssignments(info ReqInfo, data *types.GetGroupAssignmentsRequest) (*types.GetGroupAssignmentsResponse, error) {
	kcGroup, err := h.Keycloak.GetGroup(info.Ctx, info.Session.UserSub, info.Session.GroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	assignments := make(map[string]*types.IGroupRoleAuthActions)
	assignmentsWithoutId := make(map[string]*types.IGroupRoleAuthActions)

	subgroups, err := h.Keycloak.GetGroupSubGroups(info.Ctx, info.Session.UserSub, kcGroup.Id)
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

		sgRoles, err := h.Keycloak.GetGroupSiteRoles(info.Ctx, info.Session.UserSub, sg.Id)
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

	h.Redis.Client().Set(info.Ctx, "group_role_assignments:"+info.Session.GroupId, assignmentsBytes, defaultDuration)

	return &types.GetGroupAssignmentsResponse{Assignments: assignmentsWithoutId}, nil
}

func (h *Handlers) DeleteGroup(info ReqInfo, data *types.DeleteGroupRequest) (*types.DeleteGroupResponse, error) {
	for _, id := range data.Ids {
		var groupExternalId string

		err := info.Tx.QueryRow(info.Ctx, `
			SELECT external_id FROM dbtable_schema.groups WHERE id = $1
		`, id).Scan(&groupExternalId)
		if err != nil || groupExternalId == "" {
			return nil, util.ErrCheck(err)
		}

		err = h.Keycloak.DeleteGroup(info.Ctx, info.Session.UserSub, groupExternalId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = info.Tx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.group_roles WHERE group_id = $1;
			DELETE FROM dbtable_schema.groups WHERE id = $1;
		`, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.DeleteGroupResponse{Success: true}, nil
}
