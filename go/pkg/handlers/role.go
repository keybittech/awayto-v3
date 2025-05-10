package handlers

import (
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostRole(info ReqInfo, data *types.PostRoleRequest) (*types.PostRoleResponse, error) {
	var role types.IRole
	err := info.Tx.QueryRow(info.Ctx, `
		WITH input_rows(name, created_sub) as (VALUES ($1, $2::uuid)), ins AS (
			INSERT INTO dbtable_schema.roles (name, created_sub)
			SELECT * FROM input_rows
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
	`, info.Session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.user_roles (user_id, role_id, created_sub)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userId, role.Id, info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")

	return &types.PostRoleResponse{Id: role.Id}, nil
}

func (h *Handlers) GetRoles(info ReqInfo, data *types.GetRolesRequest) (*types.GetRolesResponse, error) {
	roles, err := clients.QueryProtos[types.IRole](info.Ctx, info.Tx, `
		SELECT er.*
		FROM dbview_schema.enabled_roles er
		LEFT JOIN dbview_schema.enabled_user_roles eur ON er.id = eur."roleId"
		LEFT JOIN dbview_schema.enabled_users eu ON eu.id = eur."userId"
		WHERE eu.sub = $1
	`, info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetRolesResponse{Roles: roles}, nil
}

func (h *Handlers) GetRoleById(info ReqInfo, data *types.GetRoleByIdRequest) (*types.GetRoleByIdResponse, error) {
	role, err := clients.QueryProto[types.IRole](info.Ctx, info.Tx, `
		SELECT *
		FROM dbview_schema.enabled_roles
		WHERE id = $1
	`, data.Id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetRoleByIdResponse{Role: role}, nil
}

func (h *Handlers) DeleteRole(info ReqInfo, data *types.DeleteRoleRequest) (*types.DeleteRoleResponse, error) {
	var userId string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT id FROM dbtable_schema.users WHERE sub = $1
	`, info.Session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	ids := data.GetIds()
	for _, id := range strings.Split(ids, ",") {
		if id != "" {
			_, err = info.Tx.Exec(info.Ctx, `
				DELETE FROM dbtable_schema.user_roles
				WHERE role_id = $1 AND user_id = $2
			`, id, userId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")

	return &types.DeleteRoleResponse{Success: true}, nil
}
