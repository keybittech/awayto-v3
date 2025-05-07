package api

import (
	"context"
	"crypto/rsa"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	integrationTest = &types.IntegrationTest{}
	publicKey       *rsa.PublicKey
)

func init() {
	jsonBytes, err := os.ReadFile(filepath.Join(os.Getenv("PROJECT_DIR"), "go", "integrations", "integration_results.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = protojson.Unmarshal(jsonBytes, integrationTest)
	if err != nil {
		log.Fatal(err)
	}

	kc := clients.InitKeycloak()
	publicKey = kc.Client.PublicKey
	kc.Close()
}

func doBenchmarkRateLimit(b *testing.B, limit rate.Limit, burst int, path string) {
	rl := NewRateLimit("api", limit, burst, time.Duration(5*time.Second))
	api := getTestApi(rl)
	user := integrationTest.TestUsers[1]

	req := getTestReq(user.TestToken, http.MethodGet, path, nil)
	recorder := httptest.NewRecorder()

	reset(b)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()
		api.Server.Handler.ServeHTTP(recorder, req)
	}

	good := checkResponseFor(recorder.Body.Bytes(), []byte{'{', 'T'})
	if !good {
		b.Errorf("Response body (status %d) did not start with '{'. Got: %s", recorder.Code, string(recorder.Body.Bytes()))
	}
}

func BenchmarkApiCache(b *testing.B) {
	// b.Run("1 req per second, 1 burst", func(bb *testing.B) {
	// 	doBenchmarkRateLimit(bb, 1, 1, "/api/v1/quotes")
	// })
	// b.Run("10 req per second, 10 burst", func(bb *testing.B) {
	// 	doBenchmarkRateLimit(bb, 10, 10, "/api/v1/quotes")
	// })
	b.Run("100 req per second, 100 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 100, 100, "/api/v1/quotes")
	})
	b.Run("1000 req per second, 1000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 1000, 1000, "/api/v1/quotes")
	})
	b.Run("10000 req per second, 10000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 10000, 10000, "/api/v1/quotes")
	})
}

func BenchmarkApiNoCache(b *testing.B) {
	// b.Run("1 req per second, 1 burst", func(bb *testing.B) {
	// 	doBenchmarkRateLimit(bb, 1, 1, "/api/v1/sock/ticket")
	// })
	// b.Run("10 req per second, 10 burst", func(bb *testing.B) {
	// 	doBenchmarkRateLimit(bb, 10, 10, "/api/v1/sock/ticket")
	// })
	b.Run("100 req per second, 100 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 100, 100, "/api/v1/sock/ticket")
	})
	b.Run("1000 req per second, 1000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 1000, 1000, "/api/v1/sock/ticket")
	})
	b.Run("10000 req per second, 10000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 10000, 10000, "/api/v1/sock/ticket")
	})
}

func TestNewAPI(t *testing.T) {
	type args struct {
		httpsPort int
	}
	tests := []struct {
		name string
		args args
		want *API
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := NewAPI(tt.args.httpsPort); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewAPI(%v) = %v, want %v", tt.args.httpsPort, got, tt.want)
			// }
		})
	}
}

func TestRoutes(t *testing.T) {
	rl := NewRateLimit("api", 1000, 1000, time.Duration(5*time.Second))
	a := getTestApi(rl)

	tests := []struct {
		name, method, url string
		user              *types.TestUser
		api               *API
		body              io.Reader
	}{
		{"get roles", http.MethodGet, "/api/v1/roles", integrationTest.TestUsers[0], a, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _, err := getKeycloakToken(tt.user)
			if err != nil {
				t.Fatalf("bad token get %v", err)
			}
			req := getTestReq(token, tt.method, tt.url, tt.body)
			recorder := httptest.NewRecorder()
			a.Handlers.Redis.RedisClient.FlushAll(context.Background())
			start := time.Now()
			tt.api.Server.Handler.ServeHTTP(recorder, req)
			end := time.Since(start)
			good := checkResponseFor(recorder.Body.Bytes(), []byte{'{'})
			if !good {
				t.Errorf("Response body (status %d) did not start with '{'. Got: %s", recorder.Code, string(recorder.Body.Bytes()))
			} else {
				t.Logf("TestRoutes(%s %s) finished successfully after %s", tt.method, tt.url, end.String())
			}
		})
	}

}
