package api

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ResponseHandler func(w http.ResponseWriter, results []reflect.Value) (int, error)

func ProtoResponseHandler(w http.ResponseWriter, results []reflect.Value) (int, error) {
	var resLen int
	if resData, ok := results[0].Interface().(protoreflect.ProtoMessage); ok {
		pbJsonBytes, err := protojson.Marshal(resData)
		if err != nil {
			return 0, util.ErrCheck(err)
		}
		resLen = len(pbJsonBytes)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", resLen))
		w.Write(pbJsonBytes)
	}

	return resLen, nil
}

func MultipartResponseHandler(w http.ResponseWriter, results []reflect.Value) (int, error) {
	var resLen int
	if resData, ok := results[0].Interface().([]byte); ok {
		resLen = len(resData)
		_, err := w.Write(resData)
		if err != nil {
			return 0, util.ErrCheck(err)
		}
	}

	return resLen, nil
}
