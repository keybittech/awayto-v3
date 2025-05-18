package handlers

import (
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"
)

func (h *Handlers) PostRole(info ReqInfo, data *types.PostRoleRequest) (*types.PostRoleResponse, error) {
	var role types.IRole
	err := info.Tx.QueryRow(info.Ctx, `
		WITH input_rows(name, created_sub) as (VALUES ($1, $2::uuid)), ins AS (
			INSERT INTO dbtable_schema.roles (name, created_sub)
			SELECT name, created_sub FROM input_rows
			ON CONFLICT (name) DO NOTHING
			RETURNING id, name
		)
		SELECT id, name
		FROM ins
		UNION ALL
		SELECT s.id, s.name
		FROM input_rows
		JOIN dbtable_schema.roles s USING (name);
	`, data.GetName(), h.Database.AdminSub()).Scan(&role.Id, &role.Name)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var userId string
	err = info.Tx.QueryRow(info.Ctx, `
		SELECT id FROM dbtable_schema.users WHERE sub = $1
	`, info.Session.GetUserSub()).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.user_roles (user_id, role_id, created_sub)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userId, role.Id, info.Session.GetUserSub())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"profile/details")

	return &types.PostRoleResponse{Id: role.Id}, nil
}

func (h *Handlers) GetRoles(info ReqInfo, data *types.GetRolesRequest) (*types.GetRolesResponse, error) {
	roles := util.BatchQuery[types.IRole](info.Batch, `
		SELECT er.id, er.name, er."createdOn"
		FROM dbview_schema.enabled_roles er
		LEFT JOIN dbview_schema.enabled_user_roles eur ON er.id = eur."roleId"
		LEFT JOIN dbview_schema.enabled_users eu ON eu.id = eur."userId"
		WHERE eu.sub = $1
	`, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.GetRolesResponse{Roles: *roles}, nil
}

func (h *Handlers) GetRoleById(info ReqInfo, data *types.GetRoleByIdRequest) (*types.GetRoleByIdResponse, error) {
	role := util.BatchQueryRow[types.IRole](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_roles
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.GetRoleByIdResponse{Role: *role}, nil
}

func (h *Handlers) DeleteRole(info ReqInfo, data *types.DeleteRoleRequest) (*types.DeleteRoleResponse, error) {
	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.user_roles
		WHERE role_id = ANY($1) AND created_sub = $2
	`, pq.Array(strings.Split(data.Ids, ",")), info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"profile/details")

	return &types.DeleteRoleResponse{Success: true}, nil
}
