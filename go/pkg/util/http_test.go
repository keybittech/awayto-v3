package util

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func mockServer() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query()) > 0 {
			w.Write([]byte("response"))
		} else {
			w.Write([]byte("OK"))
		}
	})

	handler.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	handler.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Write([]byte("Post OK"))
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			w.Write([]byte("Form Submitted"))
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(handler)
}

func TestUtilGet(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		headers http.Header
		want    []byte
		wantErr bool
	}{
		{name: "valid URL", url: "/test", headers: nil, want: []byte("OK"), wantErr: false},
		{name: "invalid URL", url: "/error", headers: nil, want: nil, wantErr: true},
	}

	server := mockServer()
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + tt.url

			got, err := Get(url, tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilGetWithParams(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		headers http.Header
		params  url.Values
		want    []byte
		wantErr bool
	}{
		{name: "valid URL with query params", url: "/test", headers: nil, params: url.Values{"key": {"value"}}, want: []byte("OK"), wantErr: false},
		{name: "invalid URL with query params", url: "/error", headers: nil, params: url.Values{"key": {"value"}}, want: nil, wantErr: true},
	}

	server := mockServer()
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + tt.url

			got, err := GetWithParams(url, tt.headers, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("GetWithParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilMutate(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		url     string
		headers http.Header
		data    []byte
		want    []byte
		wantErr bool
	}{
		{name: "POST request with valid data", method: "POST", url: "/test", headers: http.Header{"Content-Type": {"application/json"}}, data: []byte(`{"key":"value"}`), want: []byte("OK"), wantErr: false},
		{name: "POST request with invalid data", method: "POST", url: "/error", headers: http.Header{"Content-Type": {"application/json"}}, data: []byte(`{"key":}`), want: nil, wantErr: true},
	}

	server := mockServer()
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + tt.url

			got, err := Mutate(tt.method, url, tt.headers, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Mutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtilPostFormData(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		headers http.Header
		data    url.Values
		want    []byte
		wantErr bool
	}{
		{name: "valid POST form data", url: "/test", headers: nil, data: url.Values{"key": {"value"}}, want: []byte("OK"), wantErr: false},
		{name: "invalid URL with POST form data", url: "/invalid", headers: nil, data: url.Values{"key": {"value"}}, want: nil, wantErr: true},
	}

	server := mockServer()
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + tt.url

			got, err := PostFormData(url, tt.headers, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostFormData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("PostFormData() = %v, want %v", got, tt.want)
			}
		})
	}
}
