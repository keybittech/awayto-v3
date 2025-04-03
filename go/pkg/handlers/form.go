package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostForm(w http.ResponseWriter, req *http.Request, data *types.PostFormRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PostFormResponse, error) {
	var id string
	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.forms (name, created_on, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, data.GetForm().GetName(), time.Now().Local().UTC(), session.UserSub).Scan(&id)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormResponse{Id: id}, nil
}

func (h *Handlers) PostFormVersion(w http.ResponseWriter, req *http.Request, data *types.PostFormVersionRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PostFormVersionResponse, error) {
	formJson, err := data.GetVersion().GetForm().MarshalJSON()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var versionId string
	err = tx.QueryRow(`
		INSERT INTO dbtable_schema.form_versions (form_id, form, created_on, created_sub)
		VALUES ($1::uuid, $2::jsonb, $3, $4::uuid)
		RETURNING id
	`, data.GetVersion().GetFormId(), formJson, time.Now().Local().UTC(), session.UserSub).Scan(&versionId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.GetName(), time.Now().Local().UTC(), session.UserSub, data.GetVersion().GetFormId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostFormVersionResponse{Id: versionId}, nil
}

func (h *Handlers) PatchForm(w http.ResponseWriter, req *http.Request, data *types.PatchFormRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PatchFormResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.forms
		SET name = $1, updated_on = $2, updated_sub = $3
		WHERE id = $4
	`, data.GetForm().GetName(), time.Now().Local().UTC(), session.UserSub, data.GetForm().GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchFormResponse{Success: true}, nil
}

func (h *Handlers) GetForms(w http.ResponseWriter, req *http.Request, data *types.GetFormsRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.GetFormsResponse, error) {
	var forms []*types.IProtoForm

	err := tx.QueryRows(&forms, `
		SELECT * FROM dbview_schema.enabled_forms
	`)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetFormsResponse{Forms: forms}, nil
}

func (h *Handlers) GetFormById(w http.ResponseWriter, req *http.Request, data *types.GetFormByIdRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.GetFormByIdResponse, error) {
	var forms []*types.IProtoForm

	err := tx.QueryRows(&forms, `
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

func (h *Handlers) DeleteForm(w http.ResponseWriter, req *http.Request, data *types.DeleteFormRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.DeleteFormResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.forms
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteFormResponse{Success: true}, nil
}

func (h *Handlers) DisableForm(w http.ResponseWriter, req *http.Request, data *types.DisableFormRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.DisableFormResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.forms
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableFormResponse{Success: true}, nil
}
