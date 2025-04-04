package util

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestUtilErrLog_Println(t *testing.T) {
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
			e := &ErrLog{
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
				t.Error("ErrLog_PrintLn() did not write to log file")
			}
		})
	}

	os.Remove(filePath)
}

func TestUtilUserError(t *testing.T) {
	type args struct {
		err string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{name: "Empty error", args: args{""}, wantErr: errors.New("ERROR_FOR_USER  ERROR_FOR_USER")},
		{name: "Regular error", args: args{"error"}, wantErr: errors.New("ERROR_FOR_USER error ERROR_FOR_USER")},
		{name: "Negative page size", args: args{"923hf923ghf923"}, wantErr: errors.New("ERROR_FOR_USER 923hf923ghf923 ERROR_FOR_USER")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UserError(tt.args.err); err.Error() != tt.wantErr.Error() {
				t.Errorf("UserError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUtilSnipUserError(t *testing.T) {
	type args struct {
		err string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Empty error", args: args{"ERROR_FOR_USER error ERROR_FOR_USER"}, want: "error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnipUserError(tt.args.err); got != tt.want {
				t.Errorf("SnipUserError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilRequestError(t *testing.T) {
	defaultIgnored := []string{"state", "sizeCache", "unknownFields"}
	testPbStruct := &types.IUserProfile{
		FirstName: "john",
		RoleName:  "role_name",
	}
	type args struct {
		w            http.ResponseWriter
		givenErr     string
		ignoreFields []string
		pbVal        reflect.Value
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Returns an error with provided data",
			args: args{
				w:            httptest.NewRecorder(),
				givenErr:     "test error",
				ignoreFields: defaultIgnored,
				pbVal:        reflect.ValueOf(testPbStruct).Elem(),
			},
			wantErr: true,
		},
		{
			name: "Prevents ignored fields from being logged",
			args: args{
				w:            httptest.NewRecorder(),
				givenErr:     "test error",
				ignoreFields: slices.Concat(defaultIgnored, []string{"FirstName"}),
				pbVal:        reflect.ValueOf(testPbStruct).Elem(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequestError(tt.args.w, tt.args.givenErr, tt.args.ignoreFields, tt.args.pbVal)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequestError() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the response
			response := tt.args.w.(*httptest.ResponseRecorder)

			// Check the HTTP status code
			if response.Code != http.StatusInternalServerError {
				t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, response.Code)
			}

			// Check the error message contains request ID
			if !strings.Contains(response.Body.String(), "Request Id:") &&
				!strings.Contains(response.Body.String(), "An error occurred") {
				t.Errorf("Response body doesn't contain expected error message: %s", response.Body.String())
			}

			// For the second test, verify that FirstName is not in the error
			if tt.name == "Prevents ignored fields from being logged" {
				if errText := err.Error(); err != nil && strings.Contains(errText, testPbStruct.FirstName) {
					t.Errorf("Error contains ignored fields: error = %v, fields = %v", errText, tt.args.ignoreFields)
				}
			}
		})
	}
}

func TestUtilErrCheck(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "nil error", args: args{err: nil}, wantErr: false},
		{name: "non-nil error", args: args{err: errors.New("test error")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ErrCheck(tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("ErrCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUtilNewNullString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want sql.NullString
	}{
		{name: "empty string", args: args{s: ""}, want: sql.NullString{String: "", Valid: false}},
		{name: "non-empty string", args: args{s: "test"}, want: sql.NullString{String: "test", Valid: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullString(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilIsUUID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Normal uuid", args: args{id: "00000000-0000-0000-0000-000000000000"}, want: true},
		{name: "Non uuid", args: args{id: "test"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUUID(tt.args.id); got != tt.want {
				t.Errorf("IsUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilIsEpoch(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Normal epoch", args: args{id: "0123456789"}, want: true},
		{name: "Non epoch", args: args{id: "test"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEpoch(tt.args.id); got != tt.want {
				t.Errorf("IsEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilPaddedLen(t *testing.T) {
	type args struct {
		padTo  int
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "No padding", args: args{0, 4}, want: "4"},
		{name: "Normal padding", args: args{5, 3}, want: "00003"},
		{name: "Negative padding", args: args{-5, 3}, want: "3"},
		{name: "Negative both", args: args{-5, -3}, want: "-3"},
		{name: "Negative length", args: args{5, -3}, want: "000-3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PaddedLen(tt.args.padTo, tt.args.length); got != tt.want {
				t.Errorf("PaddedLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilEnvFile(t *testing.T) {
	tmpDirName := "test-project-dir"
	tmpDir, err := os.MkdirTemp("", tmpDirName)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	os.Setenv("PROJECT_DIR", tmpDir)
	type args struct {
		loc     string
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "valid file", args: args{loc: "test.txt", content: "test content\n\n"}, want: "test content", wantErr: false},
		{name: "empty file", args: args{loc: "empty.txt", content: ""}, want: "", wantErr: false},
		{name: "nonexistent file", args: args{loc: "nonexistent.txt", content: ""}, wantErr: true},
	}

	for _, tc := range tests {
		if tc.name != "nonexistent file" {
			filePath := filepath.Join(tmpDir, tc.args.loc)
			err := os.WriteFile(filePath, []byte(tc.args.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", tc.args.loc, err)
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvFile(tt.args.loc)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EnvFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilAnonIp(t *testing.T) {
	type args struct {
		ipAddr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "valid ip", args: args{"1.1.1.1"}, want: "1.1.1.0"},
		{name: "bad ip", args: args{"1.1.1"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AnonIp(tt.args.ipAddr); got != tt.want {
				t.Errorf("AnonIp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilStringIn(t *testing.T) {
	type args struct {
		s  string
		ss []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "has string", args: args{"test", []string{"test"}}, want: true},
		{name: "does not have string", args: args{"test", []string{}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringIn(tt.args.s, tt.args.ss); got != tt.want {
				t.Errorf("StringIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilStringOut(t *testing.T) {
	type args struct {
		s  string
		ss []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "no value", args: args{"test", []string{}}, want: []string{}},
		{name: "remove single", args: args{"test", []string{"test"}}, want: []string{}},
		{name: "remove from many", args: args{"test", []string{"test", "case"}}, want: []string{"case"}},
		{name: "no string in", args: args{"test", []string{"none"}}, want: []string{"none"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringOut(tt.args.s, tt.args.ss); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringOut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilExeTime(t *testing.T) {
	var buf bytes.Buffer
	originalLogger := ErrorLog
	ErrorLog = &ErrLog{log.New(&buf, "", 0)}
	defer func() { ErrorLog = originalLogger }()

	t.Run("standard logging", func(t *testing.T) {
		testName := "testFunction"
		endFunc := ExeTime(testName)

		if !strings.Contains(buf.String(), "beginning execution for "+testName) {
			t.Errorf("Expected log to contain 'beginning execution for %s', got: %s", testName, buf.String())
		}

		buf.Reset()

		testInfo := "test completed"
		endFunc(testInfo)

		logMsg := buf.String()
		if !strings.Contains(logMsg, testName+" execution time:") {
			t.Errorf("Log doesn't contain function name and execution time text, got: %s", logMsg)
		}
		if !strings.Contains(logMsg, testInfo) {
			t.Errorf("Log doesn't contain the info text '%s', got: %s", testInfo, logMsg)
		}
	})
}

func TestUtilWriteSigned(t *testing.T) {
	type args struct {
		name          string
		unsignedValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "no value", args: args{"", ""}, wantErr: false},
		{name: "values", args: args{"bb", "aa"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signedValue := WriteSigned(tt.args.name, tt.args.unsignedValue)

			err := VerifySigned(tt.args.name, signedValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySigned() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUtilVerifySigned(t *testing.T) {
	type args struct {
		name        string
		signedValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "no value", args: args{"", WriteSigned("", "")}, wantErr: false},
		{name: "valid signed value", args: args{"validName", WriteSigned("validName", "validValue")}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifySigned(tt.args.name, tt.args.signedValue); (err != nil) != tt.wantErr {
				t.Errorf("VerifySigned() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
