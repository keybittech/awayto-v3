package api

import (
	"net/http"
	"reflect"
	"sync"
	"testing"

	"golang.org/x/time/rate"
)

func TestNewRateLimit(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name  string
		args  args
		want  *sync.Mutex
		want1 map[string]*LimitedClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := NewRateLimit(tt.args.name)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRateLimit(%v) got = %v, want %v", tt.args.name, got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewRateLimit(%v) got1 = %v, want %v", tt.args.name, got1, tt.want1)
			}
		})
	}
}

func TestLimiter(t *testing.T) {
	type args struct {
		mu             *sync.Mutex
		limitedClients map[string]*LimitedClient
		limit          rate.Limit
		burst          int
		identifier     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Limiter(tt.args.mu, tt.args.limitedClients, tt.args.limit, tt.args.burst, tt.args.identifier); got != tt.want {
				t.Errorf("Limiter(%v, %v, %v, %v, %v) = %v, want %v", tt.args.mu, tt.args.limitedClients, tt.args.limit, tt.args.burst, tt.args.identifier, got, tt.want)
			}
		})
	}
}

func TestWriteLimit(t *testing.T) {
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteLimit(tt.args.w)
		})
	}
}
