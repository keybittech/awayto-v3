package api

import "testing"

func TestAPI_InitProtoHandlers(t *testing.T) {
	tests := []struct {
		name string
		a    *API
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.InitProtoHandlers()
		})
	}
}
