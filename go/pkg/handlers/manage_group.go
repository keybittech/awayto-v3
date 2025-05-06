package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"github.com/lib/pq"
)

func (h *Handlers) PostManageGroups(info ReqInfo, data *types.PostManageGroupsRequest) (*types.PostManageGroupsResponse, error) {
	var group types.IGroup

	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.groups (name, created_on, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id, name
	`, data.GetName(), time.Now().Local().UTC(), info.Session.UserSub).Scan(&group.Id, &group.Name)
	if err != nil {
		var sqlErr *pq.Error
		if errors.As(err, &sqlErr) && sqlErr.Constraint == "unique_group_owner" {
			return nil, util.ErrCheck(util.UserError("Only 1 group can be managed at a time."))
		}
		return nil, util.ErrCheck(err)
	}

	for _, roleId := range data.GetRoles() {
		_, err := info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.uuid_roles (parent_uuid, role_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4::uuid)
			ON CONFLICT (parent_uuid, role_id) DO NOTHING
		`, group.GetId(), roleId, time.Now().Local().UTC(), info.Session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.PostManageGroupsResponse{Id: group.GetId(), Name: group.GetName(), Roles: data.GetRoles()}, nil
}

func (h *Handlers) PatchManageGroups(info ReqInfo, data *types.PatchManageGroupsRequest) (*types.PatchManageGroupsResponse, error) {
	var group types.IGroup

	// Perform the update operation
	err := info.Tx.QueryRow(info.Ctx, `
		UPDATE dbtable_schema.groups
		SET name = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
		RETURNING id, name
	`, data.GetId(), data.GetName(), info.Session.UserSub, time.Now().Local().UTC()).Scan(&group.Id, &group.Name)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Determine the existing roles
	existingRoleIDs := make(map[string]struct{})
	dbRoles := []*types.IGroupRole{}

	err = h.Database.QueryRows(info.Ctx, info.Tx, &dbRoles, `
		SELECT id, role_id as "roleId" FROM uuid_roles WHERE parent_uuid = $1
	`, group.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, role := range dbRoles {
		existingRoleIDs[role.RoleId] = struct{}{}
	}

	// Deleting unneeded roles
	for _, role := range dbRoles {
		if _, exists := data.GetRoles()[role.GetRoleId()]; !exists {
			_, err = info.Tx.Exec(info.Ctx, `DELETE FROM uuid_roles WHERE id = $1`, role.GetId())
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	// Inserting new roles
	for roleId := range data.GetRoles() {
		if _, exists := existingRoleIDs[roleId]; !exists {
			_, err := info.Tx.Exec(info.Ctx, `
				INSERT INTO dbtable_schema.uuid_roles (parent_uuid, role_id, created_on, created_sub)
				VALUES ($1, $2, $3, $4::uuid)
				ON CONFLICT (parent_uuid, role_id) DO NOTHING
			`, group.Id, roleId, time.Now().Local().UTC(), info.Session.UserSub)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	return &types.PatchManageGroupsResponse{Success: true}, nil
}

func (h *Handlers) GetManageGroups(info ReqInfo, data *types.GetManageGroupsRequest) (*types.GetManageGroupsResponse, error) {
	groups := []*types.IGroup{}

	err := h.Database.QueryRows(info.Ctx, info.Tx, &groups, `
		SELECT * FROM dbview_schema.enabled_groups_ext
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetManageGroupsResponse{Groups: groups}, nil
}

func (h *Handlers) DeleteManageGroups(info ReqInfo, data *types.DeleteManageGroupsRequest) (*types.DeleteManageGroupsResponse, error) {
	for _, id := range strings.Split(data.GetIds(), ",") {
		_, err := info.Tx.Exec(info.Ctx, `
			DELETE FROM dbtable_schema.groups
			WHERE id = $1
		`, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.DeleteManageGroupsResponse{Success: true}, nil
}
