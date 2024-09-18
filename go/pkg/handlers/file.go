package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (h *Handlers) PostFileContents(w http.ResponseWriter, req *http.Request, data *types.PostFileContentsRequest) (*types.PostFileContentsResponse, error) {
	session := h.Redis.ReqSession(req)

	tx, _ := h.Database.ReqTx(req)

	defer tx.Rollback()

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
			INSERT INTO dbtable_schema.file_contents (uuid, name, content, created_on, created_sub)
			VALUES ($1::uuid, $2, $3::bytea, $4, $5)
		`, fileUuid.String(), file.GetName(), pdfDoc, time.Now().Local().UTC(), session.UserSub)

		if err != nil {
			return nil, util.ErrCheck(err)
		}

		newUuids[idx] = string(fileUuid.String())
	}

	tx.Commit()

	return &types.PostFileContentsResponse{Ids: newUuids}, nil
}

func (h *Handlers) PatchFileContents(w http.ResponseWriter, req *http.Request, data *types.PatchFileContentsRequest) (*types.PatchFileContentsResponse, error) {
	// expiration := time.Now().Local().UTC().AddDate(0, 1, 0) // Adds 30 days to current time
	// err := h.FS.PatchFile(data.GetId(), data.GetName(), expiration)
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	return &types.PatchFileContentsResponse{Success: true}, nil
}

func (h *Handlers) GetFileContents(w http.ResponseWriter, req *http.Request, data *types.GetFileContentsRequest) (*[]byte, error) {

	var fileContent []byte

	err := h.Database.Client().QueryRow(`
		SELECT content FROM dbtable_schema.file_contents
		WHERE uuid = $1
	`, data.GetFileId()).Scan(&fileContent)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// fmt.Printf("\n got file contents %X", fileContent)

	// file, err := h.FS.GetFile(data.GetFileId())
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	return &fileContent, nil
}

func (h *Handlers) PostFile(w http.ResponseWriter, req *http.Request, data *types.PostFileRequest) (*types.PostFileResponse, error) {
	session := h.Redis.ReqSession(req)
	file := data.GetFile()
	var fileID string
	err := h.Database.Client().QueryRow(`
		INSERT INTO dbtable_schema.files (uuid, name, mime_type, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5::uuid)
		RETURNING id
	`, file.GetUuid(), file.GetName(), file.GetMimeType(), time.Now().Local().UTC(), session.UserSub).Scan(&fileID)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.PostFileResponse{Id: fileID, Uuid: file.GetUuid()}, nil
}

func (h *Handlers) PatchFile(w http.ResponseWriter, req *http.Request, data *types.PatchFileRequest) (*types.PatchFileResponse, error) {
	session := h.Redis.ReqSession(req)
	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.files
		SET name = $2, updated_on = $3, updated_sub = $4
		WHERE id = $1
	`, data.GetId(), data.GetName(), time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.PatchFileResponse{Success: true}, nil
}

func (h *Handlers) GetFiles(w http.ResponseWriter, req *http.Request) (*types.GetFilesResponse, error) {
	var files []*types.IFile
	err := h.Database.QueryRows(&files, "SELECT * FROM dbview_schema.enabled_files")
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.GetFilesResponse{Files: files}, nil
}

func (h *Handlers) GetFileById(w http.ResponseWriter, req *http.Request, data *types.GetFileByIdRequest) (*types.GetFileByIdResponse, error) {
	var files []*types.IFile
	err := h.Database.QueryRows(&files, "SELECT * FROM dbview_schema.enabled_files WHERE id = $1", data.GetId())
	if err != nil || len(files) == 0 {
		return nil, util.ErrCheck(err)
	}
	return &types.GetFileByIdResponse{File: files[0]}, nil
}

func (h *Handlers) DeleteFile(w http.ResponseWriter, req *http.Request, data *types.DeleteFileRequest) (*types.DeleteFileResponse, error) {
	_, err := h.Database.Client().Exec("DELETE FROM dbtable_schema.files WHERE id = $1", data.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.DeleteFileResponse{Id: data.GetId()}, nil
}

func (h *Handlers) DisableFile(w http.ResponseWriter, req *http.Request, data *types.DisableFileRequest) (*types.DisableFileResponse, error) {
	session := h.Redis.ReqSession(req)
	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.files
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	return &types.DisableFileResponse{Id: data.GetId()}, nil
}
