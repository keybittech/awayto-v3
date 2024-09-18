package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	debugMode = flag.Bool("debug", false, "Debug mode")
)

const ForbiddenResponse = `{ "error": { "status": 403 } }`

func WriteErrorToDisk(err error) {
	exec.Command("sh", "-c", fmt.Sprintf("echo \"%s\" >> errors.log", err.Error())).Run()
}

func Debug(message string, args ...interface{}) {
	if *debugMode {
		fmt.Printf(message+"\n", args...)
	}
}

func ErrDebug(err error, args ...string) error {
	argsStr := ""
	for _, str := range args {
		argsStr += str + " "
	}
	return errors.New(fmt.Sprintf("%s %s", err.Error(), argsStr))
}

func ErrCheck(err error) error {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)

		if !ok {
			log.Fatal("error check fatal: ", err)
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

		return errors.New(fmt.Sprintf("%s File: %s; Line: %d; Function: %s; UserError: %s", time.Now().String(), file, line, functionName, err))
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
