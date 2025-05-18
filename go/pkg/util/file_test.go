package util

import (
	"os"
	"reflect"
	"testing"
)

func TestGetCleanPath(t *testing.T) {
	type args struct {
		loc  string
		flag int
	}
	tests := []struct {
		name    string
		args    args
		want    *os.File
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCleanPath(tt.args.loc, tt.args.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCleanPath(%v, %v) error = %v, wantErr %v", tt.args.loc, tt.args.flag, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCleanPath(%v, %v) = %v, want %v", tt.args.loc, tt.args.flag, got, tt.want)
			}
		})
	}
}

func TestGetEnvFile(t *testing.T) {
	type args struct {
		envFilePath string
		byteSize    uint16
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEnvFile(tt.args.envFilePath, tt.args.byteSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEnvFile(%v, %v) error = %v, wantErr %v", tt.args.envFilePath, tt.args.byteSize, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEnvFile(%v, %v) = %v, want %v", tt.args.envFilePath, tt.args.byteSize, got, tt.want)
			}
		})
	}
}
