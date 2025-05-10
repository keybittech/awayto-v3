package handlers

import (
	"slices"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/lib/pq"
)

func (h *Handlers) PostGroupRole(info ReqInfo, data *types.PostGroupRoleRequest) (*types.PostGroupRoleResponse, error) {
	kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(info.Ctx, info.Session.UserSub, info.Session.GroupExternalId, data.Name)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var groupRoleId string
	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (group_id, role_id) DO NOTHING
		RETURNING id
	`, info.Session.GroupId, data.RoleId, kcSubGroup.Id, time.Now(), info.Session.UserSub).Scan(&groupRoleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// update the default role
	if data.DefaultRole {
		_, err = info.Tx.Exec(info.Ctx, `
			UPDATE dbtable_schema.groups
			SET default_role_id = $2
			WHERE id = $1
		`, info.Session.GroupId, data.RoleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	newSubGroupPath := info.Session.GroupPath + "/" + data.Name
	h.Cache.SetCachedSubGroup(newSubGroupPath, kcSubGroup.Id, data.Name, info.Session.GroupPath)
	h.Cache.SetGroupSessionVersion(info.Session.GroupId)

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/roles")
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/assignments")

	return &types.PostGroupRoleResponse{GroupRoleId: groupRoleId}, nil
}

func (h *Handlers) PatchGroupRole(info ReqInfo, data *types.PatchGroupRoleRequest) (*types.PatchGroupRoleResponse, error) {
	var existingRoleExternalId, defaultRoleId, existingRoleName string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT gr.external_id, g.default_role_id, r.name
		FROM dbtable_schema.group_roles gr
		JOIN dbtable_schema.groups g ON g.id = gr.group_id
		JOIN dbtable_schema.roles r ON r.id = gr.role_id
		WHERE gr.group_id = $1 AND gr.role_id = $2
	`, info.Session.GroupId, data.RoleId).Scan(&existingRoleExternalId, &defaultRoleId, &existingRoleName)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// create new general role or get existing id
	postRoleResponse, err := h.PostRole(info, &types.PostRoleRequest{Name: data.Name})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// if attempting to change the default role while unsetting it as the default, error out
	if postRoleResponse.Id == data.RoleId && defaultRoleId == data.RoleId && data.DefaultRole == false {
		return nil, util.ErrCheck(util.UserError("Update a different role to be the default role, then modify this role."))
	}

	// update the default role
	if data.GetDefaultRole() {
		_, err = info.Tx.Exec(info.Ctx, `
			UPDATE dbtable_schema.groups
			SET default_role_id = $2
			WHERE id = $1
		`, info.Session.GroupId, postRoleResponse.GetId())
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	if postRoleResponse.GetId() != data.GetRoleId() {
		// remove the old group role entry and
		_, err := info.Tx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.group_roles
			WHERE group_id = $1 AND role_id = $2
		`, info.Session.GroupId, data.GetRoleId())
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		// create a new group_role attachment pointing to the new role id, retaining the keycloak external id
		_, err = info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (group_id, role_id) DO NOTHING
		`, info.Session.GroupId, postRoleResponse.GetId(), existingRoleExternalId, time.Now(), info.Session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		newSubGroupPath := info.Session.GroupPath + "/" + data.Name
		h.Cache.SetCachedSubGroup(newSubGroupPath, existingRoleExternalId, data.Name, info.Session.GroupPath)

		// update the name of the keycloak group which controls this role
		err = h.Keycloak.UpdateGroup(info.Ctx, info.Session.UserSub, existingRoleExternalId, data.Name)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		oldSubGroupPath := info.Session.GroupPath + "/" + existingRoleName
		h.Cache.UnsetCachedSubGroup(oldSubGroupPath)
		h.Cache.SetGroupSessionVersion(info.Session.GroupId)

		h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")
		h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/roles")
		h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/assignments")
	}

	return &types.PatchGroupRoleResponse{Success: true}, nil
}

func (h *Handlers) PatchGroupRoles(info ReqInfo, data *types.PatchGroupRolesRequest) (*types.PatchGroupRolesResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.groups
		SET default_role_id = $2, updated_sub = $3, updated_on = $4 
		WHERE id = $1
	`, info.Session.GroupId, data.GetDefaultRoleId(), info.Session.UserSub, time.Now())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	roleIds := make([]string, 0, len(data.GetRoles()))
	for roleId := range data.GetRoles() {
		roleIds = append(roleIds, roleId)
	}

	var diffs []*types.IGroupRole

	// Remove all unused roles from keycloak and the group records
	rows, err := info.Tx.Query(info.Ctx, `
		SELECT egr.id, egr."roleId", er.name
		FROM dbview_schema.enabled_group_roles egr
		JOIN dbview_schema.enabled_roles er ON er.id = egr."roleId"
		WHERE er.name != 'Admin' AND egr."groupId" = $1 AND egr."roleId" != ANY($2::uuid[])
	`, info.Session.GroupId, pq.Array(roleIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer rows.Close()

	for rows.Next() {
		var groupRoleId, roleId, roleName string
		rows.Scan(&groupRoleId, &roleId, &roleName)
		diffs = append(diffs, &types.IGroupRole{
			Id: groupRoleId,
			Role: &types.IRole{
				Id:   roleId,
				Name: roleName,
			},
		})
	}

	if len(diffs) > 0 {

		kcGroup, err := h.Keycloak.GetGroup(info.Ctx, info.Session.UserSub, info.Session.GroupExternalId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		diffIds := make([]string, 0, len(diffs))
		diffRoleNames := make([]string, 0, len(diffs))
		for _, diff := range diffs {
			diffIds = append(diffIds, diff.GetId())
			diffRoleNames = append(diffRoleNames, diff.GetRole().GetName())
		}

		if len(kcGroup.SubGroups) > 0 {
			for _, subGroup := range kcGroup.SubGroups {
				if util.StringIn(subGroup.Name, diffRoleNames) {
					err = h.Keycloak.DeleteGroup(info.Ctx, info.Session.UserSub, subGroup.Id)
					if err != nil {
						return nil, util.ErrCheck(err)
					}

					h.Cache.UnsetCachedSubGroup(subGroup.Path)
				}
			}
		}

		_, err = info.Tx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.group_roles
			WHERE id = ANY($1::uuid[])
		`, pq.Array(diffIds))

		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	// Add new roles to keycloak and group records
	for _, roleId := range roleIds {
		role := data.GetRoles()[roleId]

		if strings.ToLower(role.Name) == "admin" {
			continue
		}

		kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(info.Ctx, info.Session.UserSub, info.Session.GroupExternalId, role.Name)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4, $5::uuid)
			ON CONFLICT (group_id, role_id) DO NOTHING
		`, info.Session.GroupId, roleId, kcSubGroup.Id, time.Now(), info.Session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		newSubGroupPath := info.Session.GroupPath + "/" + role.Name
		h.Cache.SetCachedSubGroup(newSubGroupPath, kcSubGroup.Id, role.Name, info.Session.GroupPath)
	}

	h.Cache.SetGroupSessionVersion(info.Session.GroupId)
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")

	return &types.PatchGroupRolesResponse{Success: true}, nil
}

func (h *Handlers) GetGroupRoles(info ReqInfo, data *types.GetGroupRolesRequest) (*types.GetGroupRolesResponse, error) {
	roles, err := clients.QueryProtos[types.IGroupRole](info.Ctx, info.Tx, `
		SELECT 
			TO_JSONB(er) as role,
			egr.id,
			egr."roleId",
			egr."groupId",
			egr."createdOn"
		FROM dbview_schema.enabled_group_roles egr
		JOIN dbview_schema.enabled_roles er ON er.id = egr."roleId"
		WHERE egr."groupId" = $1 AND er.name != 'Admin'
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupRolesResponse{GroupRoles: roles}, nil
}

func (h *Handlers) DeleteGroupRole(info ReqInfo, data *types.DeleteGroupRoleRequest) (*types.DeleteGroupRoleResponse, error) {
	groupRoleIds := strings.Split(data.GetIds(), ",")

	// handle default role id update
	var defaultGroupRoleId string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT gr.id
		FROM dbtable_schema.group_roles gr
		JOIN dbtable_schema.groups g ON g.default_role_id = gr.role_id
		WHERE g.id = $1
	`, info.Session.GroupId).Scan(&defaultGroupRoleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if slices.Contains(groupRoleIds, defaultGroupRoleId) {
		return nil, util.ErrCheck(util.UserError("The default role may not be deleted. Update a different role to be default, then try again."))
	}

	for _, id := range groupRoleIds {
		if id != "" {
			var name string
			err := info.Tx.QueryRow(info.Ctx, `
				SELECT r.name
				FROM dbtable_schema.roles r
				JOIN dbtable_schema.group_roles gr ON gr.role_id = r.id
				WHERE gr.id = $1
			`, id).Scan(&name)

			var subGroupExternalId string
			err = info.Tx.QueryRow(info.Ctx, `
				DELETE FROM dbtable_schema.group_roles
				WHERE id = $1 AND group_id = $2
				RETURNING external_id
			`, id, info.Session.GroupId).Scan(&subGroupExternalId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			err = h.Keycloak.DeleteGroup(info.Ctx, info.Session.UserSub, subGroupExternalId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			oldSubGroupPath := info.Session.GroupPath + "/" + name
			h.Cache.UnsetCachedSubGroup(oldSubGroupPath)
		}
	}

	h.Cache.SetGroupSessionVersion(info.Session.GroupId)
	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"group/roles")

	return &types.DeleteGroupRoleResponse{Success: true}, nil
}
