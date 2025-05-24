package handlers

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/lib/pq"
)

func postRole(tx *clients.PoolTx, ctx context.Context, name, adminSub string) (string, error) {
	var roleId string
	err := tx.QueryRow(ctx, `
		WITH input_rows(name, created_sub) as (VALUES ($1, $2::uuid)), ins AS (
			INSERT INTO dbtable_schema.roles (name, created_sub)
			SELECT name, created_sub FROM input_rows
			ON CONFLICT (name) DO NOTHING
			RETURNING id
		)
		SELECT id
		FROM ins
		UNION ALL
		SELECT s.id
		FROM input_rows
		JOIN dbtable_schema.roles s USING (name);
	`, name, adminSub).Scan(&roleId)
	if err != nil {
		return "", err
	}
	return roleId, nil
}

func (h *Handlers) PostGroupRole(info ReqInfo, data *types.PostGroupRoleRequest) (*types.PostGroupRoleResponse, error) {
	userSub := info.Session.GetUserSub()
	groupId := info.Session.GetGroupId()

	roleId, err := postRole(info.Tx, info.Ctx, data.GetName(), h.Database.AdminSub())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var existingRoleExternalId string
	err = info.Tx.QueryRow(info.Ctx, `
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE role_id = $1
	`, roleId).Scan(&existingRoleExternalId)

	var undos []func()

	defer func() {
		if undos != nil && len(undos) > 0 {
			for _, undo := range undos {
				undo()
			}
		}
	}()

	kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(info.Ctx, userSub, info.Session.GetGroupExternalId(), data.GetName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	if kcSubGroup.Id == "" {
		return nil, util.ErrCheck(errors.New("error creating keycloak role subgroup"))
	}
	if existingRoleExternalId == kcSubGroup.Id {
		return nil, util.ErrCheck(util.UserError("The group already has that role."))
	}

	undos = append(undos, func() {
		err = h.Keycloak.DeleteGroup(info.Ctx, userSub, kcSubGroup.GetId())
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
		}
	})

	var groupRoleId string
	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (group_id, role_id) DO NOTHING
		RETURNING id
	`, groupId, roleId, kcSubGroup.Id, time.Now(), userSub).Scan(&groupRoleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// update the default role
	if data.DefaultRole {
		_, err = info.Tx.Exec(info.Ctx, `
			UPDATE dbtable_schema.groups
			SET default_role_id = $2
			WHERE id = $1
		`, groupId, roleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	undos = nil
	return &types.PostGroupRoleResponse{GroupRoleId: groupRoleId, RoleId: roleId}, nil
}

func (h *Handlers) PatchGroupRole(info ReqInfo, data *types.PatchGroupRoleRequest) (*types.PatchGroupRoleResponse, error) {
	userSub := info.Session.GetUserSub()
	groupId := info.Session.GetGroupId()

	var existingRoleExternalId, defaultRoleId, existingRoleName string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT gr.external_id, g.default_role_id, r.name
		FROM dbtable_schema.group_roles gr
		JOIN dbtable_schema.groups g ON g.id = gr.group_id
		JOIN dbtable_schema.roles r ON r.id = gr.role_id
		WHERE gr.group_id = $1 AND gr.role_id = $2
	`, groupId, data.GetRoleId()).Scan(&existingRoleExternalId, &defaultRoleId, &existingRoleName)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// create new general role or get existing id
	roleId, err := postRole(info.Tx, info.Ctx, data.Name, h.Database.AdminSub())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// if attempting to change the default role while unsetting it as the default, error out
	if roleId == data.GetRoleId() && defaultRoleId == data.GetRoleId() && data.DefaultRole == false {
		return nil, util.ErrCheck(util.UserError("Update a different role to be the default role, then modify this role."))
	}

	// update the default role
	if data.GetDefaultRole() {
		_, err = info.Tx.Exec(info.Ctx, `
			UPDATE dbtable_schema.groups
			SET default_role_id = $2
			WHERE created_sub = $1
		`, userSub, roleId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	if roleId != data.GetRoleId() {
		// remove the old group role entry and
		_, err := info.Tx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.group_roles
			WHERE group_id = $1 AND role_id = $2
		`, groupId, data.GetRoleId())
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		// create a new group_role attachment pointing to the new role id, retaining the keycloak external id
		_, err = info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (group_id, role_id) DO NOTHING
		`, groupId, roleId, existingRoleExternalId, time.Now(), userSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		// update the name of the keycloak group which controls this role
		err = h.Keycloak.UpdateGroup(info.Ctx, userSub, existingRoleExternalId, data.GetName())
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		h.Cache.SubGroups.Delete(info.Session.GetSubGroupPath())
	}

	return &types.PatchGroupRoleResponse{Success: true}, nil
}

func (h *Handlers) PatchGroupRoles(info ReqInfo, data *types.PatchGroupRolesRequest) (*types.PatchGroupRolesResponse, error) {
	userSub := info.Session.GetUserSub()
	groupId := info.Session.GetGroupId()
	groupExternalId := info.Session.GetGroupExternalId()

	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.groups
		SET default_role_id = $2, updated_sub = $3, updated_on = $4 
		WHERE created_sub = $1
	`, userSub, data.GetDefaultRoleId(), userSub, time.Now())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	dataRoles := data.GetRoles()
	if len(dataRoles) == 0 {
		return &types.PatchGroupRolesResponse{Success: true}, nil
	}

	roleIds := make([]string, 0, len(dataRoles))
	for _, role := range data.GetRoles() {
		roleIds = append(roleIds, role.GetRoleId())
	}

	var diffs []*types.IGroupRole

	// Remove all unused roles from keycloak and the group records
	rows, err := info.Tx.Query(info.Ctx, `
		SELECT egr.id, egr."roleId", er.name
		FROM dbview_schema.enabled_group_roles egr
		JOIN dbview_schema.enabled_roles er ON er.id = egr."roleId"
		WHERE er.name != 'Admin' AND egr."groupId" = $1 AND egr."roleId" != ANY($2::uuid[])
	`, groupId, pq.Array(roleIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer rows.Close()

	for rows.Next() {
		var groupRoleId, roleId, roleName string
		err = rows.Scan(&groupRoleId, &roleId, &roleName)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			continue
		}
		diffs = append(diffs, &types.IGroupRole{
			Id:     groupRoleId,
			RoleId: roleId,
			Name:   roleName,
		})
	}

	if len(diffs) > 0 {

		kcGroup, err := h.Keycloak.GetGroup(info.Ctx, userSub, groupExternalId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		diffIds := make([]string, 0, len(diffs))
		diffRoleNames := make([]string, 0, len(diffs))
		for _, diff := range diffs {
			diffIds = append(diffIds, diff.GetId())
			diffRoleNames = append(diffRoleNames, diff.GetName())
		}

		if len(kcGroup.SubGroups) > 0 {
			for _, subGroup := range kcGroup.SubGroups {
				if slices.Contains(diffRoleNames, subGroup.Name) {
					err = h.Keycloak.DeleteGroup(info.Ctx, userSub, subGroup.Id)
					if err != nil {
						return nil, util.ErrCheck(err)
					}

					h.Cache.SubGroups.Delete(subGroup.Path)
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

	var undos []func()

	defer func() {
		if undos != nil && len(undos) > 0 {
			for _, undo := range undos {
				undo()
			}
		}
	}()

	// Add new roles to keycloak and group records
	for _, role := range dataRoles {

		if strings.ToLower(role.GetName()) == "admin" {
			continue
		}

		kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(info.Ctx, userSub, groupExternalId, role.GetName())
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		undos = append(undos, func() {
			err = h.Keycloak.DeleteGroup(info.Ctx, userSub, kcSubGroup.GetId())
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
			}
		})

		_, err = info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4, $5::uuid)
			ON CONFLICT (group_id, role_id) DO NOTHING
		`, groupId, role.GetRoleId(), kcSubGroup.Id, time.Now(), userSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	undos = nil
	return &types.PatchGroupRolesResponse{Success: true}, nil
}

func (h *Handlers) GetGroupRoles(info ReqInfo, data *types.GetGroupRolesRequest) (*types.GetGroupRolesResponse, error) {
	groupId := info.Session.GetGroupId()
	if groupId == "" {
		return nil, nil
	}

	roles := util.BatchQuery[types.IGroupRole](info.Batch, `
		SELECT egr.id, er.name, egr."roleId", egr."createdOn"
		FROM dbview_schema.enabled_group_roles egr
		JOIN dbview_schema.enabled_roles er ON er.id = egr."roleId"
		WHERE egr."groupId" = $1 AND er.name != 'Admin'
	`, groupId)

	info.Batch.Send(info.Ctx)

	return &types.GetGroupRolesResponse{GroupRoles: *roles}, nil
}

func (h *Handlers) DeleteGroupRole(info ReqInfo, data *types.DeleteGroupRoleRequest) (*types.DeleteGroupRoleResponse, error) {
	groupRoleIds := strings.Split(data.GetIds(), ",")

	userSub := info.Session.GetUserSub()
	groupId := info.Session.GetGroupId()

	// handle default role id update
	var defaultGroupRoleId string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT gr.id
		FROM dbtable_schema.groups g
		JOIN dbtable_schema.group_roles gr ON gr.role_id = g.default_role_id
		WHERE g.created_sub = $1
	`, userSub).Scan(&defaultGroupRoleId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
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
			`, id, groupId).Scan(&subGroupExternalId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			err = h.Keycloak.DeleteGroup(info.Ctx, userSub, subGroupExternalId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			h.Cache.SubGroups.Delete(info.Session.GetGroupPath() + "/" + name)
		}
	}

	return &types.DeleteGroupRoleResponse{Success: true}, nil
}
