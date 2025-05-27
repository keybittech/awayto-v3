package api

import (
	"net/http"
	"net/http/httputil"
	"testing"
)

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
			tt.a.InitAuthProxy()
		})
	}
}
