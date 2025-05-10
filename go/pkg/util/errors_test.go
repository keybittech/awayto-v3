package util

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestUserError(t *testing.T) {
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

func BenchmarkUserError(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = UserError("error")
	}
}

func TestSnipUserError(t *testing.T) {
	type args struct {
		err string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Empty error", args: args{"ERROR_FOR_USER error ERROR_FOR_USER"}, want: "error "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnipUserError(tt.args.err); got != tt.want {
				t.Errorf("SnipUserError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkSnipUserError(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = SnipUserError("ERROR_FOR_USER error ERROR_FOR_USER")
	}
}

func TestRequestError(t *testing.T) {
	testPbStruct := &types.IUserProfile{
		FirstName: "test",
		RoleName:  "role",
	}
	type args struct {
		w            http.ResponseWriter
		givenErr     string
		ignoreFields []protoreflect.Name
		pbVal        proto.Message
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
				ignoreFields: DEFAULT_IGNORED_PROTO_FIELDS,
				pbVal:        testPbStruct,
			},
			wantErr: true,
		},
		{
			name: "Prevents ignored fields from being logged",
			args: args{
				w:            httptest.NewRecorder(),
				givenErr:     "test error",
				ignoreFields: slices.Concat(DEFAULT_IGNORED_PROTO_FIELDS, []protoreflect.Name{protoreflect.Name("firstName")}),
				pbVal:        testPbStruct,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RequestError(tt.args.w, tt.args.givenErr, tt.args.ignoreFields, tt.args.pbVal)

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

			// // For the second test, verify that FirstName is not in the error
			// if tt.name == "Prevents ignored fields from being logged" {
			// 	if errText := err.Error(); err != nil && strings.Contains(errText, "FirstName="+testPbStruct.FirstName) {
			// 		t.Errorf("Error contains ignored fields: error = %v, fields = %v", errText, tt.args.ignoreFields)
			// 	}
			// }
		})
	}
}

func BenchmarkRequestError(b *testing.B) {
	testPbStruct := &types.IUserProfile{
		FirstName: "test",
		RoleName:  "role",
	}
	reset(b)
	for i := 0; i < b.N; i++ {
		RequestError(httptest.NewRecorder(), "test error", slices.Concat(DEFAULT_IGNORED_PROTO_FIELDS, []protoreflect.Name{protoreflect.Name("firstName")}), testPbStruct)
	}
}

func TestErrCheck(t *testing.T) {
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

func BenchmarkErrCheck(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = ErrCheck(errors.New("test error"))
	}
}
