package util

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	LoggingMode = flag.String("log", "", "Debug mode")
	AccessLog   *CustomLogger
	AuthLog     *CustomLogger
	DebugLog    *CustomLogger
	ErrorLog    *CustomLogger
	SockLog     *CustomLogger
)

func init() {
	if LoggingMode == nil || *LoggingMode == "" {
		logLevel := os.Getenv("LOG_LEVEL")
		LoggingMode = &logLevel
	}
}

type CustomLogger struct {
	Logger *log.Logger
}

func (e *CustomLogger) Println(v ...any) {
	if v == nil || e == nil || e.Logger == nil {
		return
	}

	e.Logger.Println(v...)

	if *LoggingMode == "debug" && DebugLog != nil {
		fmt.Println("DEBUG:", fmt.Sprint(v...))
		DebugLog.Logger.Printf("DEBUG: %v\n", fmt.Sprint(v...))
	}
}

func makeLogger(prop string) *CustomLogger {
	fileName := os.Getenv(prop)
	if fileName == "" {
		log.Fatal(errors.New("Empty file path for log file " + prop))
	}

	filePath := filepath.Join(os.Getenv("LOG_DIR"), fileName)

	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		log.Fatal(errors.New("Failed to open " + prop + " log" + err.Error()))
	}

	return &CustomLogger{log.New(logFile, "", log.Ldate|log.Ltime)}
}

func MakeLoggers() {
	AccessLog = makeLogger("GO_ACCESS_LOG")
	AuthLog = makeLogger("GO_AUTH_LOG")
	DebugLog = makeLogger("GO_DEBUG_LOG")
	ErrorLog = makeLogger("GO_ERROR_LOG")
	SockLog = makeLogger("GO_SOCK_LOG")
}

func getIp(req *http.Request, ip ...string) (string, error) {
	if len(ip) > 0 {
		return ip[0], nil
	} else {
		reqIp, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			ErrorLog.Println(ErrCheck(err))
			return "", err
		}
		return reqIp, nil
	}
}

func getRequestLogString(req *http.Request) *strings.Builder {
	var sb strings.Builder
	sb.WriteString(req.Method)
	sb.WriteString(" ")
	sb.WriteString(req.URL.Path)
	sb.WriteString(" ")
	sb.WriteString(req.Proto)
	sb.WriteString(" ")
	return &sb
}

// fail2ban-regex /path/to/your/access.log /path/to/filter/http-access.conf

// ^.* (GET|POST|PUT|DELETE|PATCH) .* (<HOST>)$
func WriteAuthRequest(req *http.Request, sub, role string, ip ...string) error {
	reqIp, err := getIp(req, ip...)
	if err != nil {
		return err
	}

	sb := getRequestLogString(req)
	sb.WriteString(sub)
	sb.WriteString(" ")
	sb.WriteString(role)
	sb.WriteString(" ")
	sb.WriteString(reqIp)
	AuthLog.Println(sb.String())
	return nil
}

// f2b regex = ^.* (GET|POST|PUT|DELETE|PATCH) .* (401|403|429) (<HOST>)$
func WriteAccessRequest(req *http.Request, duration int64, statusCode int, ip ...string) error {
	reqIp, err := getIp(req, ip...)
	if err != nil {
		return err
	}

	sb := getRequestLogString(req)
	sb.WriteString(strconv.FormatInt(duration, 10))
	sb.WriteString("ms ")
	sb.WriteString(strconv.Itoa(statusCode))
	sb.WriteString(" ")
	sb.WriteString(reqIp)
	AccessLog.Println(sb.String())
	return nil
}
