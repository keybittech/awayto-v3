package api

import (
	"net/http"
	"testing"
)

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
