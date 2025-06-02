package util

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	AccessLog *CustomLogger
	AuthLog   *CustomLogger
	DebugLog  *CustomLogger
	ErrorLog  *CustomLogger
	SockLog   *CustomLogger
)

type CustomLogger struct {
	*log.Logger
}

func debugLog(val string) {
	if E_LOG_LEVEL == "debug" {
		fmt.Println("DEBUG:", val)
	}
}

func (e *CustomLogger) Printf(format string, v ...any) {
	debugLog(fmt.Sprintf(format, v...))
	e.Logger.Printf(format, v...)
}

func (e *CustomLogger) Println(v ...any) {
	debugLog(fmt.Sprint(v...))
	e.Logger.Println(v...)
}

func makeLogger(prop string) *CustomLogger {
	loc := envVarStrs[prop]
	if loc == "" {
		log.Fatalf("Empty file path for log file %s", prop)
	}

	cleanLoc := filepath.Clean(loc)

	if strings.Contains(cleanLoc, "..") {
		log.Fatal("invalid file path: path traversal attempt detected")
	}

	logFilePath := filepath.Join(E_LOG_DIR, loc)

	if !strings.HasPrefix(filepath.Clean(logFilePath), filepath.Clean(E_LOG_DIR)) {
		log.Fatalf("invalid file path: path is outside of log directory, %s", logFilePath)
	}

	println("Creating a log file at", logFilePath)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		log.Fatalf("Failed to open %s log %v", prop, err)
	}

	return &CustomLogger{Logger: log.New(logFile, "", log.Ldate|log.Ltime)}
}

func makeLoggers() {
	AccessLog = makeLogger("GO_ACCESS_LOG")
	AuthLog = makeLogger("GO_AUTH_LOG")
	DebugLog = makeLogger("GO_DEBUG_LOG")
	ErrorLog = makeLogger("GO_ERROR_LOG")
	SockLog = makeLogger("GO_SOCK_LOG")
}

func getIp(req *http.Request, ip ...string) string {
	if len(ip) > 0 {
		return ip[0]
	} else {
		reqIp, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			return req.RemoteAddr
		}
		return reqIp
	}
}

func writeRequestLogString(sb *strings.Builder, req *http.Request) {
	sb.WriteString(req.Method)
	sb.WriteString(" ")
	sb.WriteString(req.URL.Path)
	sb.WriteString(" ")
	sb.WriteString(req.Proto)
	sb.WriteString(" ")
}

// f2b regex = ^.* (GET|POST|PUT|DELETE|PATCH) .* (<HOST>)$
func WriteAuthRequest(req *http.Request, sub, role string, ip ...string) {
	reqIp := getIp(req, ip...)
	var sb strings.Builder
	writeRequestLogString(&sb, req)
	sb.WriteString(sub)
	sb.WriteString(" ")
	sb.WriteString(role)
	sb.WriteString(" ")
	sb.WriteString(reqIp)
	AuthLog.Println(sb.String())
}

// f2b regex = ^.* (GET|POST|PUT|DELETE|PATCH) .* (401|403|429) (<HOST>)$
func WriteAccessRequest(req *http.Request, duration int64, statusCode int, ip ...string) {
	reqIp := getIp(req, ip...)
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(statusCode))
	sb.WriteString(" ")
	writeRequestLogString(&sb, req)
	sb.WriteString(strconv.FormatInt(duration, 10))
	sb.WriteString("ms ")
	sb.WriteString(reqIp)
	AccessLog.Println(sb.String())
}

var runTimers bool

func RunTimer(values ...any) func() {
	if !runTimers {
		return func() {}
	}
	pc, _, _, _ := runtime.Caller(1)
	start := time.Now()

	return func() {
		name := runtime.FuncForPC(pc).Name()
		name = name[strings.LastIndex(name, ".")+1:]
		duration := time.Since(start)
		if r := recover(); r != nil {
			DebugLog.Println(name, "function panicked after", duration, values, r)
			panic(r)
		} else {
			DebugLog.Println(name, "function completed successfully in", duration, values)
		}
	}
}
