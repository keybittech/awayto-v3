package api

import (
	"net/http"
	"reflect"
	"testing"
)

func TestAPI_InitMux(t *testing.T) {
	tests := []struct {
		name string
		a    *API
		want *http.ServeMux
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.InitMux(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.InitMux() = %v, want %v", got, tt.want)
			}
		})
	}
}
