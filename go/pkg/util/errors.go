package util

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
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

func RequestError(w http.ResponseWriter, givenErr string, ignoreFields []string, pbVal reflect.Value) error {
	requestId := uuid.NewString()
	defaultErr := fmt.Sprintf("%s\nAn error occurred. Please try again later or contact your administrator with the request id provided.", requestId)

	var reqParams string
	if pbVal.IsValid() {
		pbValType := pbVal.Type()
		for j := 0; j < pbVal.NumField(); j++ {
			field := pbVal.Field(j)

			fName := pbValType.Field(j).Name

			if !slices.Contains(ignoreFields, fName) {
				if runtimeName, ok := field.Interface().(string); ok {
					reqParams += fName + "=" + runtimeName + " "
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
