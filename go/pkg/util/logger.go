package util

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

type CustomLogger struct {
	Logger *log.Logger
}

func (e *CustomLogger) Println(v ...any) {
	if v == nil {
		return
	}

	e.Logger.Println(v...)

	if *LoggingMode == "debug" {
		println(fmt.Sprintf("DEBUG: %v", fmt.Sprint(v...)))
		DebugLog.Logger.Printf("DEBUG: %v\n", fmt.Sprint(v...))
	}
}

func makeLogger(prop string) *CustomLogger {
	fileName := os.Getenv(prop)
	if fileName == "" {
		log.Fatal(errors.New("Empty file path for log file " + prop))
	}

	filePath := os.Getenv("GO_LOG_DIR") + "/" + fileName

	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(errors.New("Failed to open " + prop + " log" + err.Error()))
	}

	return &CustomLogger{log.New(logFile, "", log.Ldate|log.Ltime)}
}

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

	AccessLog = makeLogger("GO_ACCESS_LOG")
	AuthLog = makeLogger("GO_AUTH_LOG")
	DebugLog = makeLogger("GO_DEBUG_LOG")
	ErrorLog = makeLogger("GO_ERROR_LOG")
	SockLog = makeLogger("GO_SOCK_LOG")
}
