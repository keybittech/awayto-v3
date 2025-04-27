package util

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
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

func RequestError(w http.ResponseWriter, givenErr string, ignoreFields []string, pbVal proto.Message) error {
	requestId := uuid.NewString()
	defaultErr := fmt.Sprintf("%s\nAn error occurred. Please try again later or contact your administrator with the request id provided.", requestId)

	var reqParams string
	if pbVal != nil {
		reflectMsg := pbVal.ProtoReflect()
		descriptor := reflectMsg.Descriptor()
		fields := descriptor.Fields()

		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			fieldName := string(field.Name())

			if !slices.Contains(ignoreFields, fieldName) {
				// Check if it's a string field
				if field.Kind() == protoreflect.StringKind {
					// Get the field value
					value := reflectMsg.Get(field)
					strValue := value.String()
					reqParams += fieldName + "=" + strValue + " "
				}
			}
		}
	}

	reqErr := requestId + " " + givenErr

	if reqParams != "" {
		reqErr = reqErr + " " + reqParams
	}

	ErrorLog.Println(reqErr)

	errRes := defaultErr

	if strings.Contains(reqErr, ErrorForUser) {
		errRes = "Request Id: " + requestId + "\n" + SnipUserError(reqErr)
	}

	http.Error(w, errRes, http.StatusInternalServerError)

	return errors.New(reqErr)
}

func ErrCheck(err error) error {
	if err == nil {
		return nil
	}

	_, file, line, _ := runtime.Caller(1)

	errStr := err.Error() + " " + file + ":" + strconv.Itoa(line)

	return errors.New(errStr)
}
