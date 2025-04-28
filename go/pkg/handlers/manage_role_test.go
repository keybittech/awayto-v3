package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostManageRoles(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostManageRolesRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostManageRolesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostManageRoles(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostManageRoles(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostManageRoles(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchManageRoles(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PatchManageRolesRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchManageRolesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchManageRoles(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchManageRoles(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchManageRoles(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetManageRoles(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetManageRolesRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetManageRolesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetManageRoles(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetManageRoles(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetManageRoles(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteManageRoles(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DeleteManageRolesRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteManageRolesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteManageRoles(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteManageRoles(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteManageRoles(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}
