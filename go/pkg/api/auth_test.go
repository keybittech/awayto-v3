package api

import (
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func Test_ValidateToken(t *testing.T) {
	type args struct {
		token    string
		timezone string
		anonIp   string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.UserSession
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, err := ValidateToken(tt.args.token, tt.args.timezone, tt.args.anonIp)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("ValidateToken(%v, %v, %v) error = %v, wantErr %v", tt.args.token, tt.args.timezone, tt.args.anonIp, err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ValidateToken(%v, %v, %v) = %v, want %v", tt.args.token, tt.args.timezone, tt.args.anonIp, got, tt.want)
			// }
		})
	}
}

func TestSetForwardingHeadersAndServe(t *testing.T) {
	type args struct {
		prox *httputil.ReverseProxy
		w    http.ResponseWriter
		r    *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetForwardingHeadersAndServe(tt.args.prox, tt.args.w, tt.args.r)
		})
	}
}

func TestAPI_InitAuthProxy(t *testing.T) {
	type args struct {
		mux *http.ServeMux
	}
	tests := []struct {
		name string
		a    *API
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.InitAuthProxy(tt.args.mux)
		})
	}
}

func TestValidateToken(t *testing.T) {
	type args struct {
		token    string
		timezone string
		anonIp   string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.UserSession
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, err := ValidateToken(tt.args.token, tt.args.timezone, tt.args.anonIp)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("ValidateToken(%v, %v, %v) error = %v, wantErr %v", tt.args.token, tt.args.timezone, tt.args.anonIp, err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ValidateToken(%v, %v, %v) = %v, want %v", tt.args.token, tt.args.timezone, tt.args.anonIp, got, tt.want)
			// }
		})
	}
}
