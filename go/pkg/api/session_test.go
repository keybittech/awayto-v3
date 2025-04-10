package api

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewSessionMux(t *testing.T) {
	tests := []struct {
		name string
		want *SessionMux
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSessionMux(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSessionMux() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionMux_Handle(t *testing.T) {
	type args struct {
		pattern string
		handler SessionHandler
	}
	tests := []struct {
		name string
		sm   *SessionMux
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sm.Handle(tt.args.pattern, tt.args.handler)
		})
	}
}

func TestSessionMux_HandleFunc(t *testing.T) {
	type args struct {
		pattern string
		handler http.HandlerFunc
	}
	tests := []struct {
		name string
		sm   *SessionMux
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sm.HandleFunc(tt.args.pattern, tt.args.handler)
		})
	}
}

func TestSessionMux_ServeHTTP(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		sm   *SessionMux
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sm.ServeHTTP(tt.args.w, tt.args.r)
		})
	}
}
