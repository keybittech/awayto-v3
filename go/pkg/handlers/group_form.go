package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
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
	groupSub := info.Session.GetGroupSub()
	info.Session.SetUserSub(groupSub)

	formResp, err := h.PostForm(info, &types.PostFormRequest{Form: data.GetGroupForm().GetForm()})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var groupFormId string
	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_forms (group_id, form_id, created_sub)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (group_id, form_id) DO NOTHING
		RETURNING id
	`, info.Session.GetGroupId(), formResp.GetId(), groupSub).Scan(&groupFormId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	formJson, err := data.GetGroupForm().GetForm().GetVersion().GetForm().MarshalJSON()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.form_versions (form_id, form, created_sub)
		SELECT gf.form_id, $2::jsonb, $3::uuid
		FROM dbtable_schema.group_forms gf
		WHERE gf.id = $1::uuid
	`, groupFormId, formJson, groupSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, groupRoleId := range data.GetGroupRoleIds() {
		_, err = info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.group_form_roles (group_form_id, group_role_id, created_sub)
			SELECT gf.id, $2::uuid, $3::uuid
			FROM dbtable_schema.group_forms gf
			WHERE gf.id = $1::uuid
		`, groupFormId, groupRoleId, groupSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.PostGroupFormResponse{Id: formResp.GetId()}, nil
}

func (h *Handlers) PostGroupFormVersion(info ReqInfo, data *types.PostGroupFormVersionRequest) (*types.PostGroupFormVersionResponse, error) {
	groupSub := info.Session.GetGroupSub()
	info.Session.SetUserSub(groupSub)

	groupFormId := data.GetGroupFormId()

	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.form_versions fv
		SET active = false
		FROM dbtable_schema.group_forms gf
		WHERE fv.form_id = gf.form_id AND gf.id = $1::uuid
	`, groupFormId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	formJson, err := data.GroupFormVersion.GetForm().MarshalJSON()
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var versionId string
	err = info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.form_versions (form_id, form, created_sub, active)
		SELECT gf.form_id, $2::jsonb, $3::uuid, true
		FROM dbtable_schema.group_forms gf
		WHERE gf.id = $1::uuid
		RETURNING id
	`, groupFormId, formJson, groupSub).Scan(&versionId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.forms f
		SET name = $1, updated_on = $2, updated_sub = $3
		FROM dbtable_schema.group_forms gf
		WHERE f.id = gf.form_id AND gf.id = $4::uuid
	`, data.GetName(), time.Now(), groupSub, groupFormId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.group_form_roles gfr
		USING dbtable_schema.group_forms gf
		WHERE gfr.group_form_id = gf.id AND gf.id = $1::uuid
	`, groupFormId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, groupRoleId := range data.GetGroupRoleIds() {
		_, err = info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.group_form_roles (group_form_id, group_role_id, created_sub)
			SELECT gf.id, $2::uuid, $3::uuid
			FROM dbtable_schema.group_forms gf
			WHERE gf.id = $1::uuid
		`, groupFormId, groupRoleId, groupSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.PostGroupFormVersionResponse{Id: versionId}, nil
}

func (h *Handlers) PatchGroupForm(info ReqInfo, data *types.PatchGroupFormRequest) (*types.PatchGroupFormResponse, error) {
	form := data.GetGroupForm().GetForm()

	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.forms f
		SET name = $1, updated_on = $2, updated_sub = $3
		FROM dbtable_schema.group_forms gf
		WHERE f.id = gf.form_id AND f.id = $4::uuid
	`, form.GetName(), time.Now(), info.Session.GetUserSub(), form.GetId())

	info.Batch.Send(info.Ctx)

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
	groupFormReq := util.BatchQueryRow[types.IGroupForm](info.Batch, `
		SELECT id, "formId", "groupId", form
		FROM dbview_schema.enabled_group_forms_ext 
		WHERE "groupId" = $1 AND "formId" = $2
	`, info.Session.GetGroupId(), data.GetFormId())

	formVersionIds := util.BatchQueryRow[struct {
		Ids []string `json:"ids"`
	}](info.Batch, `
		SELECT ARRAY_AGG(fv.id ORDER BY fv.created_on DESC) AS ids
		FROM dbtable_schema.form_versions fv
		JOIN dbtable_schema.group_forms gf ON gf.form_id = fv.form_id
		WHERE fv.form_id = $1
	`, data.GetFormId())

	info.Batch.Send(info.Ctx)

	groupForm := *groupFormReq

	info.Batch.Reset()

	groupRoleIdsReq := util.BatchQuery[types.ILookup](info.Batch, `
		SELECT gfr.group_role_id as id
		FROM dbtable_schema.group_form_roles gfr
		JOIN dbtable_schema.group_forms gf ON gf.id = gfr.group_form_id
		WHERE gfr.group_form_id = $1::uuid
	`, groupForm.GetId())

	info.Batch.Send(info.Ctx)

	var groupRoleIds []string
	for _, idLookup := range *groupRoleIdsReq {
		groupRoleIds = append(groupRoleIds, (*idLookup).GetId())
	}

	return &types.GetGroupFormByIdResponse{GroupForm: groupForm, GroupRoleIds: groupRoleIds, VersionIds: (*formVersionIds).Ids}, nil
}

func (h *Handlers) PatchGroupFormActiveVersion(info ReqInfo, data *types.PatchGroupFormActiveVersionRequest) (*types.PatchGroupFormActiveVersionResponse, error) {

	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.form_versions fv
		SET active = false
		FROM dbtable_schema.group_forms gf
		WHERE fv.form_id = gf.form_id AND gf.form_id = $1::uuid
	`, data.GetFormId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.form_versions fv
		SET active = true
		FROM dbtable_schema.group_forms gf
		WHERE fv.form_id = gf.form_id AND gf.form_id = $1::uuid AND fv.id = $2::uuid
	`, data.GetFormId(), data.GetFormVersionId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchGroupFormActiveVersionResponse{Success: true}, nil
}

func (h *Handlers) GetGroupFormActiveVersion(info ReqInfo, data *types.GetGroupFormActiveVersionRequest) (*types.GetGroupFormActiveVersionResponse, error) {
	groupFormReq := util.BatchQueryRow[types.IGroupForm](info.Batch, `
		SELECT id, "formId", "groupId", form
		FROM dbview_schema.enabled_group_forms_active
		WHERE "groupId" = $1 AND "formId" = $2
	`, info.Session.GetGroupId(), data.GetFormId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupFormActiveVersionResponse{GroupForm: *groupFormReq}, nil
}

func (h *Handlers) GetGroupFormVersionById(info ReqInfo, data *types.GetGroupFormVersionByIdRequest) (*types.GetGroupFormVersionByIdResponse, error) {
	version := util.BatchQueryRow[types.IProtoFormVersion](info.Batch, `
		SELECT fv."formId", fv.form, fv."createdOn"
		FROM dbview_schema.enabled_form_versions fv
		JOIN dbtable_schema.group_forms gf ON gf.form_id = fv."formId"
		WHERE fv.id = $1::uuid
	`, data.GetFormVersionId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupFormVersionByIdResponse{Version: *version}, nil
}

func (h *Handlers) DeleteGroupForm(info ReqInfo, data *types.DeleteGroupFormRequest) (*types.DeleteGroupFormResponse, error) {
	formIds := strings.Split(data.Ids, ",")

	for _, formId := range formIds {
		var serviceName, formName string
		err := info.Tx.QueryRow(info.Ctx, `
			SELECT s.name AS sn, f.name AS fn
			FROM dbtable_schema.service_forms sf
			JOIN dbtable_schema.group_services gs ON gs.service_id = sf.service_id
			JOIN dbtable_schema.services s ON s.id = sf.service_id
			JOIN dbtable_schema.forms f ON f.id = sf.form_id
			WHERE sf.form_id = $1
		`, formId).Scan(&serviceName, &formName)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, util.ErrCheck(err)
		}

		if serviceName != "" {
			userErr := fmt.Sprintf("The form %s is currently being used by the service %s and cannot be deleted.", formName, serviceName)
			return nil, util.ErrCheck(util.UserError(userErr))
		}
	}

	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.group_forms
		WHERE form_id = ANY($1)
	`, pq.Array(formIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.forms f
		USING dbtable_schema.group_forms gf
		WHERE f.id = gf.form_id AND f.id = ANY($1)
	`, pq.Array(formIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteGroupFormResponse{Success: true}, nil
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func sanitizeColName(name string) string {
	clean := nonAlphanumericRegex.ReplaceAllString(name, "_")
	return strings.ToLower(strings.Trim(clean, "_"))
}

func getFormTemplate(info ReqInfo, formVersionId string) (*types.IProtoFormTemplate, error) {
	var form string

	err := info.Tx.QueryRow(info.Ctx, `
		SELECT fv.form
		FROM dbtable_schema.form_versions fv
		JOIN dbtable_schema.group_forms gf ON gf.form_id = fv.form_id
		WHERE fv.id = $1::uuid
	`, formVersionId).Scan(&form)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	pbForm := &structpb.Value{}

	if err := protojson.Unmarshal([]byte(form), pbForm); err != nil {
		return nil, util.ErrCheck(err)
	}

	formVersion := &types.IProtoFormVersion{
		Id:   formVersionId,
		Form: pbForm,
	}

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
			} else {
				panic(util.ErrCheck(fmt.Errorf("field unmarshal error: %s %v", string(fieldBytes), err)))
			}
		}

		formTemplate.Rows[rowKey] = &types.IProtoFieldRow{
			Fields: typedFields,
		}
	}

	return formTemplate, nil
}

func (h *Handlers) GetGroupFormVersionData(info ReqInfo, data *types.GetGroupFormVersionDataRequest) (*types.GetGroupFormVersionDataResponse, error) {
	formTemplate, err := getFormTemplate(info, data.GetFormVersionId())
	if err != nil {
		return nil, util.ErrCheck(err)
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
		FROM dbtable_schema.form_version_submissions fvs
		JOIN dbtable_schema.form_versions fv ON fv.id = fvs.form_version_id
		JOIN dbtable_schema.group_forms gf ON gf.form_id = fv.form_id
		WHERE fvs.form_version_id = '%s'
	`, strings.Join(dataCols, ", "), data.GetFormVersionId())

	println("the query ", query)

	return &types.GetGroupFormVersionDataResponse{}, nil
}

func (h *Handlers) GetGroupFormVersionReport(info ReqInfo, data *types.GetGroupFormVersionReportRequest) (*types.GetGroupFormVersionReportResponse, error) {
	formVersionId := data.GetFormVersionId()

	formTemplate, err := getFormTemplate(info, formVersionId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var targetField *types.IProtoField
	found := false

	for _, row := range formTemplate.GetRows() {
		for _, field := range row.GetFields() {
			if field.GetI() == data.GetFieldId() {
				targetField = field
				found = true
				break
			}
		}
	}

	if !found {
		return nil, util.ErrCheck(fmt.Errorf("field %s not found in form version", data.GetFieldId()))
	}

	var query string
	isMultiSelect := false

	switch targetField.GetT() {
	case "multi-select":
		isMultiSelect = true
		var sumStatements []string
		for _, opt := range targetField.GetO() {
			stmt := fmt.Sprintf(
				`SUM(CASE WHEN submission->'%s' @> '"%s"' THEN 1 ELSE 0 END) as "%s"`,
				targetField.GetI(),
				opt.GetV(),
				opt.GetL(),
			)
			sumStatements = append(sumStatements, stmt)
		}
		if len(sumStatements) == 0 {
			return &types.GetGroupFormVersionReportResponse{DataPoints: []*types.IProtoFormDataPoint{}}, nil
		}
		query = fmt.Sprintf(`
			SELECT %s
			FROM dbtable_schema.form_version_submissions fvs
			JOIN dbtable_schema.form_versions fv ON fv.id = fvs.form_version_id
			JOIN dbtable_schema.group_forms gf ON gf.form_id = fv.form_id
			WHERE fvs.form_version_id = '%s'
		`, strings.Join(sumStatements, ", "), formVersionId)
	case "boolean", "single-select", "number":
		coalesceNullText := "No Data"
		if targetField.GetT() == "boolean" {
			coalesceNullText = "false"
		}
		colSelect := fmt.Sprintf("submission->>'%s'", targetField.GetI())
		query = fmt.Sprintf(`
			SELECT COALESCE(%s, '%s') AS label, COUNT(*) AS value
			FROM dbtable_schema.form_version_submissions fvs
			JOIN dbtable_schema.form_versions fv ON fv.id = fvs.form_version_id
			JOIN dbtable_schema.group_forms gf ON gf.form_id = fv.form_id
			WHERE fvs.form_version_id = '%s'
			GROUP BY 1
			ORDER BY 2 DESC
		`, colSelect, coalesceNullText, formVersionId)
	default:
		return nil, util.ErrCheck(fmt.Errorf("reporting not available for field type: %s", targetField.GetT()))
	}

	rows, err := info.Tx.Query(info.Ctx, query)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer rows.Close()

	var results []*types.IProtoFormDataPoint

	if isMultiSelect {
		if rows.Next() {
			fieldDescs := rows.FieldDescriptions()

			values := make([]any, len(fieldDescs))
			scanArgs := make([]any, len(fieldDescs))

			for i := range values {
				scanArgs[i] = &values[i]
			}

			if err := rows.Scan(scanArgs...); err != nil {
				return nil, util.ErrCheck(err)
			}

			for i, fd := range fieldDescs {
				var countVal int64
				switch v := values[i].(type) {
				case int64:
					countVal = v
				default:
					countVal = 0
				}

				results = append(results, &types.IProtoFormDataPoint{
					Label: fd.Name,
					Value: countVal,
				})
			}
		}
	} else {
		for rows.Next() {
			var label string
			var value int64
			if err := rows.Scan(&label, &value); err != nil {
				return nil, util.ErrCheck(err)
			}

			results = append(results, &types.IProtoFormDataPoint{
				Label: label,
				Value: value,
			})
		}

	}

	return &types.GetGroupFormVersionReportResponse{DataPoints: results}, nil
}
