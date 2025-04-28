package handlers

import (
	"errors"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostManageRoles(info ReqInfo, data *types.PostManageRolesRequest) (*types.PostManageRolesResponse, error) {
	var roles []*types.IRole

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &roles, `
		INSERT INTO dbtable_schema.roles (name, created_sub)
		VALUES ($1, $2::uuid)
		RETURNING id, name
	`, data.GetName(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(roles) == 0 {
		return nil, util.ErrCheck(errors.New("failed to create role"))
	}

	return &types.PostManageRolesResponse{Id: roles[0].GetId(), Name: roles[0].GetName()}, nil
}

func (h *Handlers) PatchManageRoles(info ReqInfo, data *types.PatchManageRolesRequest) (*types.PatchManageRolesResponse, error) {
	var roles []*types.IRole

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &roles, `
		UPDATE dbtable_schema.roles
		SET name = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
		RETURNING id, name
	`, data.GetId(), data.GetName(), info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(roles) == 0 {
		return nil, util.ErrCheck(errors.New("no role found during update"))
	}

	return &types.PatchManageRolesResponse{Success: true}, nil
}

func (h *Handlers) GetManageRoles(info ReqInfo, data *types.GetManageRolesRequest) (*types.GetManageRolesResponse, error) {
	var roles []*types.IRole

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &roles, `
		SELECT * FROM dbview_schema.enabled_roles
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetManageRolesResponse{Roles: roles}, nil
}

func (h *Handlers) DeleteManageRoles(info ReqInfo, data *types.DeleteManageRolesRequest) (*types.DeleteManageRolesResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		DELETE FROM dbtable_schema.roles
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteManageRolesResponse{Success: true}, nil
}
