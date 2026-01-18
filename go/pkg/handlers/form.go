package handlers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func sanitizeColName(name string) string {
	clean := nonAlphanumericRegex.ReplaceAllString(name, "_")
	return strings.ToLower(strings.Trim(clean, "_"))
}

func (h *Handlers) GetFormData(info ReqInfo, data *types.GetFormDataRequest) (*types.GetFormDataResponse, error) {

	version := util.BatchQueryRow[types.IProtoFormVersion](info.Batch, `
		SELECT id, form
		FROM dbtable_schema.form_versions
		WHERE form_id = $1::uuid
		ORDER BY created_on DESC
		LIMIT 1
	`, data.GetFormId())

	info.Batch.Send(info.Ctx)

	formVersion := *version

	formTemplate := &types.IProtoFormTemplate{
		Rows: make(map[string]*types.IProtoFieldRow),
	}

	for rowKey, rowVal := range formVersion.GetForm().GetStructValue().Fields {
		listVal := rowVal.GetListValue()
		if listVal == nil {
			continue
		}

		var typedFields []*types.IProtoField

		for _, rawField := range listVal.Values {
			fieldBytes, _ := protojson.Marshal(rawField)
			fieldMsg := &types.IProtoField{}
			if err := protojson.Unmarshal(fieldBytes, fieldMsg); err == nil {
				typedFields = append(typedFields, fieldMsg)
			}
		}

		formTemplate.Rows[rowKey] = &types.IProtoFieldRow{
			Fields: typedFields,
		}
	}

	dataCols := []string{
		"id",
		"created_on",
		"created_sub",
	}

	rowKeys := make([]string, 0, len(formTemplate.Rows))
	for k := range formTemplate.GetRows() {
		rowKeys = append(rowKeys, k)
	}
	sort.Strings(rowKeys)

	for _, rowKey := range rowKeys {
		fieldRow := formTemplate.Rows[rowKey]

		for _, field := range fieldRow.GetFields() {

			if field.GetT() == "labelntext" {
				continue
			}

			colName := sanitizeColName(field.GetL())

			fieldId := field.GetI()

			switch field.GetT() {
			case "multi-select":
				for _, opt := range field.GetO() {
					optLabel := sanitizeColName(opt.GetL())
					fullColName := fmt.Sprintf("%s_%s", colName, optLabel)

					colDef := fmt.Sprintf(`(submission->'%s' @> '"%s"') AS "%s"`, fieldId, opt.GetV(), fullColName)
					dataCols = append(dataCols, colDef)
				}
			case "boolean":
				colDef := fmt.Sprintf(`(submission->>'%s')::BOOLEAN AS "%s"`, fieldId, colName)
				dataCols = append(dataCols, colDef)
			case "number":
				colDef := fmt.Sprintf(`(submission->>'%s')::NUMERIC AS "%s"`, fieldId, colName)
				dataCols = append(dataCols, colDef)
			default:
				colDef := fmt.Sprintf(`submission->>'%s' AS "%s"`, fieldId, colName)
				dataCols = append(dataCols, colDef)
			}
		}
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM dbtable_schema.form_version_submissions
		WHERE form_version_id = '%s'
	`, strings.Join(dataCols, ", "), formVersion.GetId())

	println("the query ", query)

	return &types.GetFormDataResponse{}, nil
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
