package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"errors"
	"net/http"
	"time"
)

func (h *Handlers) PostManageRoles(w http.ResponseWriter, req *http.Request, data *types.PostManageRolesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostManageRolesResponse, error) {
	var roles []*types.IRole

	err := tx.QueryRows(&roles, `
		INSERT INTO dbtable_schema.roles (name, created_sub)
		VALUES ($1, $2::uuid)
		RETURNING id, name
	`, data.GetName(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(roles) == 0 {
		return nil, util.ErrCheck(errors.New("failed to create role"))
	}

	return &types.PostManageRolesResponse{Id: roles[0].GetId(), Name: roles[0].GetName()}, nil
}

func (h *Handlers) PatchManageRoles(w http.ResponseWriter, req *http.Request, data *types.PatchManageRolesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchManageRolesResponse, error) {
	var roles []*types.IRole

	err := tx.QueryRows(&roles, `
		UPDATE dbtable_schema.roles
		SET name = $2, updated_sub = $3, updated_on = $4
		WHERE id = $1
		RETURNING id, name
	`, data.GetId(), data.GetName(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(roles) == 0 {
		return nil, util.ErrCheck(errors.New("no role found during update"))
	}

	return &types.PatchManageRolesResponse{Success: true}, nil
}

func (h *Handlers) GetManageRoles(w http.ResponseWriter, req *http.Request, data *types.GetManageRolesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetManageRolesResponse, error) {
	var roles []*types.IRole

	err := tx.QueryRows(&roles, `
		SELECT * FROM dbview_schema.enabled_roles
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetManageRolesResponse{Roles: roles}, nil
}

func (h *Handlers) DeleteManageRoles(w http.ResponseWriter, req *http.Request, data *types.DeleteManageRolesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteManageRolesResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.roles
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteManageRolesResponse{Success: true}, nil
}
