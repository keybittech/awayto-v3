package api

import (
	"fmt"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ResponseHandler func(w http.ResponseWriter, results proto.Message) (int, error)

func ProtoResponseHandler(w http.ResponseWriter, results proto.Message) (int, error) {
	var resLen int
	pbJsonBytes, err := protojson.Marshal(results)
	if err != nil {
		return 0, util.ErrCheck(err)
	}
	resLen = len(pbJsonBytes)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", resLen))
	_, err = w.Write(pbJsonBytes)
	if err != nil {
		return 0, util.ErrCheck(err)
	}

	return resLen, nil
}

func MultipartResponseHandler(w http.ResponseWriter, results proto.Message) (int, error) {
	var resLen int
	if resData, ok := results.(*types.GetFileContentsResponse); ok {
		resLen = len(resData.Content)
		_, err := w.Write(resData.Content)
		if err != nil {
			return 0, util.ErrCheck(err)
		}
	}

	return resLen, nil
}
