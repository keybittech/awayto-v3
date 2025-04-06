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

	return httptest.NewServer(handler)
}

func TestGet(t *testing.T) {
	headers := http.Header{}
	headers.Add("test-header-1", "test 1")
	tests := []struct {
		name    string
		url     string
		headers http.Header
		want    []byte
		wantErr bool
	}{
		{name: "valid URL", url: "/test", headers: nil, want: []byte("OK"), wantErr: false},
		{name: "valid URL w headers", url: "/test", headers: headers, want: []byte("OK"), wantErr: false},
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

func BenchmarkGet(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = Get("/test", nil)
	}
}

func BenchmarkGetHeaders(b *testing.B) {
	headers := http.Header{}
	headers.Add("test-header-1", "test 1")
	headers.Add("test-header-2", "test 2")
	headers.Add("test-header-3", "test 3")
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = Get("/test", headers)
	}
}

func BenchmarkGetError(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = Get("/error", nil)
	}
}

func TestGetWithParams(t *testing.T) {
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

func BenchmarkGetWithParams(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = GetWithParams("/test", nil, url.Values{"key": {"value"}})
	}
}

func BenchmarkGetWithParamsError(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = GetWithParams("/error", nil, url.Values{"key": {"value"}})
	}
}

func TestMutate(t *testing.T) {
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

func BenchmarkMutate(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = Mutate("POST", "/test", http.Header{"Content-Type": {"application/json"}}, []byte(`{"key":"value"}`))
	}
}

func BenchmarkMutateError(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = Mutate("POST", "/error", http.Header{"Content-Type": {"application/json"}}, []byte(`{"key":`))
	}
}

func TestPostFormData(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		headers http.Header
		data    url.Values
		want    []byte
		wantErr bool
	}{
		{name: "valid POST form data", url: "/test", headers: nil, data: url.Values{"key": {"value"}}, want: []byte("OK"), wantErr: false},
		{name: "invalid URL with POST form data", url: "/error", headers: nil, data: url.Values{"key": {"value"}}, want: nil, wantErr: true},
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

func BenchmarkPostFormData(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_, _ = PostFormData("/test", http.Header{"Content-Type": {"application/json"}}, url.Values{"key": {"value"}})
	}
}

func Test_successStatus(t *testing.T) {
	type args struct {
		statusCode int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "success code", args: args{200}, want: true},
		{name: "bad code 1", args: args{199}, want: false},
		{name: "bad code 2", args: args{300}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := successStatus(tt.args.statusCode); got != tt.want {
				t.Errorf("successStatus(%v) = %v, want %v", tt.args.statusCode, got, tt.want)
			}
		})
	}
}

func Benchmark_successStatus(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_ = successStatus(200)
	}
}

func Benchmark_successStatusNegative(b *testing.B) {
	server := mockServer()
	defer server.Close()

	reset(b)
	for i := 0; i < b.N; i++ {
		_ = successStatus(199)
	}
}
