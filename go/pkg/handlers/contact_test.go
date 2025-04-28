package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostContact(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostContactRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostContactResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostContact(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostContact(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostContact(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchContact(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PatchContactRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchContactResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchContact(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchContact(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchContact(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetContacts(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetContactsRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetContactsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetContacts(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetContacts(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetContacts(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetContactById(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetContactByIdRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetContactByIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetContactById(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetContactById(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetContactById(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteContact(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DeleteContactRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteContactResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteContact(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteContact(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteContact(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DisableContact(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DisableContactRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DisableContactResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DisableContact(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DisableContact(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DisableContact(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}
