package main

import (
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/api"
)

func Test_setupGc(t *testing.T) {
	type args struct {
		a *api.API
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupGc(tt.args.a)
		})
	}
}
