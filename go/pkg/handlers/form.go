package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostForm(info ReqInfo, data *types.PostFormRequest) (*types.PostFormResponse, error) {
	var formId string
	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.forms (name, created_sub)
		VALUES ($1, $2::uuid)
		RETURNING id
	`, data.Form.Name, info.Session.GetUserSub()).Scan(&formId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormResponse{Id: formId}, nil
}

func (h *Handlers) PostFormVersion(info ReqInfo, data *types.PostFormVersionRequest) (*types.PostFormVersionResponse, error) {
	userSub := info.Session.GetUserSub()

	formJson, err := data.Version.Form.MarshalJSON()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var versionId string
	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.form_versions (form_id, form, created_sub)
		VALUES ($1::uuid, $2::jsonb, $3::uuid)
		RETURNING id
	`, data.Version.FormId, formJson, userSub).Scan(&versionId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.Name, time.Now(), userSub, data.Version.FormId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormVersionResponse{Id: versionId}, nil
}

func (h *Handlers) PatchForm(info ReqInfo, data *types.PatchFormRequest) (*types.PatchFormResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.Form.Name, time.Now(), info.Session.GetUserSub(), data.Form.Id)

	info.Batch.Send(info.Ctx)

	return &types.PatchFormResponse{Success: true}, nil
}

func (h *Handlers) GetForms(info ReqInfo, data *types.GetFormsRequest) (*types.GetFormsResponse, error) {
	forms := util.BatchQuery[types.IProtoForm](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_forms
		WHERE "createdSub" = $1
	`, info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.GetFormsResponse{Forms: *forms}, nil
}

func (h *Handlers) GetFormById(info ReqInfo, data *types.GetFormByIdRequest) (*types.GetFormByIdResponse, error) {
	form := util.BatchQueryRow[types.IProtoForm](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_forms
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.GetFormByIdResponse{Form: *form}, nil
}

func (h *Handlers) DeleteForm(info ReqInfo, data *types.DeleteFormRequest) (*types.DeleteFormResponse, error) {
	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.forms
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.DeleteFormResponse{Success: true}, nil
}

func (h *Handlers) DisableForm(info ReqInfo, data *types.DisableFormRequest) (*types.DisableFormResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.forms
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.Id, time.Now(), info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.DisableFormResponse{Success: true}, nil
}
