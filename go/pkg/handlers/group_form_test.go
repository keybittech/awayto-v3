package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostGroupForm(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostGroupFormRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostGroupFormResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostGroupForm(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostGroupForm(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostGroupForm(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PostGroupFormVersion(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostGroupFormVersionRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostGroupFormVersionResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostGroupFormVersion(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostGroupFormVersion(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostGroupFormVersion(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchGroupForm(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PatchGroupFormRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchGroupFormResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchGroupForm(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchGroupForm(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchGroupForm(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupForms(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupFormsRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupFormsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupForms(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupForms(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupForms(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupFormById(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupFormByIdRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupFormByIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupFormById(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupFormById(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupFormById(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteGroupForm(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DeleteGroupFormRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteGroupFormResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteGroupForm(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteGroupForm(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteGroupForm(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}
