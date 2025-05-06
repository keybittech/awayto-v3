package handlers

import (
	"errors"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostGroupFile(info ReqInfo, data *types.PostGroupFileRequest) (*types.PostGroupFileResponse, error) {
	var groupFileId string

	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_files (group_id, file_id, created_on, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
		RETURNING id
	`, info.Session.GroupId, data.GetFileId(), time.Now().Local().UTC(), info.Session.UserSub).Scan(&groupFileId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if groupFileId == "" {
		return nil, util.ErrCheck(errors.New("failed to insert group file"))
	}

	return &types.PostGroupFileResponse{Id: groupFileId}, nil
}

func (h *Handlers) PatchGroupFile(info ReqInfo, data *types.PatchGroupFileRequest) (*types.PatchGroupFileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.group_files
		SET group_id = $2, file_id = $3, updated_sub = $4, updated_on = $5
		WHERE id = $1
	`, data.GetId(), info.Session.GroupId, data.GetFileId(), info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchGroupFileResponse{Success: true}, nil
}

func (h *Handlers) GetGroupFiles(info ReqInfo, data *types.GetGroupFilesRequest) (*types.GetGroupFilesResponse, error) {
	var groupFiles []*types.IGroupFile

	err := h.Database.QueryRows(info.Ctx, info.Tx, &groupFiles, `
		SELECT * FROM dbview_schema.enabled_group_files
		WHERE "groupId" = $1
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupFilesResponse{Files: groupFiles}, nil
}

func (h *Handlers) GetGroupFileById(info ReqInfo, data *types.GetGroupFileByIdRequest) (*types.GetGroupFileByIdResponse, error) {
	var groupFiles []*types.IGroupFile

	err := h.Database.QueryRows(info.Ctx, info.Tx, &groupFiles, `
		SELECT * FROM dbview_schema.enabled_group_files
		WHERE "groupId" = $1 AND id = $2
	`, info.Session.GroupId, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupFileByIdResponse{File: groupFiles[0]}, nil
}

func (h *Handlers) DeleteGroupFile(info ReqInfo, data *types.DeleteGroupFileRequest) (*types.DeleteGroupFileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.group_files
		WHERE id = $1
	`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteGroupFileResponse{Success: true}, nil
}
