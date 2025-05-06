package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"

	"github.com/google/uuid"
)

func (h *Handlers) PostFileContents(info ReqInfo, data *types.PostFileContentsRequest) (*types.PostFileContentsResponse, error) {

	// If a file was deleted from the FileManager ui, its uuid won't be sent
	// so delete all unrepresented files
	_, err := info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.file_contents
		WHERE upload_id = $1
		AND uuid NOT IN (SELECT unnest($2::text[]))
	`, data.UploadId, pq.Array(data.ExistingIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	newUuids := make([]string, len(data.Contents))

	for idx, file := range data.GetContents() {

		var pdfDoc []byte

		if strings.HasSuffix(file.GetName(), ".pdf") {
			pdfDoc = file.GetContent()
		} else {

			fileReader := bytes.NewReader(file.GetContent())

			convertBody := &bytes.Buffer{}
			mw := multipart.NewWriter(convertBody)

			formFile, err := mw.CreateFormFile("files", file.GetName())
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			if _, err := io.Copy(formFile, fileReader); err != nil {
				return nil, util.ErrCheck(err)
			}

			mw.Close()

			url := "http://localhost:8000/forms/libreoffice/convert"

			convertReq, err := http.NewRequestWithContext(info.Ctx, http.MethodPost, url, convertBody)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			convertReq.Header.Add("Content-Type", mw.FormDataContentType())

			client := &http.Client{}
			do, err := client.Do(convertReq)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			defer do.Body.Close()

			convertedData, err := io.ReadAll(do.Body)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			pdfDoc = convertedData
		}

		fileUuid, _ := uuid.NewV7()

		_, err := info.Tx.Exec(info.Ctx, `
			INSERT INTO dbtable_schema.file_contents (uuid, name, content, content_length, created_on, created_sub, upload_id)
			VALUES ($1::uuid, $2, $3::bytea, $4, $5, $6, $7)
		`, fileUuid.String(), file.GetName(), pdfDoc, len(pdfDoc), time.Now(), info.Session.UserSub, data.UploadId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		newUuids[idx] = string(fileUuid.String())
	}

	// Remove overwritten files
	_, err = info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.file_contents
		WHERE uuid = ANY($1)
	`, pq.Array(data.OverwriteIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Check file size of current file set
	var totalSize int32

	err = info.Tx.QueryRow(info.Ctx, `
		SELECT SUM(COALESCE(content_length, 0))
		FROM dbtable_schema.file_contents
		WHERE upload_id = $1
	`, data.UploadId).Scan(&totalSize)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if totalSize > 1<<25 {
		return nil, util.ErrCheck(util.UserError("Total file size must not exceed 32MB."))
	}

	return &types.PostFileContentsResponse{Ids: newUuids}, nil
}

func (h *Handlers) PatchFileContents(info ReqInfo, data *types.PatchFileContentsRequest) (*types.PatchFileContentsResponse, error) {
	// expiration := time.Now().Local().UTC().AddDate(0, 1, 0) // Adds 30 days to current time
	// err := h.FS.PatchFile(data.GetId(), data.GetName(), expiration)
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	return &types.PatchFileContentsResponse{Success: true}, nil
}

func (h *Handlers) GetFileContents(info ReqInfo, data *types.GetFileContentsRequest) (*types.GetFileContentsResponse, error) {
	var fileContent []byte

	err := info.Tx.QueryRow(info.Ctx, `
		SELECT content FROM dbtable_schema.file_contents
		WHERE uuid = $1
	`, data.FileId).Scan(&fileContent)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetFileContentsResponse{Content: fileContent}, nil
}

func (h *Handlers) PostFile(info ReqInfo, data *types.PostFileRequest) (*types.PostFileResponse, error) {
	file := data.GetFile()
	var fileID string
	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.files (uuid, name, mime_type, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5::uuid)
		RETURNING id
	`, file.GetUuid(), file.GetName(), file.GetMimeType(), time.Now(), info.Session.UserSub).Scan(&fileID)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.PostFileResponse{Id: fileID, Uuid: file.GetUuid()}, nil
}

func (h *Handlers) PatchFile(info ReqInfo, data *types.PatchFileRequest) (*types.PatchFileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.files
		SET name = $2, updated_on = $3, updated_sub = $4
		WHERE id = $1
	`, data.GetId(), data.GetName(), time.Now(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.PatchFileResponse{Success: true}, nil
}

func (h *Handlers) GetFiles(info ReqInfo, data *types.GetFilesRequest) (*types.GetFilesResponse, error) {
	var files []*types.IFile
	err := h.Database.QueryRows(info.Ctx, info.Tx, &files, "SELECT * FROM dbview_schema.enabled_files")
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.GetFilesResponse{Files: files}, nil
}

func (h *Handlers) GetFileById(info ReqInfo, data *types.GetFileByIdRequest) (*types.GetFileByIdResponse, error) {
	var files []*types.IFile
	err := h.Database.QueryRows(info.Ctx, info.Tx, &files, "SELECT * FROM dbview_schema.enabled_files WHERE id = $1", data.GetId())
	if err != nil || len(files) == 0 {
		return nil, util.ErrCheck(err)
	}
	return &types.GetFileByIdResponse{File: files[0]}, nil
}

func (h *Handlers) DeleteFile(info ReqInfo, data *types.DeleteFileRequest) (*types.DeleteFileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `DELETE FROM dbtable_schema.files WHERE id = $1`, data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.DeleteFileResponse{Id: data.GetId()}, nil
}

func (h *Handlers) DisableFile(info ReqInfo, data *types.DisableFileRequest) (*types.DisableFileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.files
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.DisableFileResponse{Id: data.GetId()}, nil
}
