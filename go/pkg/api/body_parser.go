package api

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type BodyParser func(w http.ResponseWriter, req *http.Request, handlerOpts *util.HandlerOptions, serviceType protoreflect.MessageType) (proto.Message, error)

func ProtoBodyParser(w http.ResponseWriter, req *http.Request, handlerOpts *util.HandlerOptions, serviceType protoreflect.MessageType) (proto.Message, error) {
	pb := serviceType.New().Interface().(proto.Message)

	if req.Body != nil && req.Body != http.NoBody {
		req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
		defer req.Body.Close()

		buf, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if len(buf) > 0 {
			err = protojson.Unmarshal(buf, pb)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			err = protovalidate.Validate(pb)
			if err != nil {
				return nil, util.ErrCheck(util.UserError(err.Error()))
			}
		}
	}

	return pb, nil
}

func MultipartBodyParser(w http.ResponseWriter, req *http.Request, handlerOpts *util.HandlerOptions, serviceType protoreflect.MessageType) (proto.Message, error) {
	req.Body = http.MaxBytesReader(w, req.Body, 1<<25)

	err := req.ParseMultipartForm(1 << 25)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Attached files may not exceed 32MB."))
	}

	pbFiles := &types.PostFileContentsRequest{}

	uploadIdValue, ok := req.MultipartForm.Value["uploadId"]
	if !ok {
		return nil, util.ErrCheck(errors.New("invalid multipart request"))
	}

	if uploadIdValue[0] != "" {
		pbFiles.UploadId = uploadIdValue[0]
	} else {
		return nil, util.ErrCheck(errors.New("invalid multipart request"))
	}

	existingIdsValue, ok := req.MultipartForm.Value["existingIds"]
	if !ok {
		return nil, util.ErrCheck(errors.New("invalid multipart request"))
	}

	if existingIdsValue[0] != "" {
		pbFiles.ExistingIds = strings.Split(existingIdsValue[0], ",")
	}

	overwriteIdsValue, ok := req.MultipartForm.Value["overwriteIds"]
	if !ok {
		return nil, util.ErrCheck(errors.New("invalid multipart request"))
	}

	if overwriteIdsValue[0] != "" {
		pbFiles.OverwriteIds = strings.Split(overwriteIdsValue[0], ",")
	}

	files, ok := req.MultipartForm.File["contents"]
	if !ok {
		return nil, util.ErrCheck(errors.New("invalid multipart request"))
	}

	if len(pbFiles.ExistingIds)+len(files)-len(pbFiles.OverwriteIds) > 5 {
		return nil, util.ErrCheck(util.UserError("No more than 5 files may be uploaded in total."))
	}

	for _, f := range files {
		fileBuf := make([]byte, f.Size)

		fileData, _ := f.Open()
		fileData.Read(fileBuf)
		fileData.Close()

		fileLen := int64(len(fileBuf))

		pbFiles.Contents = append(pbFiles.Contents, &types.FileContent{
			Name:          f.Filename,
			Content:       fileBuf,
			ContentLength: fileLen,
		})
		pbFiles.TotalLength += fileLen
	}

	return pbFiles, nil
}
