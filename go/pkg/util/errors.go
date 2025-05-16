package util

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	ErrorForUser    = "ERROR_FOR_USER"
	ErrorForUserLen = len(ErrorForUser) + 1
)

func UserError(err string) error {
	var sb strings.Builder
	sb.WriteString(ErrorForUser)
	sb.WriteString(" ")
	sb.WriteString(err)
	sb.WriteString(" ")
	sb.WriteString(ErrorForUser)
	return errors.New(sb.String())
}

func SnipUserError(err string) string {
	start := strings.Index(err, ErrorForUser)
	end := strings.LastIndex(err, ErrorForUser)
	return err[start+ErrorForUserLen : end-1]
}

func RequestError(w http.ResponseWriter, givenErr string, ignoreFields []protoreflect.Name, pbVal proto.Message) {
	requestId := uuid.NewString()

	var reqParams strings.Builder
	if pbVal != nil {
		reflectMsg := pbVal.ProtoReflect()
		fields := reflectMsg.Descriptor().Fields()

		for _, ignoreField := range ignoreFields {
			if fieldDescriptor := fields.ByName(ignoreField); fieldDescriptor != nil {
				reflectMsg.Set(fieldDescriptor, protoreflect.ValueOfString("NOLOG_FIELD"))
			}
		}

		pbValJson, err := protojson.Marshal(pbVal)
		if err != nil {
			reqParams.WriteString("ERROR PARSING REQPARAMS: ")
			reqParams.WriteString(err.Error())
		} else {
			reqParams.WriteString(string(pbValJson))
		}
	}

	var reqErr strings.Builder
	reqErr.WriteString("REQUEST_ERROR ")
	reqErr.WriteString(requestId)
	reqErr.WriteByte(' ')

	reqErr.WriteString(givenErr)
	reqErr.WriteByte(' ')

	if params := reqParams.String(); params != "" {
		reqErr.WriteString("Parameters:")
		reqErr.WriteByte(' ')
		reqErr.WriteString(strings.TrimSpace(params))
	}

	reqErrStr := reqErr.String()

	ErrorLog.Println(reqErrStr)

	var userErrRes strings.Builder
	userErrRes.WriteString("Request Id: ")
	userErrRes.WriteString(requestId)
	userErrRes.WriteByte('\n')

	if strings.Index(reqErrStr, ErrorForUser) > -1 {
		userErrRes.WriteString(SnipUserError(reqErrStr))
	} else {
		userErrRes.WriteString("An error occurred. Please try again later or contact your administrator with the request id provided.")
	}

	http.Error(w, userErrRes.String(), http.StatusInternalServerError)
}

func WriteCallerErr(n int, err any, sb *strings.Builder) {
	_, file, line, _ := runtime.Caller(n + 1)
	sb.WriteString(fmt.Sprint(err))
	sb.WriteByte(' ')
	sb.WriteString(file)
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(line))
}

func ErrCheck(err error) error {
	if err == nil {
		return nil
	}

	var sb strings.Builder
	WriteCallerErr(1, err, &sb)

	return errors.New(sb.String())
}

func ErrCheckN(n int, err any) error {
	if err == nil {
		return nil
	}

	var sb strings.Builder
	WriteCallerErr(n, err, &sb)

	return errors.New(sb.String())
}
