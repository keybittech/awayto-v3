package api

import (
	"net/http"
	"testing"
)

func TestStaticGzip_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		g       StaticGzip
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("StaticGzip.Write(%v) error = %v, wantErr %v", tt.args.b, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StaticGzip.Write(%v) = %v, want %v", tt.args.b, got, tt.want)
			}
		})
	}
}

func TestStaticRedirect_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		sr      *StaticRedirect
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sr.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("StaticRedirect.Write(%v) error = %v, wantErr %v", tt.args.b, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StaticRedirect.Write(%v) = %v, want %v", tt.args.b, got, tt.want)
			}
		})
	}
}

func TestStaticRedirect_WriteHeader(t *testing.T) {
	type args struct {
		code int
	}
	tests := []struct {
		name string
		sr   *StaticRedirect
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sr.WriteHeader(tt.args.code)
		})
	}
}

func TestAPI_InitStatic(t *testing.T) {
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
			tt.a.InitStatic(tt.args.mux)
		})
	}
}
