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
