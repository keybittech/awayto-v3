package util

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	LoggingMode = flag.String("log", "", "Debug mode")
	ErrorLog    *CustomLogger
	SockLog     *CustomLogger
)

type CustomLogger struct {
	Logger *log.Logger
}

func init() {
	if LoggingMode == nil || *LoggingMode == "" {
		logLevel := os.Getenv("LOG_LEVEL")
		LoggingMode = &logLevel
	}

	errorLogFile, err := os.OpenFile(os.Getenv("GO_ERROR_LOG"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(errors.New("Failed to open error log" + err.Error()))
	}
	ErrorLog = &CustomLogger{log.New(errorLogFile, "", log.Ldate|log.Ltime)}

	sockLogFile, err := os.OpenFile(os.Getenv("GO_SOCK_LOG"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(errors.New("Failed to open sock log" + err.Error()))
	}
	SockLog = &CustomLogger{log.New(sockLogFile, "", log.Ldate|log.Ltime)}
}

func (e *CustomLogger) Println(v ...any) {
	if v == nil {
		return
	}
	e.Logger.Println(v...)

	if *LoggingMode == "debug" {
		fmt.Println(fmt.Sprintf("DEBUG: %s", v...))
	}
}
