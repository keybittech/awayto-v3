package util

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	LoggingMode = flag.String("log", "", "Debug mode")
	ErrorLog    *ErrLog
	TitleCase   cases.Caser
)

const ForbiddenResponse = `{ "error": { "status": 403 } }`
const InternalErrorResponse = `{ "error": { "status": 500 } }`
const ErrorForUser = "ERROR_FOR_USER"

type ErrLog struct {
	Logger *log.Logger
}

func (e *ErrLog) Println(v ...any) {
	if v == nil {
		return
	}
	e.Logger.Println(v...)

	if *LoggingMode == "debug" {
		fmt.Println(fmt.Sprintf("DEBUG: %s", v...))
	}
}

func init() {
	file, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	ErrorLog = &ErrLog{log.New(file, "", log.Ldate|log.Ltime)}
	TitleCase = cases.Title(language.Und)
}

func UserError(err string) error {
	return errors.New(fmt.Sprintf("%s %s %s", ErrorForUser, err, ErrorForUser))
}

func SnipUserError(err string) string {
	return strings.TrimSpace(strings.Split(err, ErrorForUser)[1])
}

func RequestError(w http.ResponseWriter, requestId, givenErr string, ignoreFields []string, pbVal reflect.Value) error {
	defaultErr := fmt.Sprintf("%s\nAn error occurred. Please try again later or contact your administrator with the request id provided.", requestId)

	var reqParams string
	if pbVal.IsValid() {
		pbValType := pbVal.Type()
		for j := 0; j < pbVal.NumField(); j++ {
			field := pbVal.Field(j)

			fName := pbValType.Field(j).Name

			if !slices.Contains(ignoreFields, fName) {
				reqParams += fmt.Sprintf("%s=%v", fName, field.Interface()) + " "
			}
		}
	}

	reqErr := errors.New(fmt.Sprintf("%s %s", requestId, givenErr))

	if reqParams != "" {
		reqErr = errors.New(fmt.Sprintf("%s %s", reqErr, reqParams))
	}

	ErrorLog.Println(reqErr)

	errRes := defaultErr

	if strings.Contains(reqErr.Error(), ErrorForUser) {
		errRes = fmt.Sprintf("Request Id: %s\n%s", requestId, SnipUserError(reqErr.Error()))
	}

	http.Error(w, errRes, http.StatusInternalServerError)

	return reqErr
}

func ErrCheck(err error) error {
	if err == nil {
		return nil
	}

	_, file, line, _ := runtime.Caller(1)

	return errors.New(fmt.Sprintf("%s %s:%d", err.Error(), file, line))
}

func CastSlice[T any](items []interface{}) ([]T, bool) {
	results := make([]T, len(items))

	for i, item := range items {
		val, ok := item.(T)
		if !ok {
			return nil, false
		}
		results[i] = val
	}

	return results, true
}

func AnonIp(ipAddr string) string {
	ipParts := strings.Split(ipAddr, ".")
	if len(ipParts) != 4 {
		return ""
	}
	ipParts[3] = "0"
	return strings.Join(ipParts, ".")
}

func StringIn(s string, ss []string) bool {
	for _, sv := range ss {
		if sv == s {
			return true
		}
	}
	return false
}

func StringOut(s string, ss []string) []string {
	var ns []string
	for _, cs := range ss {
		if cs == s {
			continue
		}
		ns = append(ns, cs)
	}
	return ns
}

func ExeTime(name string) func(info string) {
	start := time.Now()
	return func(info string) {
		fmt.Printf("%s execution time: %v %s\n", name, time.Since(start), info)
	}
}

func ToPascalCase(input string) string {
	words := strings.Split(input, "-")
	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word)
	}
	return strings.Join(words, "")
}

func Base64UrlDecode(str string) ([]byte, error) {
	s := strings.ReplaceAll(str, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.StdEncoding.DecodeString(s)
}
