package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostForm(info ReqInfo, data *types.PostFormRequest) (*types.PostFormResponse, error) {
	formInsert := util.BatchQueryRow[types.ILookup](info.Batch, `
		INSERT INTO dbtable_schema.forms (name, created_on, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, data.Form.Name, time.Now(), info.Session.UserSub)

	info.Batch.Send(info.Ctx)

	return &types.PostFormResponse{Id: (*formInsert).Id}, nil
}

func (h *Handlers) PostFormVersion(info ReqInfo, data *types.PostFormVersionRequest) (*types.PostFormVersionResponse, error) {
	formJson, err := data.Version.Form.MarshalJSON()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	versionInsert := util.BatchQueryRow[types.ILookup](info.Batch, `
		INSERT INTO dbtable_schema.form_versions (form_id, form, created_on, created_sub)
		VALUES ($1::uuid, $2::jsonb, $3, $4::uuid)
		RETURNING id
	`, data.Version.FormId, formJson, time.Now(), info.Session.UserSub)

	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.Name, time.Now(), info.Session.UserSub, data.Version.FormId)

	info.Batch.Send(info.Ctx)

	return &types.PostFormVersionResponse{Id: (*versionInsert).Id}, nil
}

func (h *Handlers) PatchForm(info ReqInfo, data *types.PatchFormRequest) (*types.PatchFormResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.Form.Name, time.Now(), info.Session.UserSub, data.Form.Id)

	info.Batch.Send(info.Ctx)

	return &types.PatchFormResponse{Success: true}, nil
}

func (h *Handlers) GetForms(info ReqInfo, data *types.GetFormsRequest) (*types.GetFormsResponse, error) {
	forms := util.BatchQuery[types.IProtoForm](info.Batch, `
		SELECT id, name, "createdOn"
		FROM dbview_schema.enabled_forms
		WHERE "createdSub" = $1
	`, info.Session.UserSub)

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
	`, data.Id, time.Now(), info.Session.UserSub)

	info.Batch.Send(info.Ctx)

	return &types.DisableFormResponse{Success: true}, nil
}
