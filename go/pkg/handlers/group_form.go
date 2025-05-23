package handlers

import (
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"
)

func (h *Handlers) PostGroupForm(info ReqInfo, data *types.PostGroupFormRequest) (*types.PostGroupFormResponse, error) {
	var formExists bool
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM dbtable_schema.forms f
			LEFT JOIN dbtable_schema.group_forms gf ON gf.form_id = f.id
			WHERE f.name = $1 AND gf.group_id = $2
		)
	`, data.Name, info.Session.GetGroupId()).Scan(&formExists)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if formExists {
		return nil, util.ErrCheck(util.UserError("A form with this name already exists."))
	}

	// Group forms should be owned by the group user
	info.Session.SetUserSub(info.Session.GetGroupSub())

	formResp, err := h.PostForm(info, &types.PostFormRequest{Form: data.GetGroupForm().GetForm()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.PostFormVersion(info, &types.PostFormVersionRequest{
		Name: data.GetGroupForm().GetForm().GetName(),
		Version: &types.IProtoFormVersion{
			FormId: formResp.Id,
			Form:   data.GetGroupForm().GetForm().GetVersion().GetForm(),
		},
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.group_forms (group_id, form_id, created_sub)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (group_id, form_id) DO NOTHING
	`, info.Session.GetGroupId(), formResp.Id, info.Session.GetGroupSub())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostGroupFormResponse{Id: formResp.Id}, nil
}

func (h *Handlers) PostGroupFormVersion(info ReqInfo, data *types.PostGroupFormVersionRequest) (*types.PostGroupFormVersionResponse, error) {
	formVersionResp, err := h.PostFormVersion(info, &types.PostFormVersionRequest{Name: data.GetName(), Version: data.GetGroupFormVersion()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostGroupFormVersionResponse{Id: formVersionResp.GetId()}, nil
}

func (h *Handlers) PatchGroupForm(info ReqInfo, data *types.PatchGroupFormRequest) (*types.PatchGroupFormResponse, error) {
	_, err := h.PatchForm(info, &types.PatchFormRequest{Form: data.GetGroupForm().GetForm()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchGroupFormResponse{Success: true}, nil
}

func (h *Handlers) GetGroupForms(info ReqInfo, data *types.GetGroupFormsRequest) (*types.GetGroupFormsResponse, error) {
	forms := util.BatchQuery[types.IProtoForm](info.Batch, `
		SELECT ef.id, ef.name, ef."createdOn"
		FROM dbview_schema.enabled_group_forms egf
		LEFT JOIN dbview_schema.enabled_forms ef ON ef.id = egf."formId"
		WHERE egf."groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	var groupForms []*types.IGroupForm
	for _, f := range *forms {
		groupForms = append(groupForms, &types.IGroupForm{FormId: f.Id, Form: f})
	}

	return &types.GetGroupFormsResponse{GroupForms: groupForms}, nil
}

func (h *Handlers) GetGroupFormById(info ReqInfo, data *types.GetGroupFormByIdRequest) (*types.GetGroupFormByIdResponse, error) {
	groupForm := util.BatchQueryRow[types.IGroupForm](info.Batch, `
		SELECT "formId", "groupId", form
		FROM dbview_schema.enabled_group_forms_ext 
		WHERE "groupId" = $1 AND "formId" = $2
	`, info.Session.GetGroupId(), data.FormId)

	info.Batch.Send(info.Ctx)

	return &types.GetGroupFormByIdResponse{GroupForm: *groupForm}, nil
}

func (h *Handlers) DeleteGroupForm(info ReqInfo, data *types.DeleteGroupFormRequest) (*types.DeleteGroupFormResponse, error) {
	formIds := strings.Split(data.Ids, ",")

	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.group_forms
		WHERE form_id = ANY($1)
	`, pq.Array(formIds))

	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.forms
		WHERE id = ANY($1)
	`, pq.Array(formIds))

	info.Batch.Send(info.Ctx)

	return &types.DeleteGroupFormResponse{Success: true}, nil
}
