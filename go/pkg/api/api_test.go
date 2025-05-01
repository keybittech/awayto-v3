package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/encoding/protojson"
)

var integrationTest = &types.IntegrationTest{}

func init() {
	jsonBytes, err := os.ReadFile(filepath.Join("..", "..", "integrations", "integration_results.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = protojson.Unmarshal(jsonBytes, integrationTest)
	if err != nil {
		log.Fatal(err)
	}

	println("init api_test")
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func getTestApi(rl *RateLimiter) *API {
	a := &API{
		Server:   &http.Server{},
		Handlers: handlers.NewHandlers(),
	}
	a.InitMux(rl)
	return a
}

func getTestReq(b *testing.B, token, method, url string) *http.Request {
	testReq, err := http.NewRequest(method, url, nil)
	if err != nil {
		b.Fatal(err)
	}
	testReq.Header.Set("Authorization", "Bearer "+token)
	testReq.Header.Set("Accept", "application/json")
	testReq.Header.Set("X-TZ", "America/Los_Angeles")

	return testReq
}

func doBenchmarkRateLimit(b *testing.B, limit rate.Limit, burst int) {
	rl := NewRateLimit("api", limit, burst, time.Duration(5*time.Second))
	api := getTestApi(rl)
	user := integrationTest.TestUsers[0]
	req := getTestReq(b, user.TestToken, http.MethodGet, "/api/v1/profile/details")
	recorder := httptest.NewRecorder()
	reset(b)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()
		api.Server.Handler.ServeHTTP(recorder, req)
	}
}

func BenchmarkApiRateLimit(b *testing.B) {
	b.Run("1 req per second, 1 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 1, 1)
	})
	b.Run("10 req per second, 10 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 10, 10)
	})
	b.Run("100 req per second, 100 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 100, 100)
	})
	b.Run("1000 req per second, 1000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 1000, 1000)
	})
	b.Run("10000 req per second, 10000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 10000, 10000)
	})
}

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
			if got := tt.a.InitMux(NewRateLimit("test-init", 5, 5, time.Duration(time.Second))); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.InitMux() = %v, want %v", got, tt.want)
			}
		})
	}
}
