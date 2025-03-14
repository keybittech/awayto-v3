package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	LoggingMode    = flag.String("log", "", "Debug mode")
	ErrorLog       *ErrLog
	TitleCase      cases.Caser
	DefaultPadding int
	SigningToken   []byte

	LOGIN_SIGNATURE_NAME = "login_signature_name"
)

func init() {
	var err error
	SigningToken, err = os.ReadFile(os.Getenv("SIGNING_TOKEN_FILE"))
	if err != nil {
		println("Failed to get signing token")
		log.Fatal(err)
	}

	file, err := os.OpenFile(os.Getenv("GO_ERROR_LOG"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		println("Failed to open error log")
		log.Fatal(err)
	}
	ErrorLog = &ErrLog{log.New(file, "", log.Ldate|log.Ltime)}
	TitleCase = cases.Title(language.Und)
	DefaultPadding = 5

	if LoggingMode == nil || *LoggingMode == "" {
		logLevel := os.Getenv("LOG_LEVEL")
		LoggingMode = &logLevel
	}
}

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

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func IsUUID(id string) bool {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidPattern.MatchString(id)
}

func IsEpoch(id string) bool {
	_, err := strconv.ParseInt(id, 10, 64)
	return err == nil
}

func PaddedLen(padTo int, length int) string {
	strLen := strconv.Itoa(length)
	for len(strLen) < padTo {
		strLen = "0" + strLen
	}
	return strLen
}

func EnvFile(loc string) (string, error) {
	envFile, err := os.ReadFile(loc)
	if err != nil {
		return "", ErrCheck(err)
	}

	return strings.Trim(string(envFile), "\n"), nil
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
	fmt.Printf("beginning execution for %s\n", name)
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

func WriteSigned(name, unsignedValue string) string {
	mac := hmac.New(sha256.New, SigningToken)
	mac.Write([]byte(name))
	mac.Write([]byte(unsignedValue))
	signature := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature) + unsignedValue
}

func VerifySigned(name, signedValue string) error {
	if len(signedValue) < sha256.Size {
		return errors.New("signed value too small")
	}

	signatureEncoded := signedValue[:base64.StdEncoding.EncodedLen(sha256.Size)]
	signature, err := base64.StdEncoding.DecodeString(signatureEncoded)
	if err != nil {
		return errors.New("invalid base64 signature encoding")
	}

	value := signedValue[base64.StdEncoding.EncodedLen(sha256.Size):]

	mac := hmac.New(sha256.New, SigningToken)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return errors.New("invalid signature equality")
	}

	return nil
}
