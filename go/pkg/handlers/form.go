package handlers

import (
	"database/sql"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostForm(info ReqInfo, data *types.PostFormRequest) (*types.PostFormResponse, error) {
	var id string
	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.forms (name, created_on, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, data.GetForm().GetName(), time.Now().Local().UTC(), info.Session.UserSub).Scan(&id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormResponse{Id: id}, nil
}

func (h *Handlers) PostFormVersion(info ReqInfo, data *types.PostFormVersionRequest) (*types.PostFormVersionResponse, error) {
	formJson, err := data.GetVersion().GetForm().MarshalJSON()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var versionId string
	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.form_versions (form_id, form, created_on, created_sub)
		VALUES ($1::uuid, $2::jsonb, $3, $4::uuid)
		RETURNING id
	`, data.GetVersion().GetFormId(), formJson, time.Now().Local().UTC(), info.Session.UserSub).Scan(&versionId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.GetName(), time.Now().Local().UTC(), info.Session.UserSub, data.GetVersion().GetFormId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormVersionResponse{Id: versionId}, nil
}

func (h *Handlers) PatchForm(info ReqInfo, data *types.PatchFormRequest) (*types.PatchFormResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.GetForm().GetName(), time.Now().Local().UTC(), info.Session.UserSub, data.GetForm().GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchFormResponse{Success: true}, nil
}

func (h *Handlers) GetForms(info ReqInfo, data *types.GetFormsRequest) (*types.GetFormsResponse, error) {
	var forms []*types.IProtoForm

	err := h.Database.QueryRows(info.Ctx, info.Tx, &forms, `
		SELECT * FROM dbview_schema.enabled_forms
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetFormsResponse{Forms: forms}, nil
}

func (h *Handlers) GetFormById(info ReqInfo, data *types.GetFormByIdRequest) (*types.GetFormByIdResponse, error) {
	var forms []*types.IProtoForm

	err := h.Database.QueryRows(info.Ctx, info.Tx, &forms, `
		SELECT * FROM dbview_schema.enabled_forms
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(forms) == 0 {
		return nil, sql.ErrNoRows
	}

	return &types.GetFormByIdResponse{Form: forms[0]}, nil
}

func (h *Handlers) DeleteForm(info ReqInfo, data *types.DeleteFormRequest) (*types.DeleteFormResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.forms
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteFormResponse{Success: true}, nil
}

func (h *Handlers) DisableForm(info ReqInfo, data *types.DisableFormRequest) (*types.DisableFormResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.forms
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableFormResponse{Success: true}, nil
}
