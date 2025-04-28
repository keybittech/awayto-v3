package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostGroupFile(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostGroupFileRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostGroupFileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostGroupFile(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostGroupFile(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostGroupFile(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchGroupFile(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PatchGroupFileRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchGroupFileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchGroupFile(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchGroupFile(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchGroupFile(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupFiles(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupFilesRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupFilesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupFiles(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupFiles(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupFiles(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupFileById(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupFileByIdRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupFileByIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupFileById(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupFileById(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupFileById(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteGroupFile(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DeleteGroupFileRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteGroupFileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteGroupFile(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteGroupFile(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteGroupFile(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}
