package util

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	type args struct {
		url     string
		headers http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.url, tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithParams(t *testing.T) {
	type args struct {
		url         string
		headers     http.Header
		queryParams url.Values
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWithParams(tt.args.url, tt.args.headers, tt.args.queryParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMutate(t *testing.T) {
	type args struct {
		method   string
		url      string
		headers  http.Header
		dataBody []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mutate(tt.args.method, tt.args.url, tt.args.headers, tt.args.dataBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostFormData(t *testing.T) {
	type args struct {
		url     string
		headers http.Header
		data    url.Values
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PostFormData(tt.args.url, tt.args.headers, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostFormData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostFormData() = %v, want %v", got, tt.want)
			}
		})
	}
}
