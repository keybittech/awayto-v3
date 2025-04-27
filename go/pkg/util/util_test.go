package util

import (
	"bytes"
	"database/sql"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestNewNullString(t *testing.T) {
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

func BenchmarkNewNullString(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = NewNullString("test error")
	}
}

func TestIsUUID(t *testing.T) {
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

func BenchmarkIsUUID(b *testing.B) {
	str := "00000000-0000-0000-0000-000000000000"
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = IsUUID(str)
	}
}

func BenchmarkIsUUIDNegative(b *testing.B) {
	str := "test"
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = IsUUID(str)
	}
}

func TestIsEpoch(t *testing.T) {
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

func BenchmarkIsEpoch(b *testing.B) {
	var goodId = "0123456789"
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = IsEpoch(goodId)
	}
}

func BenchmarkIsEpochNegative(b *testing.B) {
	var badId = "test"
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = IsEpoch(badId)
	}
}

func TestPaddedLen(t *testing.T) {
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

func BenchmarkPaddedLen(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = PaddedLen(5, 3)
	}
}

func BenchmarkPaddedLenNegative(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = PaddedLen(-5, -3)
	}
}

func TestEnvFile(t *testing.T) {
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

func BenchmarkEnvFile(b *testing.B) {
	tmpDirName := "test-project-dir"
	tmpDir, err := os.MkdirTemp("", tmpDirName)
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	oldDir := os.Getenv("PROJECT_DIR")
	os.Setenv("PROJECT_DIR", tmpDir)
	defer func() { os.Setenv("PROJECT_DIR", oldDir) }()
	filePath := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(filePath, []byte("test content\n\n"), 0644)

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = EnvFile("test.txt")
	}
}

func TestAnonIp(t *testing.T) {
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

func BenchmarkAnonIp(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = AnonIp("1.1.1.1")
	}
}

func BenchmarkAnonIpNegative(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = AnonIp("1.1.1")
	}
}

func TestStringIn(t *testing.T) {
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

func BenchmarkStringIn(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = StringIn("test", []string{"test"})
	}
}

func TestStringOut(t *testing.T) {
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

func BenchmarkStringOut(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = StringOut("test", []string{"test", "case"})
	}
}

func TestExeTime(t *testing.T) {
	var buf bytes.Buffer
	originalLogger := AccessLog
	AccessLog = &CustomLogger{log.New(&buf, "", 0)}
	defer func() { AccessLog = originalLogger }()

	t.Run("standard logging", func(t *testing.T) {
		testName := "testFunction"
		start, endFunc := ExeTime(testName)
		testInfo := "test completed"
		endFunc(start, testInfo)

		logMsg := buf.String()
		if !strings.Contains(logMsg, testName) {
			t.Errorf("Log doesn't contain function name and execution time text, got: %s", logMsg)
		}
		if !strings.Contains(logMsg, testInfo) {
			t.Errorf("Log doesn't contain the info text '%s', got: %s", testInfo, logMsg)
		}
	})
}

func BenchmarkExeTime(b *testing.B) {
	var buf bytes.Buffer
	originalLogger := ErrorLog
	ErrorLog = &CustomLogger{log.New(&buf, "", 0)}
	defer func() { ErrorLog = originalLogger }()
	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = ExeTime("testFunction")
	}
}

// func BenchmarkExeTimeDeferFunc(b *testing.B) {
// 	var buf bytes.Buffer
// 	originalLogger := ErrorLog
// 	ErrorLog = &CustomLogger{log.New(&buf, "", 0)}
// 	defer func() { ErrorLog = originalLogger }()
// 	reset(b)
// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		start, deferFunc := ExeTime("testFunction")
// 		b.StartTimer()
// 		deferFunc(start, "test")
// 	}
// }

func TestWriteSigned(t *testing.T) {
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

func BenchmarkWriteSigned(b *testing.B) {
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = WriteSigned("a", "b")
	}
}

func TestVerifySigned(t *testing.T) {
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

func BenchmarkVerifySigned(b *testing.B) {
	signedValue := WriteSigned("a", "b")
	reset(b)
	for i := 0; i < b.N; i++ {
		_ = VerifySigned("a", signedValue)
	}
}

func TestNullConn_Read(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.Read(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("NullConn.Read(%v) error = %v, wantErr %v", tt.args.b, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NullConn.Read(%v) = %v, want %v", tt.args.b, got, tt.want)
			}
		})
	}
}

func TestNullConn_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("NullConn.Write(%v) error = %v, wantErr %v", tt.args.b, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NullConn.Write(%v) = %v, want %v", tt.args.b, got, tt.want)
			}
		})
	}
}

func TestNullConn_Close(t *testing.T) {
	tests := []struct {
		name    string
		n       NullConn
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Close(); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNullConn_LocalAddr(t *testing.T) {
	tests := []struct {
		name string
		n    NullConn
		want net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.LocalAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NullConn.LocalAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullConn_RemoteAddr(t *testing.T) {
	tests := []struct {
		name string
		n    NullConn
		want net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.RemoteAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NullConn.RemoteAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullConn_SetDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.SetDeadline(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.SetDeadline(%v) error = %v, wantErr %v", tt.args.t, err, tt.wantErr)
			}
		})
	}
}

func TestNullConn_SetReadDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.SetReadDeadline(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.SetReadDeadline(%v) error = %v, wantErr %v", tt.args.t, err, tt.wantErr)
			}
		})
	}
}

func TestNullConn_SetWriteDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.SetWriteDeadline(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.SetWriteDeadline(%v) error = %v, wantErr %v", tt.args.t, err, tt.wantErr)
			}
		})
	}
}

func TestNewNullConn(t *testing.T) {
	tests := []struct {
		name string
		want net.Conn
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullConn(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullConn() = %v, want %v", got, tt.want)
			}
		})
	}
}
