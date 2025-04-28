package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PatchGroupUser(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PatchGroupUserRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchGroupUserResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchGroupUser(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchGroupUser(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchGroupUser(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupUsers(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupUsersRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupUsersResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupUsers(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupUsers(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupUsers(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupUserById(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupUserByIdRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupUserByIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupUserById(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupUserById(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupUserById(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteGroupUser(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DeleteGroupUserRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteGroupUserResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteGroupUser(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteGroupUser(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteGroupUser(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_LockGroupUser(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.LockGroupUserRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.LockGroupUserResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.LockGroupUser(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.LockGroupUser(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.LockGroupUser(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_UnlockGroupUser(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.UnlockGroupUserRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.UnlockGroupUserResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.UnlockGroupUser(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.UnlockGroupUser(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.UnlockGroupUser(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}
