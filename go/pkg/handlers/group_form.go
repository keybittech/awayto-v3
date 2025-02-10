package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"errors"
	"net/http"
	"strings"
)

func (h *Handlers) PostGroupForm(w http.ResponseWriter, req *http.Request, data *types.PostGroupFormRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostGroupFormResponse, error) {
	var groupFormExists string

	err := tx.QueryRow(`
		SELECT f.id FROM dbtable_schema.forms f
		LEFT JOIN dbtable_schema.group_forms gf ON gf.form_id = f.id
		WHERE f.name = $1 AND gf.group_id = $2
	`, data.GetName(), session.GroupId).Scan(&groupFormExists)
	if groupFormExists != "" && err != nil {
		return nil, util.ErrCheck(err)
	}
	if groupFormExists != "" {
		return nil, util.ErrCheck(util.UserError("A form with this name already exists."))
	}

	err = tx.SetDbVar("user_sub", session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	formResp, err := h.PostForm(w, req, &types.PostFormRequest{Form: data.GetGroupForm().GetForm()}, &clients.UserSession{UserSub: session.GroupSub}, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.SetDbVar("user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	groupFormId := formResp.GetId()

	_, err = h.PostFormVersion(w, req, &types.PostFormVersionRequest{
		Name: data.GetGroupForm().GetForm().GetName(),
		Version: &types.IProtoFormVersion{
			FormId: groupFormId,
			Form:   data.GetGroupForm().GetForm().GetVersion().GetForm(),
		},
	}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		INSERT INTO dbtable_schema.group_forms (group_id, form_id, created_sub)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (group_id, form_id) DO NOTHING
	`, session.GroupId, groupFormId, session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/forms")

	return &types.PostGroupFormResponse{Id: groupFormId}, nil
}

func (h *Handlers) PostGroupFormVersion(w http.ResponseWriter, req *http.Request, data *types.PostGroupFormVersionRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostGroupFormVersionResponse, error) {
	formVersionResp, err := h.PostFormVersion(w, req, &types.PostFormVersionRequest{Name: data.GetName(), Version: data.GetGroupFormVersion()}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/forms")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/forms/"+data.FormId)

	return &types.PostGroupFormVersionResponse{Id: formVersionResp.GetId()}, nil
}

func (h *Handlers) PatchGroupForm(w http.ResponseWriter, req *http.Request, data *types.PatchGroupFormRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchGroupFormResponse, error) {
	_, err := h.PatchForm(w, req, &types.PatchFormRequest{Form: data.GetGroupForm().GetForm()}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/forms")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/forms/"+data.GetGroupForm().GetForm().GetId())

	return &types.PatchGroupFormResponse{Success: true}, nil
}

func (h *Handlers) GetGroupForms(w http.ResponseWriter, req *http.Request, data *types.GetGroupFormsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetGroupFormsResponse, error) {
	var forms []*types.IProtoForm

	err := tx.QueryRows(&forms, `
		SELECT es.*
		FROM dbview_schema.enabled_group_forms eus
		LEFT JOIN dbview_schema.enabled_forms es ON es.id = eus."formId"
		WHERE eus."groupId" = $1
	`, session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var groupForms []*types.IGroupForm
	for _, f := range forms {
		groupForms = append(groupForms, &types.IGroupForm{FormId: f.GetId(), Form: f})
	}

	return &types.GetGroupFormsResponse{GroupForms: groupForms}, nil
}

func (h *Handlers) GetGroupFormById(w http.ResponseWriter, req *http.Request, data *types.GetGroupFormByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetGroupFormByIdResponse, error) {
	var groupForms []*types.IGroupForm

	err := tx.QueryRows(&groupForms, `
		SELECT egfe.*
		FROM dbview_schema.enabled_group_forms_ext egfe
		WHERE egfe."groupId" = $1 and egfe."formId" = $2
	`, session.GroupId, data.GetFormId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupForms) == 0 {
		return nil, util.ErrCheck(errors.New("group form not found"))
	}

	return &types.GetGroupFormByIdResponse{GroupForm: groupForms[0]}, nil
}

func (h *Handlers) DeleteGroupForm(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupFormRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteGroupFormResponse, error) {

	for _, formId := range strings.Split(data.GetIds(), ",") {
		_, err := tx.Exec(`DELETE FROM dbtable_schema.forms WHERE id = $1`, formId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/forms")

	return &types.DeleteGroupFormResponse{Success: true}, nil
}
