package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"

	"github.com/google/uuid"
)

func (h *Handlers) PostFileContents(w http.ResponseWriter, req *http.Request, data *types.PostFileContentsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostFileContentsResponse, error) {

	// If a file was deleted from the FileManager ui, its uuid won't be sent
	// so delete all unrepresented files
	_, err := tx.Exec(`
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

			convertReq, err := http.NewRequestWithContext(req.Context(), http.MethodPost, url, convertBody)
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

		_, err := tx.Exec(`
			INSERT INTO dbtable_schema.file_contents (uuid, name, content, content_length, created_on, created_sub, upload_id)
			VALUES ($1::uuid, $2, $3::bytea, $4, $5, $6, $7)
		`, fileUuid.String(), file.GetName(), pdfDoc, len(pdfDoc), time.Now(), session.UserSub, data.UploadId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		newUuids[idx] = string(fileUuid.String())
	}

	// Remove overwritten files
	_, err = tx.Exec(`
		DELETE FROM dbtable_schema.file_contents
		WHERE uuid = ANY($1)
	`, pq.Array(data.OverwriteIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// Check file size of current file set
	var totalSize int32

	err = tx.QueryRow(`
		SELECT SUM(COALESCE(content_length, 0))
		FROM dbtable_schema.file_contents
		WHERE upload_id = $1
	`, data.UploadId).Scan(&totalSize)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if totalSize > 20480000 {
		return nil, util.ErrCheck(util.UserError("Total file size must not exceed 20MB."))
	}

	return &types.PostFileContentsResponse{Ids: newUuids}, nil
}

func (h *Handlers) PatchFileContents(w http.ResponseWriter, req *http.Request, data *types.PatchFileContentsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchFileContentsResponse, error) {
	// expiration := time.Now().Local().UTC().AddDate(0, 1, 0) // Adds 30 days to current time
	// err := h.FS.PatchFile(data.GetId(), data.GetName(), expiration)
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	return &types.PatchFileContentsResponse{Success: true}, nil
}

func (h *Handlers) GetFileContents(w http.ResponseWriter, req *http.Request, data *types.GetFileContentsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*[]byte, error) {
	var fileContent []byte

	err := tx.QueryRow(`
		SELECT content FROM dbtable_schema.file_contents
		WHERE uuid = $1
	`, data.FileId).Scan(&fileContent)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &fileContent, nil
}

func (h *Handlers) PostFile(w http.ResponseWriter, req *http.Request, data *types.PostFileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostFileResponse, error) {
	file := data.GetFile()
	var fileID string
	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.files (uuid, name, mime_type, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5::uuid)
		RETURNING id
	`, file.GetUuid(), file.GetName(), file.GetMimeType(), time.Now(), session.UserSub).Scan(&fileID)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.PostFileResponse{Id: fileID, Uuid: file.GetUuid()}, nil
}

func (h *Handlers) PatchFile(w http.ResponseWriter, req *http.Request, data *types.PatchFileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchFileResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.files
		SET name = $2, updated_on = $3, updated_sub = $4
		WHERE id = $1
	`, data.GetId(), data.GetName(), time.Now(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.PatchFileResponse{Success: true}, nil
}

func (h *Handlers) GetFiles(w http.ResponseWriter, req *http.Request, data *types.GetFilesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetFilesResponse, error) {
	var files []*types.IFile
	err := tx.QueryRows(&files, "SELECT * FROM dbview_schema.enabled_files")
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.GetFilesResponse{Files: files}, nil
}

func (h *Handlers) GetFileById(w http.ResponseWriter, req *http.Request, data *types.GetFileByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetFileByIdResponse, error) {
	var files []*types.IFile
	err := tx.QueryRows(&files, "SELECT * FROM dbview_schema.enabled_files WHERE id = $1", data.GetId())
	if err != nil || len(files) == 0 {
		return nil, util.ErrCheck(err)
	}
	return &types.GetFileByIdResponse{File: files[0]}, nil
}

func (h *Handlers) DeleteFile(w http.ResponseWriter, req *http.Request, data *types.DeleteFileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteFileResponse, error) {
	_, err := tx.Exec("DELETE FROM dbtable_schema.files WHERE id = $1", data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.DeleteFileResponse{Id: data.GetId()}, nil
}

func (h *Handlers) DisableFile(w http.ResponseWriter, req *http.Request, data *types.DisableFileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableFileResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.files
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.DisableFileResponse{Id: data.GetId()}, nil
}
