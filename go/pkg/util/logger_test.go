package util

import (
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
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
	logger := log.New(file, "", log.Ldate|log.Ltime)
	errLogger := &CustomLogger{logger}

	reset(b)

	for i := 0; i < b.N; i++ {
		errLogger.Println("test")
	}
	b.StopTimer()
	file.Close()
	os.Remove(filePath)
}

func Test_makeLogger(t *testing.T) {
	type args struct {
		prop string
	}
	tests := []struct {
		name string
		args args
		want *CustomLogger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeLogger(tt.args.prop); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeLogger(%v) = %v, want %v", tt.args.prop, got, tt.want)
			}
		})
	}
}

func TestMakeLoggers(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MakeLoggers()
		})
	}
}

func TestWriteAuthRequest(t *testing.T) {
	type args struct {
		req  *http.Request
		sub  string
		role string
		ip   []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteAuthRequest(tt.args.req, tt.args.sub, tt.args.role, tt.args.ip...); (err != nil) != tt.wantErr {
				t.Errorf("WriteAuthRequest(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.req, tt.args.sub, tt.args.role, tt.args.ip, err, tt.wantErr)
			}
		})
	}
}

func TestWriteAccessRequest(t *testing.T) {
	type args struct {
		req        *http.Request
		duration   int64
		statusCode int
		ip         []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteAccessRequest(tt.args.req, tt.args.duration, tt.args.statusCode, tt.args.ip...); (err != nil) != tt.wantErr {
				t.Errorf("WriteAccessRequest(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.req, tt.args.duration, tt.args.statusCode, tt.args.ip, err, tt.wantErr)
			}
		})
	}
}
