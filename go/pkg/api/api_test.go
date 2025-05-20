package api

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

func BenchmarkApiCache(b *testing.B) {
	b.Run("100 req per second, 100 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 100, 100, "/api/v1/quotes")
	})
	b.Run("1000 req per second, 1000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 1000, 1000, "/api/v1/quotes")
	})
	b.Run("10000 req per second, 10000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 10000, 10000, "/api/v1/quotes")
	})
	b.Run("100000 req per second, 100000 burst", func(bb *testing.B) {
		doBenchmarkRateLimit(bb, 100000, 100000, "/api/v1/quotes")
	})
}

func BenchmarkApiNoCache(b *testing.B) {
	b.Run("sock_ticket", func(bb *testing.B) {
		doBenchmarkNoCache(bb, "/api/v1/sock/ticket")
	})
	b.Run("profile_details", func(bb *testing.B) {
		doBenchmarkNoCache(bb, "/api/v1/profile/details")
	})
}

func BenchmarkApiProtoWire(b *testing.B) {
	name := " Master Creation Test"
	i := time.Now().UnixNano()
	doProtoBenchmark(b, http.MethodPost, "/api/v1/schedules", "application/x-protobuf", getApiProtoSchedule(), func(msg proto.Message) {
		msg.(*types.PostScheduleRequest).Name = strconv.FormatInt(i, 10) + name
		i++
	})
}

func BenchmarkApiProtoJson(b *testing.B) {
	name := " Master Creation Test"
	i := time.Now().UnixNano()
	doProtoBenchmark(b, http.MethodPost, "/api/v1/schedules", "application/json", getApiProtoSchedule(), func(msg proto.Message) {
		msg.(*types.PostScheduleRequest).Name = strconv.FormatInt(i, 10) + name
		i++
	})
}

func TestNewAPI(t *testing.T) {
	tests := []struct {
		name        string
		httpsPort   int
		wantSuccess bool
		wantFunc    func(*testing.T, *API) bool
	}{
		{"can cache groups", 7443, true, func(t *testing.T, a *API) bool {
			if a.Cache == nil {
				t.Fatal("group cache is nil")
			}
			a.Cache.Groups.Store("test-group", &types.ConcurrentCachedGroup{})
			if a.Cache.Groups.Len() != 1 {
				return false
			}
			return true
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := NewAPI(tt.httpsPort)
			if api == nil {
				t.Errorf("NewAPI returned a nil object")
			}
			if got := tt.wantFunc(t, api); got != tt.wantSuccess {
				t.Errorf("wanted success %t, got: %t", tt.wantSuccess, got)
			}
		})
	}
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		name, method, path, contentType string
		userId                          int32
		body                            proto.Message
		expectedStatus                  int
	}{
		{"get roles basic is successful", http.MethodGet, "/api/v1/roles", "application/json", 0, nil, http.StatusOK},
		// {"get role by id is successful", http.MethodGet, "/api/v1/roles/123e4567-e89b-12d3-a456-426614174000", "application/json", 0, nil, http.StatusOK},
		// {"post role creates new role", http.MethodPost, "/api/v1/roles", "application/json", 0, &types.PostRoleRequest{Name: "TestRole"}, http.StatusOK},
		// {"post role with existing name returns existing role id", http.MethodPost, "/api/v1/roles", "application/json", 0, &types.PostRoleRequest{Name: "ExistingRole"}, http.StatusOK},
		// {"post role with empty name fails validation", http.MethodPost, "/api/v1/roles", "application/json", 0, &types.PostRoleRequest{Name: ""}, http.StatusBadRequest},
		// {"delete role is successful", http.MethodDelete, "/api/v1/roles/123e4567-e89b-12d3-a456-426614174000", "application/json", 0, nil, http.StatusOK},
		// {"delete multiple roles is successful", http.MethodDelete, "/api/v1/roles/123e4567-e89b-12d3-a456-426614174000,223e4567-e89b-12d3-a456-426614174001", "application/json", 0, nil, http.StatusOK},
		// {"delete non-existent role is still successful", http.MethodDelete, "/api/v1/roles/non-existent-id", "application/json", 0, nil, http.StatusOK},
		// {"get role by invalid id format returns error", http.MethodGet, "/api/v1/roles/invalid-uuid-format", "application/json", 0, nil, http.StatusBadRequest},
		// {"post role with malformed json fails", http.MethodPost, "/api/v1/roles", "application/json", 0, &types.PostRoleRequest{}, http.StatusBadRequest},
		// {"get roles accepts correct content type", http.MethodGet, "/api/v1/roles", "application/json", 0, nil, http.StatusOK},
		// {"post role accepts correct content type", http.MethodPost, "/api/v1/roles", "application/json", 0, &types.PostRoleRequest{Name: "ContentTypeTest"}, http.StatusOK},
		// {"get roles rejects incorrect content type", http.MethodGet, "/api/v1/roles", "text/plain", 0, nil, http.StatusUnsupportedMediaType},
		// {"post role rejects incorrect content type", http.MethodPost, "/api/v1/roles", "text/plain", 0, &types.PostRoleRequest{Name: "ContentTypeTest"}, http.StatusUnsupportedMediaType}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routeStr := fmt.Sprintf("TestRoutes(%s %s %s)", tt.method, tt.path, tt.contentType)

			api, req, recorder := setupRouteRequest(tt.userId, tt.method, tt.path, tt.contentType)

			setRouteRequestBody(req, tt.body, tt.contentType)

			api.Handlers.Redis.RedisClient.FlushAll(req.Context())

			timeout := API_READ_TIMEOUT + API_WRITE_TIMEOUT
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer cancel()
			req = req.WithContext(ctx)

			done := make(chan struct{})

			startTime := time.Now()
			go func() {
				api.Server.Handler.ServeHTTP(recorder, req)
				close(done)
			}()

			select {
			case <-done:
				elapsed := time.Since(startTime)
				t.Logf("%s finished after %s", routeStr, elapsed)
				if recorder.Code != tt.expectedStatus {
					t.Errorf("%s did not have expected response status %d, got status: %d body: %s", routeStr, tt.expectedStatus, recorder.Code, recorder.Body)
				}
			case <-time.After(timeout):
				t.Errorf("%s did not complete before r/w timeout %fs", routeStr, timeout.Seconds())
			}
		})
	}
}
