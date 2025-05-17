package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ResponseHandler func(w http.ResponseWriter, results proto.Message) int

func ProtoResponseHandler(w http.ResponseWriter, results proto.Message) int {
	if results == nil {
		return 0
	}

	pbJsonBytes, err := protojson.Marshal(results)
	if err != nil {
		panic(util.ErrCheck(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(pbJsonBytes)))

	n, err := w.Write(pbJsonBytes)
	if err != nil {
		panic(util.ErrCheck(err))
	}

	return n
}

func MultipartResponseHandler(w http.ResponseWriter, results proto.Message) int {
	if results == nil {
		return 0
	}

	resData, ok := results.(*types.GetFileContentsResponse)
	if !ok {
		panic(util.ErrCheck(errors.New("multipart response is not the right proto")))
	}

	n, err := w.Write(resData.Content)
	if err != nil {
		panic(util.ErrCheck(err))
	}

	return n
}
