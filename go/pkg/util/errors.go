package util

import (
	"errors"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	ErrorForUser = "ERROR_FOR_USER"
)

func UserError(err string) error {
	return errors.New(ErrorForUser + " " + err + " " + ErrorForUser)
}

const ErrorForUserLen = len(ErrorForUser) + 1

func SnipUserError(err string) string {
	result := err[ErrorForUserLen:]
	return result[:len(result)-ErrorForUserLen]
}

func RequestError(w http.ResponseWriter, givenErr string, ignoreFields []protoreflect.Name, pbVal proto.Message) error {
	requestId := uuid.NewString()

	var reqParams strings.Builder
	if pbVal != nil {
		reflectMsg := pbVal.ProtoReflect()
		fields := reflectMsg.Descriptor().Fields()

		for _, ignoreField := range ignoreFields {
			if fieldDescriptor := fields.ByName(ignoreField); fieldDescriptor != nil {
				reflectMsg.Clear(fieldDescriptor)
			}
		}

		reqParams.WriteString(prototext.Format(pbVal))
	}

	var reqErr strings.Builder
	reqErr.WriteString("REQUEST_ERROR ")
	reqErr.WriteString(requestId)
	reqErr.WriteString(" ")
	reqErr.WriteString(givenErr)

	if reqParams.String() != "" {
		reqErr.WriteString(" ")
		reqErr.WriteString(reqParams.String())
	}

	ErrorLog.Println(reqErr.String())

	reqErrStr := reqErr.String()

	var userErrRes strings.Builder
	userErrRes.WriteString("Request Id: ")
	userErrRes.WriteString(requestId)
	userErrRes.WriteString("\n")

	if strings.Index(reqErrStr, ErrorForUser) > -1 {
		userErrRes.WriteString(SnipUserError(reqErrStr))
	} else {
		userErrRes.WriteString("An error occurred. Please try again later or contact your administrator with the request id provided.")
	}

	http.Error(w, userErrRes.String(), http.StatusInternalServerError)

	return errors.New(reqErr.String())
}

func ErrCheck(err error) error {
	if err == nil {
		return nil
	}

	_, file, line, _ := runtime.Caller(1)

	var errStr strings.Builder
	errStr.WriteString(err.Error())
	errStr.WriteString(" ")
	errStr.WriteString(file)
	errStr.WriteString(":")
	errStr.WriteString(strconv.Itoa(line))
	errStr.WriteString("\n")

	return errors.New(errStr.String())
}
