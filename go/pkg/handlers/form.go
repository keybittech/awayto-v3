package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"database/sql"
	"net/http"
	"time"
)

func (h *Handlers) PostForm(w http.ResponseWriter, req *http.Request, data *types.PostFormRequest) (*types.PostFormResponse, error) {
	session := h.Redis.ReqSession(req)

	userSub := req.Context().Value("UserSub")
	if userSub == "" {
		userSub = session.UserSub
	}

	var id string
	err := h.Database.Client().QueryRow(`
		INSERT INTO dbtable_schema.forms (name, created_on, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, data.GetForm().GetName(), time.Now().Local().UTC(), userSub).Scan(&id)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormResponse{Id: id}, nil
}

func (h *Handlers) PostFormVersion(w http.ResponseWriter, req *http.Request, data *types.PostFormVersionRequest) (*types.PostFormVersionResponse, error) {
	session := h.Redis.ReqSession(req)
	var versionId string
	err := h.Database.Client().QueryRow(`
		INSERT INTO dbtable_schema.form_versions (form_id, form, created_on, created_sub)
		VALUES ($1::uuid, $2::jsonb, $3, $4::uuid)
		RETURNING id
	`, data.GetVersion().GetFormId(), data.GetVersion().GetForm(), time.Now().Local().UTC(), session.UserSub).Scan(&versionId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.Database.Client().Exec(`
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.GetName(), time.Now().Local().UTC(), session.UserSub, data.GetVersion().GetFormId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormVersionResponse{Id: versionId}, nil
}

func (h *Handlers) PatchForm(w http.ResponseWriter, req *http.Request, data *types.PatchFormRequest) (*types.PatchFormResponse, error) {
	session := h.Redis.ReqSession(req)
	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.GetForm().GetName(), time.Now().Local().UTC(), session.UserSub, data.GetForm().GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchFormResponse{Success: true}, nil
}

func (h *Handlers) GetForms(w http.ResponseWriter, req *http.Request, data *types.GetFormsRequest) (*types.GetFormsResponse, error) {
	var forms []*types.IProtoForm

	err := h.Database.QueryRows(&forms, `
		SELECT * FROM dbview_schema.enabled_forms
	`)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetFormsResponse{Forms: forms}, nil
}

func (h *Handlers) GetFormById(w http.ResponseWriter, req *http.Request, data *types.GetFormByIdRequest) (*types.GetFormByIdResponse, error) {
	var forms []*types.IProtoForm

	err := h.Database.QueryRows(&forms, `
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

func (h *Handlers) DeleteForm(w http.ResponseWriter, req *http.Request, data *types.DeleteFormRequest) (*types.DeleteFormResponse, error) {
	_, err := h.Database.Client().Exec(`
		DELETE FROM dbtable_schema.forms
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteFormResponse{Success: true}, nil
}

func (h *Handlers) DisableForm(w http.ResponseWriter, req *http.Request, data *types.DisableFormRequest) (*types.DisableFormResponse, error) {
	session := h.Redis.ReqSession(req)
	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.forms
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableFormResponse{Success: true}, nil
}
