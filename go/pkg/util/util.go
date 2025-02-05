package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	DebugMode = flag.Bool("debug", false, "Debug mode")
	ErrorLog  *log.Logger
	TitleCase cases.Caser
)

func init() {
	file, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	ErrorLog = log.New(file, "ERROR: ", log.Ldate|log.Ltime)
	TitleCase = cases.Title(language.Und)
}

const ForbiddenResponse = `{ "error": { "status": 403 } }`
const InternalErrorResponse = `{ "error": { "status": 500 } }`
const ErrorForUser = "ERROR_FOR_USER"

func UserError(err string) error {
	return errors.New(fmt.Sprintf("%s %s %s", ErrorForUser, err, ErrorForUser))
}

func SnipUserError(err string) string {
	return strings.TrimSpace(strings.Split(err, ErrorForUser)[1])
}

func ErrCheck(err error) error {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)

		if !ok {
			log.Fatal(fmt.Sprintf("fatal error when checking err %s", err))
		}

		function := runtime.FuncForPC(pc)
		functionName := ""
		if function != nil {
			functionParts := strings.SplitAfterN(function.Name(), ".", -1)
			if len(functionParts) > 2 {
				functionName = functionParts[2]
			} else {
				functionName = function.Name()
			}
		}

		errStr := err.Error()

		if strings.Contains(errStr, "SysErrMsg:") {
			errStr = "SPAWNED FROM: " + errStr
		}

		return errors.New(fmt.Sprintf("File: %s; Line: %d; Function: %s; SysErrMsg: %s", file, line, functionName, errStr))
	}

	return nil
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

func ParseJWT(token string) (map[string]interface{}, map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid JWT, expected 3 parts but got %d", len(parts))
	}

	headerBytes, err := Base64UrlDecode(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode header: %v", err)
	}

	payloadBytes, err := Base64UrlDecode(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode payload: %v", err)
	}

	var header, payload map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal header: %v", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	return header, payload, nil
}
