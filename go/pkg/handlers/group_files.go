package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupFile(info ReqInfo, data *types.PostGroupFileRequest) (*types.PostGroupFileResponse, error) {
	groupFileInsert := util.BatchQueryRow[types.ILookup](info.Batch, `
		INSERT INTO dbtable_schema.group_files (group_id, file_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		RETURNING id
	`, info.Session.GetGroupId(), data.GetFileId(), info.Session.GetUserSub())

	info.Batch.Send(info.Ctx)

	return &types.PostGroupFileResponse{Id: (*groupFileInsert).Id}, nil
}

func (h *Handlers) PatchGroupFile(info ReqInfo, data *types.PatchGroupFileRequest) (*types.PatchGroupFileResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.group_files
		SET group_id = $2, file_id = $3, updated_sub = $4, updated_on = $5
		WHERE id = $1
	`, data.GetId(), info.Session.GetGroupId(), data.GetFileId(), info.Session.GetUserSub(), time.Now())

	info.Batch.Send(info.Ctx)

	return &types.PatchGroupFileResponse{Success: true}, nil
}

func (h *Handlers) GetGroupFiles(info ReqInfo, data *types.GetGroupFilesRequest) (*types.GetGroupFilesResponse, error) {
	groupFiles := util.BatchQuery[types.IGroupFile](info.Batch, `
		SELECT id, name, "fileId", "createdOn"
		FROM dbview_schema.enabled_group_files
		WHERE "groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupFilesResponse{Files: *groupFiles}, nil
}

func (h *Handlers) GetGroupFileById(info ReqInfo, data *types.GetGroupFileByIdRequest) (*types.GetGroupFileByIdResponse, error) {
	groupFile := util.BatchQueryRow[types.IGroupFile](info.Batch, `
		SELECT id, name, "fileId", "createdOn"
		FROM dbview_schema.enabled_group_files
		WHERE "groupId" = $1 AND id = $2
	`, info.Session.GetGroupId(), data.Id)

	info.Batch.Send(info.Ctx)

	return &types.GetGroupFileByIdResponse{File: *groupFile}, nil
}

func (h *Handlers) DeleteGroupFile(info ReqInfo, data *types.DeleteGroupFileRequest) (*types.DeleteGroupFileResponse, error) {
	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.group_files
		WHERE id = $1
	`, data.Id)

	info.Batch.Send(info.Ctx)

	return &types.DeleteGroupFileResponse{Success: true}, nil
}
