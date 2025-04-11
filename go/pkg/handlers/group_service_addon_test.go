package handlers

import (
	"database/sql"
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostGroupServiceAddon(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PostGroupServiceAddonRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostGroupServiceAddonResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostGroupServiceAddon(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostGroupServiceAddon(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostGroupServiceAddon(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupServiceAddons(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetGroupServiceAddonsRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupServiceAddonsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupServiceAddons(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupServiceAddons(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupServiceAddons(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteGroupServiceAddon(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.DeleteGroupServiceAddonRequest
		session *types.UserSession
		tx      *sql.Tx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteGroupServiceAddonResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteGroupServiceAddon(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteGroupServiceAddon(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteGroupServiceAddon(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}
