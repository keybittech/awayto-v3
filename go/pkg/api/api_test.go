package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	integrationTest = &types.IntegrationTest{}
)

func TestMain(m *testing.M) {
	util.ParseEnv()

	jsonBytes, err := os.ReadFile(filepath.Join(util.E_PROJECT_DIR, "go", "integrations", "integration_results.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = protojson.Unmarshal(jsonBytes, integrationTest)
	if err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func BenchmarkApiCache(b *testing.B) {
	userId := int32(0)
	method := http.MethodGet
	path := "/api/v1/quotes"
	contentType := "application/json"
	checkBytes := []byte{'{', 'T'} // T for Too many requests expected
	b.Run("100 req per second, 100 burst", func(bb *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 100, 100, method, path, contentType)
		doApiBenchmark(bb, api, req, recorder, checkBytes)
	})
	b.Run("1000 req per second, 1000 burst", func(bb *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 1000, 1000, method, path, contentType)
		doApiBenchmark(bb, api, req, recorder, checkBytes)
	})
	b.Run("10000 req per second, 10000 burst", func(bb *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 10000, 10000, method, path, contentType)
		doApiBenchmark(bb, api, req, recorder, checkBytes)
	})
	b.Run("100000 req per second, 100000 burst", func(bb *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 100000, 100000, method, path, contentType)
		doApiBenchmark(bb, api, req, recorder, checkBytes)
	})
}

func BenchmarkApiNoCache(b *testing.B) {
	userId := int32(0)
	method := http.MethodGet
	contentType := "application/json"
	checkBytes := []byte{'{', '['}
	b.Run("sock_ticket", func(bb *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 100000, 100000, method, "/api/v1/sock/ticket", contentType)
		doApiBenchmark(bb, api, req, recorder, checkBytes)
	})
	b.Run("profile_details", func(bb *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 100000, 100000, method, "/api/v1/profile/details", contentType)
		doApiBenchmark(bb, api, req, recorder, checkBytes)
	})
}

func BenchmarkApiProto(b *testing.B) {
	userId := int32(0)
	method := http.MethodPost
	path := "/api/v1/schedules"
	checkBytes := []byte{'{', '['}
	scheduleName := " Master Creation Test"

	// pre-empt cache
	contentType := "application/json"
	api, req, recorder := setupRouteRequest(userId, 100000, 100000, method, path, contentType)
	testSchedule := getApiProtoSchedule()
	testSchedule.Name = strconv.FormatInt(time.Now().UnixNano(), 10) + scheduleName
	setRouteRequestBody(req, testSchedule, contentType)
	api.Server.Handler.ServeHTTP(recorder, req)
	checkRouteRequest(recorder, checkBytes)

	i := time.Now().UnixNano()
	contentType = "application/x-protobuf"
	b.Run("wire format", func(b *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 100000, 100000, method, path, contentType)
		doApiBenchmarkWithBody(b, api, req, recorder, contentType, checkBytes, getApiProtoSchedule(), func(msg proto.Message) {
			msg.(*types.PostScheduleRequest).Name = strconv.FormatInt(i, 10) + scheduleName
			i++
		})
	})

	contentType = "application/json"
	b.Run("json format", func(b *testing.B) {
		api, req, recorder := setupRouteRequest(userId, 100000, 100000, method, path, contentType)
		doApiBenchmarkWithBody(b, api, req, recorder, contentType, checkBytes, getApiProtoSchedule(), func(msg proto.Message) {
			msg.(*types.PostScheduleRequest).Name = strconv.FormatInt(i, 10) + scheduleName
			i++
		})
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
		{"get roles basic is successful", http.MethodGet, "/api/v1/group/roles", "application/json", 0, nil, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routeStr := fmt.Sprintf("TestRoutes(%s %s %s)", tt.method, tt.path, tt.contentType)

			api, req, recorder := setupRouteRequest(tt.userId, 100000, 100000, tt.method, tt.path, tt.contentType)

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
