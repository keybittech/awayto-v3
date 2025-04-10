package api

import "testing"

func TestAPI_RedirectHTTP(t *testing.T) {
	type args struct {
		httpPort int
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
			tt.a.RedirectHTTP(tt.args.httpPort)
		})
	}
}
