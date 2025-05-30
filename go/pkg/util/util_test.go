package util

import (
	"database/sql"
	"reflect"
	"sync"
	"testing"
)

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	ParseEnv()
	m.Run()
}

func BenchmarkMapSyncWrite(b *testing.B) {
	var m sync.Map
	reset(b)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Store(i, i)
			i++
		}
	})
}

func BenchmarkMapMutexWrite(b *testing.B) {
	var mu sync.Mutex
	m := make(map[int]int)
	reset(b)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mu.Lock()
			m[i] = i
			mu.Unlock()
			i++
		}
	})
}

func BenchmarkNoop(b *testing.B) {
	var noop = func() {}
	reset(b)
	for b.Loop() {
		noop()
	}
}

func BenchmarkNoopNil(b *testing.B) {
	var noop func() = nil
	reset(b)
	for b.Loop() {
		if noop != nil {
			// Noop
		}
	}
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
	for b.Loop() {
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
	for b.Loop() {
		_ = IsUUID(str)
	}
}

func BenchmarkIsUUIDNegative(b *testing.B) {
	str := "test"
	reset(b)
	for b.Loop() {
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
	for b.Loop() {
		_ = IsEpoch(goodId)
	}
}

func BenchmarkIsEpochNegative(b *testing.B) {
	var badId = "test"
	reset(b)
	for b.Loop() {
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
	for b.Loop() {
		_ = PaddedLen(5, 3)
	}
}

func BenchmarkPaddedLenNegative(b *testing.B) {
	reset(b)
	for b.Loop() {
		_ = PaddedLen(-5, -3)
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
	for b.Loop() {
		_ = AnonIp("1.1.1.1")
	}
}

func BenchmarkAnonIpNegative(b *testing.B) {
	reset(b)
	for b.Loop() {
		_ = AnonIp("1.1.1")
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
	for b.Loop() {
		_ = StringOut("test", []string{"test", "case"})
	}
}

// func BenchmarkExeTimeDeferFunc(b *testing.B) {
// 	var buf bytes.Buffer
// 	originalLogger := ErrorLog
// 	ErrorLog = &CustomLogger{log.New(&buf, "", 0)}
// 	defer func() { ErrorLog = originalLogger }()
// 	reset(b)
// 	for b.Loop() {
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
			signedValue, err := WriteSigned(tt.args.name, tt.args.unsignedValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySigned.WriteSigned() error = %v, wantErr %v", err, tt.wantErr)
			}

			_, err = VerifySigned(tt.args.name, signedValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySigned() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkWriteSigned(b *testing.B) {
	reset(b)
	for b.Loop() {
		_, _ = WriteSigned("a", "b")
	}
}

func TestVerifySigned(t *testing.T) {
	emptySigned, _ := WriteSigned("", "")
	validSigned, _ := WriteSigned("validName", "validValue")
	type args struct {
		name        string
		signedValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "no value", args: args{"", emptySigned}, wantErr: false},
		{name: "valid signed value", args: args{"validName", validSigned}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := VerifySigned(tt.args.name, tt.args.signedValue); (err != nil) != tt.wantErr {
				t.Errorf("VerifySigned() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkVerifySigned(b *testing.B) {
	signedValue, _ := WriteSigned("a", "b")
	reset(b)
	for b.Loop() {
		_, _ = VerifySigned("a", signedValue)
	}
}

func TestStringsToSiteRoles(t *testing.T) {
	type args struct {
		roles []string
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringsToSiteRoles(tt.args.roles); got != tt.want {
				t.Errorf("StringsToBitmask(%v) = %v, want %v", tt.args.roles, got, tt.want)
			}
		})
	}
}

func TestAtoi32(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Atoi32(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Atoi32(%v) error = %v, wantErr %v", tt.args.s, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Atoi32(%v) = %v, want %v", tt.args.s, got, tt.want)
			}
		})
	}
}

func TestItoi32(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Itoi32(tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("Itoi32(%v) error = %v, wantErr %v", tt.args.i, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Itoi32(%v) = %v, want %v", tt.args.i, got, tt.want)
			}
		})
	}
}
