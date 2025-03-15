package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/bufbuild/protovalidate-go"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type BodyParser func(w http.ResponseWriter, req *http.Request, handlerOpts *util.HandlerOptions, serviceType protoreflect.MessageType) (protoreflect.ProtoMessage, error)

func ProtoBodyParser(w http.ResponseWriter, req *http.Request, handlerOpts *util.HandlerOptions, serviceType protoreflect.MessageType) (protoreflect.ProtoMessage, error) {

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer req.Body.Close()

	pb := serviceType.New().Interface()
	if len(body) > 0 {
		err = protojson.Unmarshal(body, pb)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	err = protovalidate.Validate(pb)
	if err != nil {
		return nil, util.ErrCheck(util.UserError(err.Error()))
	}

	return pb, nil
}

func MultipartBodyParser(w http.ResponseWriter, req *http.Request, handlerOpts *util.HandlerOptions, serviceType protoreflect.MessageType) (protoreflect.ProtoMessage, error) {

	req.Body = http.MaxBytesReader(w, req.Body, 20480000)

	// reader, err := req.MultipartReader()
	// if err != nil {
	// 	deferredError = util.ErrCheck(err)
	// 	return
	// }
	//
	// part, err := reader.NextPart()
	// if err != nil && err != io.EOF {
	// 	deferredError = util.ErrCheck(err)
	// 	return
	// }
	//
	// if part.FormName() != "contents" {
	// 	println("GOT NAME", part.FormName())
	// 	deferredError = util.ErrCheck(err)
	// 	return
	// }
	//
	// buf := bufio.NewReader(part)
	// contentTypeBytes, _ := buf.Peek(512)
	// contentType := http.DetectContentType(contentTypeBytes)
	// println("CONTENT TYPE", contentType)
	//
	// deferredError = util.ErrCheck(util.UserError("test"))
	//
	// return

	req.ParseMultipartForm(20480000)

	files, ok := req.MultipartForm.File["contents"]

	if !ok {
		return nil, util.ErrCheck(errors.New("invalid multipart request"))
	}

	pbFiles := &types.PostFileContentsRequest{}

	for _, f := range files {
		fileBuf := make([]byte, f.Size)

		fileData, _ := f.Open()
		fileData.Read(fileBuf)
		fileData.Close()

		pbFiles.Contents = append(pbFiles.Contents, &types.FileContent{
			Name:    f.Filename,
			Content: fileBuf,
		})
	}

	return pbFiles, nil
}
