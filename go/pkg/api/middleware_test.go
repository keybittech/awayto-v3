package api

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
)

func TestAPI_AccessRequestMiddleware(t *testing.T) {
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		a    *API
		args args
		want http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.AccessRequestMiddleware(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.AccessRequestMiddleware(%v) = %v, want %v", tt.args.next, got, tt.want)
			}
		})
	}
}

func TestAPI_LimitMiddleware(t *testing.T) {
	type args struct {
		limit rate.Limit
		burst int
	}
	tests := []struct {
		name string
		a    *API
		args args
		want func(next http.HandlerFunc) http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.a.LimitMiddleware(tt.args.limit, tt.args.burst); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("API.LimitMiddleware(%v, %v) = %v, want %v", tt.args.limit, tt.args.burst, got, tt.want)
			// }
		})
	}
}

func TestAPI_SiteRoleCheckMiddleware(t *testing.T) {
	type args struct {
		opts *util.HandlerOptions
	}
	tests := []struct {
		name string
		a    *API
		args args
		want func(SessionHandler) SessionHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.a.SiteRoleCheckMiddleware(tt.args.opts); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("API.SiteRoleCheckMiddleware(%v) = %v, want %v", tt.args.opts, got, tt.want)
			// }
		})
	}
}

func TestAPI_CacheMiddleware(t *testing.T) {
	type args struct {
		opts *util.HandlerOptions
	}
	tests := []struct {
		name string
		a    *API
		args args
		want func(SessionHandler) SessionHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.CacheMiddleware(tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				// t.Errorf("API.CacheMiddleware(%v) = %v, want %v", tt.args.opts, got, tt.want)
			}
		})
	}
}

func Test_responseCodeWriter_WriteHeader(t *testing.T) {
	type args struct {
		code int
	}
	tests := []struct {
		name string
		rw   *responseCodeWriter
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.rw.WriteHeader(tt.args.code)
		})
	}
}
