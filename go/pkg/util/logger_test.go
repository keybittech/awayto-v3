package util

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestCustomLogger_Println(t *testing.T) {
	filePath := "/tmp/log_test"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}
	logger := log.New(file, "", log.Ldate|log.Ltime)
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		v []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "Prints a log", fields: fields{logger}, args: args{[]any{"logged message"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &CustomLogger{
				Logger: tt.fields.Logger,
			}
			e.Println(tt.args.v...)
			file.Close()

			readFile, err := os.Open(filePath)
			if err != nil {
				t.Fatal(err)
			}
			defer readFile.Close()

			fileBytes, err := io.ReadAll(readFile)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.HasSuffix(string(fileBytes), "logged message\n") {
				t.Error("CustomLogger_PrintLn() did not write to log file")
			}
		})
	}

	os.Remove(filePath)
}

func BenchmarkCustomLogger_Println(b *testing.B) {
	filePath := "/tmp/log_test"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	logger := log.New(file, "", log.Ldate|log.Ltime)
	errLogger := &CustomLogger{logger}

	reset(b)

	for i := 0; i < b.N; i++ {
		errLogger.Println("test")
	}
}
