package handlers

import (
	"database/sql"
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostGroupRole(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PostGroupRoleRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostGroupRoleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostGroupRole(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostGroupRole(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostGroupRole(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchGroupRole(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PatchGroupRoleRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchGroupRoleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchGroupRole(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchGroupRole(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchGroupRole(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchGroupRoles(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PatchGroupRolesRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchGroupRolesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchGroupRoles(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchGroupRoles(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchGroupRoles(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupRoles(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetGroupRolesRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupRolesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupRoles(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupRoles(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupRoles(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteGroupRole(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.DeleteGroupRoleRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteGroupRoleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteGroupRole(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteGroupRole(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteGroupRole(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}
