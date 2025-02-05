package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"database/sql"
	"net/http"
	"strings"
)

func (h *Handlers) PostRole(w http.ResponseWriter, req *http.Request, data *types.PostRoleRequest) (*types.PostRoleResponse, error) {
	session := h.Redis.ReqSession(req)

	var role types.IRole
	err := h.Database.Client().QueryRow(`
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

	var userId sql.NullString
	err = h.Database.Client().QueryRow(`
		SELECT id FROM dbtable_schema.users WHERE sub = $1
	`, session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.Database.Client().Exec(`
		INSERT INTO dbtable_schema.user_roles (user_id, role_id, created_sub)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userId, role.Id, session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	return &types.PostRoleResponse{Id: role.Id}, nil
}

func (h *Handlers) GetRoles(w http.ResponseWriter, req *http.Request, data *types.GetRolesRequest) (*types.GetRolesResponse, error) {
	session := h.Redis.ReqSession(req)
	var roles []*types.IRole
	err := h.Database.QueryRows(&roles, `
		SELECT eur.id, er.name, eur."createdOn" 
		FROM dbview_schema.enabled_roles er
		LEFT JOIN dbview_schema.enabled_user_roles eur ON er.id = eur."roleId"
		LEFT JOIN dbview_schema.enabled_users eu ON eu.id = eur."userId"
		WHERE eu.sub = $1
	`, session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetRolesResponse{Roles: roles}, nil
}

func (h *Handlers) GetRoleById(w http.ResponseWriter, req *http.Request, data *types.GetRoleByIdRequest) (*types.GetRoleByIdResponse, error) {
	var role types.IRole
	err := h.Database.Client().QueryRow(`
		SELECT * FROM dbview_schema.enabled_roles
		WHERE id = $1
	`, data.GetId()).Scan(&role.Id, &role.Name)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetRoleByIdResponse{Role: &role}, nil
}

func (h *Handlers) DeleteRole(w http.ResponseWriter, req *http.Request, data *types.DeleteRoleRequest) (*types.DeleteRoleResponse, error) {
	session := h.Redis.ReqSession(req)
	var userId string
	err := h.Database.Client().QueryRow(`
		SELECT id FROM dbtable_schema.users WHERE sub = $1
	`, session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	ids := data.GetIds()
	for _, id := range strings.Split(ids, ",") {
		if id != "" {
			_, err = h.Database.Client().Exec(`
			DELETE FROM dbtable_schema.user_roles
			WHERE role_id = $1 AND user_id = $2
		`, id, userId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	return &types.DeleteRoleResponse{Success: true}, nil
}
