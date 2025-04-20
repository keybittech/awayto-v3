package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostUserProfile(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PostUserProfileRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostUserProfileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostUserProfile(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostUserProfile(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostUserProfile(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchUserProfile(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PatchUserProfileRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchUserProfileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchUserProfile(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchUserProfile(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchUserProfile(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetUserProfileDetails(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetUserProfileDetailsRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetUserProfileDetailsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetUserProfileDetails(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetUserProfileDetails(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetUserProfileDetails(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetUserProfileDetailsBySub(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetUserProfileDetailsBySubRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetUserProfileDetailsBySubResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetUserProfileDetailsBySub(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetUserProfileDetailsBySub(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetUserProfileDetailsBySub(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetUserProfileDetailsById(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetUserProfileDetailsByIdRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetUserProfileDetailsByIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetUserProfileDetailsById(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetUserProfileDetailsById(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetUserProfileDetailsById(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_DisableUserProfile(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.DisableUserProfileRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DisableUserProfileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DisableUserProfile(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DisableUserProfile(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DisableUserProfile(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_ActivateProfile(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.ActivateProfileRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.ActivateProfileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.ActivateProfile(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.ActivateProfile(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.ActivateProfile(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeactivateProfile(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.DeactivateProfileRequest
		session *types.UserSession
		tx      *clients.PoolTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeactivateProfileResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeactivateProfile(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeactivateProfile(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeactivateProfile(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}
