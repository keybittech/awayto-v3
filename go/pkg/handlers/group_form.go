package handlers

import (
	"errors"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupForm(info ReqInfo, data *types.PostGroupFormRequest) (*types.PostGroupFormResponse, error) {
	var groupFormExists string

	err := info.Tx.QueryRow(info.Req.Context(), `
		SELECT f.id FROM dbtable_schema.forms f
		LEFT JOIN dbtable_schema.group_forms gf ON gf.form_id = f.id
		WHERE f.name = $1 AND gf.group_id = $2
	`, data.GetName(), info.Session.GroupId).Scan(&groupFormExists)
	if groupFormExists != "" && err != nil {
		return nil, util.ErrCheck(err)
	}
	if groupFormExists != "" {
		return nil, util.ErrCheck(util.UserError("A form with this name already exists."))
	}

	groupPoolTx, groupSession, err := h.Database.DatabaseClient.OpenPoolSessionGroupTx(info.Req.Context(), info.Session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer groupPoolTx.Tx.Rollback(info.Req.Context())

	groupInfo := ReqInfo{
		W:       info.W,
		Req:     info.Req,
		Session: groupSession,
		Tx:      groupPoolTx,
	}

	formResp, err := h.PostForm(groupInfo, &types.PostFormRequest{Form: data.GetGroupForm().GetForm()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Database.DatabaseClient.ClosePoolSessionTx(info.Req.Context(), groupPoolTx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	groupFormId := formResp.GetId()

	_, err = h.PostFormVersion(info, &types.PostFormVersionRequest{
		Name: data.GetGroupForm().GetForm().GetName(),
		Version: &types.IProtoFormVersion{
			FormId: groupFormId,
			Form:   data.GetGroupForm().GetForm().GetVersion().GetForm(),
		},
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Req.Context(), `
		INSERT INTO dbtable_schema.group_forms (group_id, form_id, created_sub)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (group_id, form_id) DO NOTHING
	`, info.Session.GroupId, groupFormId, info.Session.GroupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/forms")

	return &types.PostGroupFormResponse{Id: groupFormId}, nil
}

func (h *Handlers) PostGroupFormVersion(info ReqInfo, data *types.PostGroupFormVersionRequest) (*types.PostGroupFormVersionResponse, error) {
	formVersionResp, err := h.PostFormVersion(info, &types.PostFormVersionRequest{Name: data.GetName(), Version: data.GetGroupFormVersion()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/forms")
	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/forms/"+data.FormId)

	return &types.PostGroupFormVersionResponse{Id: formVersionResp.GetId()}, nil
}

func (h *Handlers) PatchGroupForm(info ReqInfo, data *types.PatchGroupFormRequest) (*types.PatchGroupFormResponse, error) {
	_, err := h.PatchForm(info, &types.PatchFormRequest{Form: data.GetGroupForm().GetForm()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/forms")
	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/forms/"+data.GetGroupForm().GetForm().GetId())

	return &types.PatchGroupFormResponse{Success: true}, nil
}

func (h *Handlers) GetGroupForms(info ReqInfo, data *types.GetGroupFormsRequest) (*types.GetGroupFormsResponse, error) {
	var forms []*types.IProtoForm

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &forms, `
		SELECT es.*
		FROM dbview_schema.enabled_group_forms eus
		LEFT JOIN dbview_schema.enabled_forms es ON es.id = eus."formId"
		WHERE eus."groupId" = $1
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var groupForms []*types.IGroupForm
	for _, f := range forms {
		groupForms = append(groupForms, &types.IGroupForm{FormId: f.GetId(), Form: f})
	}

	return &types.GetGroupFormsResponse{GroupForms: groupForms}, nil
}

func (h *Handlers) GetGroupFormById(info ReqInfo, data *types.GetGroupFormByIdRequest) (*types.GetGroupFormByIdResponse, error) {
	var groupForms []*types.IGroupForm

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &groupForms, `
		SELECT egfe.*
		FROM dbview_schema.enabled_group_forms_ext egfe
		WHERE egfe."groupId" = $1 and egfe."formId" = $2
	`, info.Session.GroupId, data.GetFormId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupForms) == 0 {
		return nil, util.ErrCheck(errors.New("group form not found"))
	}

	return &types.GetGroupFormByIdResponse{GroupForm: groupForms[0]}, nil
}

func (h *Handlers) DeleteGroupForm(info ReqInfo, data *types.DeleteGroupFormRequest) (*types.DeleteGroupFormResponse, error) {

	for _, formId := range strings.Split(data.GetIds(), ",") {
		_, err := info.Tx.Exec(info.Req.Context(), `DELETE FROM dbtable_schema.forms WHERE id = $1`, formId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/forms")

	return &types.DeleteGroupFormResponse{Success: true}, nil
}
