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

type BodyParser func(w http.ResponseWriter, req *http.Request, msgType protoreflect.MessageType) proto.Message

func ProtoBodyParser(w http.ResponseWriter, req *http.Request, msgType protoreflect.MessageType) proto.Message {
	pb := msgType.New().Interface().(proto.Message)

	if req.Body != nil && req.Body != http.NoBody {
		req.Body = http.MaxBytesReader(w, req.Body, 1<<20) // 1MB limit
		defer req.Body.Close()

		buf, err := io.ReadAll(req.Body)
		if err != nil {
			panic(util.ErrCheck(err))
		}

		if len(buf) > 0 {
			if err := protojson.Unmarshal(buf, pb); err != nil {
				panic(util.ErrCheck(err))
			}

			if err := protovalidate.Validate(pb); err != nil {
				panic(util.ErrCheck(util.UserError(err.Error())))
			}
		}
	}

	return pb
}

func MultipartBodyParser(w http.ResponseWriter, req *http.Request, msgType protoreflect.MessageType) proto.Message {
	req.Body = http.MaxBytesReader(w, req.Body, 1<<25)

	err := req.ParseMultipartForm(1 << 25) // 32MB payload max
	if err != nil {
		panic(util.ErrCheck(util.UserError("Attached files may not exceed 32MB.")))
	}

	pbFiles := &types.PostFileContentsRequest{}

	uploadIdValue, ok := req.MultipartForm.Value["uploadId"]
	if !ok {
		panic(util.ErrCheck(errors.New("invalid multipart request: no uploadId object")))
	}

	if uploadIdValue[0] != "" {
		pbFiles.UploadId = uploadIdValue[0]
	} else {
		panic(util.ErrCheck(errors.New("invalid multipart request: uploadId is empty")))
	}

	existingIdsValue, ok := req.MultipartForm.Value["existingIds"]
	if !ok {
		panic(util.ErrCheck(errors.New("invalid multipart request: no existingIds object")))
	}

	if existingIdsValue[0] != "" {
		pbFiles.ExistingIds = strings.Split(existingIdsValue[0], ",")
	}

	overwriteIdsValue, ok := req.MultipartForm.Value["overwriteIds"]
	if !ok {
		panic(util.ErrCheck(errors.New("invalid multipart request: no overwriteIds object")))
	}

	if overwriteIdsValue[0] != "" {
		pbFiles.OverwriteIds = strings.Split(overwriteIdsValue[0], ",")
	}

	files, ok := req.MultipartForm.File["contents"]
	if !ok {
		panic(util.ErrCheck(errors.New("invalid multipart request: no contents object")))
	}

	if len(pbFiles.ExistingIds)+len(files)-len(pbFiles.OverwriteIds) > 5 {
		panic(util.ErrCheck(util.UserError("No more than 5 files may be uploaded in total.")))
	}

	for _, f := range files {
		fileBuf := make([]byte, f.Size)

		fileData, _ := f.Open()
		_, err := fileData.Read(fileBuf)
		if err != nil {
			panic(util.ErrCheck(err))
		}

		err = fileData.Close()
		if err != nil {
			panic(util.ErrCheck(err))
		}

		fileLen := int64(len(fileBuf))

		pbFiles.Contents = append(pbFiles.Contents, &types.FileContent{
			Name:          f.Filename,
			Content:       fileBuf,
			ContentLength: fileLen,
		})
		pbFiles.TotalLength += fileLen
	}

	return pbFiles
}
