package api

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
)

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

func TestAPI_ValidateTokenMiddleware(t *testing.T) {
	type args struct {
		limit rate.Limit
		burst int
	}
	tests := []struct {
		name string
		a    *API
		args args
		want func(next SessionHandler) http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.a.ValidateTokenMiddleware(tt.args.limit, tt.args.burst); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("API.ValidateTokenMiddleware(%v, %v) = %v, want %v", tt.args.limit, tt.args.burst, got, tt.want)
			// }
		})
	}
}

func TestAPI_GroupInfoMiddleware(t *testing.T) {
	type args struct {
		next SessionHandler
	}
	tests := []struct {
		name string
		a    *API
		args args
		want SessionHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.GroupInfoMiddleware(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.GroupInfoMiddleware(%v) = %v, want %v", tt.args.next, got, tt.want)
			}
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

func TestCacheWriter_Write(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		cw      *CacheWriter
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cw.Write(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheWriter.Write(%v) error = %v, wantErr %v", tt.args.data, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CacheWriter.Write(%v) = %v, want %v", tt.args.data, got, tt.want)
			}
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
