package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	util.ParseEnv()
	testutil.LoadIntegrations()

	m.Run()
}

func BenchmarkApiCache(b *testing.B) {
	testUser := getTestUser(int32(0))
	method := http.MethodGet
	path := "/api/v1/quotes"
	contentType := "application/json"
	checkBytes := []byte{'{', 'T'} // T for Too many requests expected

	api_100_100 := getTestApi(100, 100)
	cookies, err := testUser.Login(api_100_100.Server.Handler.(*http.ServeMux))
	if err != nil {
		b.Fatalf("could not get cookies for login, err %v", err)
	}
	req, recorder := setupRouteRequest(cookies, method, path, contentType)
	if len(req.Cookies()) < 1 {
		b.Fatalf("no cookies, %+v", testUser.CookieData)
	}
	b.Run("100 req per second, 100 burst", func(bb *testing.B) {
		doApiBenchmark(bb, api_100_100, req, recorder, checkBytes) // Everything is successful up to here
	})
	testUser.Logout(api_100_100.Server.Handler.(*http.ServeMux))
	api_100_100.Server.Close()

	api_1000_1000 := getTestApi(1000, 1000)
	cookies, err = testUser.Login(api_1000_1000.Server.Handler.(*http.ServeMux))
	if err != nil {
		b.Fatalf("could not get cookies for login, err %v", err)
	}
	req, recorder = setupRouteRequest(cookies, method, path, contentType)
	if len(req.Cookies()) < 1 {
		b.Fatal("no cookies")
	}
	b.Run("1000 req per second, 1000 burst", func(bb *testing.B) {
		doApiBenchmark(bb, api_1000_1000, req, recorder, checkBytes)
	})
	testUser.Logout(api_1000_1000.Server.Handler.(*http.ServeMux))
	api_1000_1000.Server.Close()

	api_10000_10000 := getTestApi(10000, 10000)
	cookies, err = testUser.Login(api_10000_10000.Server.Handler.(*http.ServeMux))
	if err != nil {
		b.Fatalf("could not get cookies for login, err %v", err)
	}
	req, recorder = setupRouteRequest(cookies, method, path, contentType)
	if len(req.Cookies()) < 1 {
		b.Fatal("no cookies")
	}
	b.Run("10000 req per second, 10000 burst", func(bb *testing.B) {
		doApiBenchmark(bb, api_10000_10000, req, recorder, checkBytes)
	})
	testUser.Logout(api_10000_10000.Server.Handler.(*http.ServeMux))
	api_10000_10000.Server.Close()

	api_100000_100000 := getTestApi(100000, 100000)
	cookies, err = testUser.Login(api_100000_100000.Server.Handler.(*http.ServeMux))
	if err != nil {
		b.Fatalf("could not get cookies for login, err %v", err)
	}
	req, recorder = setupRouteRequest(cookies, method, path, contentType)
	if len(req.Cookies()) < 1 {
		b.Fatal("no cookies")
	}
	b.Run("100000 req per second, 100000 burst", func(bb *testing.B) {
		doApiBenchmark(bb, api_100000_100000, req, recorder, checkBytes)
	})
	testUser.Logout(api_100000_100000.Server.Handler.(*http.ServeMux))
	api_100000_100000.Server.Close()
}

func BenchmarkApiNoCache(b *testing.B) {
	testUser := getTestUser(int32(0))
	method := http.MethodGet
	contentType := "application/json"
	checkBytes := []byte{'{', '['}

	api_100000_100000 := getTestApi(100000, 100000)
	cookies, err := testUser.Login(api_100000_100000.Server.Handler.(*http.ServeMux))
	if err != nil {
		b.Fatalf("could not get cookies for login, err %v", err)
	}

	req, recorder := setupRouteRequest(cookies, method, "/api/v1/sock/ticket", contentType)
	if len(req.Cookies()) < 1 {
		b.Fatal("no cookies")
	}

	b.Run("sock_ticket", func(bb *testing.B) {
		doApiBenchmark(bb, api_100000_100000, req, recorder, checkBytes)
	})

	req, recorder = setupRouteRequest(cookies, method, "/api/v1/profile/details", contentType)
	if len(req.Cookies()) < 1 {
		b.Fatal("no cookies")
	}

	b.Run("profile_details", func(bb *testing.B) {
		doApiBenchmark(bb, api_100000_100000, req, recorder, checkBytes)
	})

	testUser.Logout(api_100000_100000.Server.Handler.(*http.ServeMux))
	api_100000_100000.Server.Close()
}

func BenchmarkApiProto(b *testing.B) {
	testUser := getTestUser(int32(0))
	method := http.MethodPost
	path := "/api/v1/schedules"
	checkBytes := []byte{'{', '['}
	scheduleName := " Master Creation Test"

	api_100000_100000 := getTestApi(100000, 100000)
	cookies, err := testUser.Login(api_100000_100000.Server.Handler.(*http.ServeMux))
	if err != nil {
		b.Fatalf("could not get cookies for login, err %v", err)
	}

	contentType := "application/json"
	req, recorder := setupRouteRequest(cookies, method, path, contentType)

	// pre-empt cache by making an initial request
	testSchedule := getApiProtoSchedule()
	testSchedule.Name = strconv.FormatInt(time.Now().UnixNano(), 10) + scheduleName
	setRouteRequestBody(req, testSchedule, contentType)
	api_100000_100000.Server.Handler.ServeHTTP(recorder, req)
	checkRouteRequest(recorder, checkBytes)

	i := time.Now().UnixNano()
	contentType = "application/x-protobuf"
	req, recorder = setupRouteRequest(cookies, method, path, contentType)
	b.Run("wire format", func(b *testing.B) {
		doApiBenchmarkWithBody(b, api_100000_100000, req, recorder, contentType, checkBytes, getApiProtoSchedule(), func(msg proto.Message) {
			msg.(*types.PostScheduleRequest).Name = strconv.FormatInt(i, 10) + scheduleName
			i++
		})
	})

	contentType = "application/json"
	req, recorder = setupRouteRequest(cookies, method, path, contentType)
	b.Run("json format", func(b *testing.B) {
		doApiBenchmarkWithBody(b, api_100000_100000, req, recorder, contentType, checkBytes, getApiProtoSchedule(), func(msg proto.Message) {
			msg.(*types.PostScheduleRequest).Name = strconv.FormatInt(i, 10) + scheduleName
			i++
		})
	})

	testUser.Logout(api_100000_100000.Server.Handler.(*http.ServeMux))
	api_100000_100000.Server.Close()
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

			testUser := getTestUser(tt.userId)
			api_100000_100000 := getTestApi(100000, 100000)
			cookies, err := testUser.Login(api_100000_100000.Server.Handler.(*http.ServeMux))
			if err != nil {
				t.Fatalf("could not login during test routes, err %v", err)
			}

			req, recorder := setupRouteRequest(cookies, tt.method, tt.path, tt.contentType)

			setRouteRequestBody(req, tt.body, tt.contentType)

			api_100000_100000.Handlers.Redis.RedisClient.FlushAll(req.Context())

			timeout := API_READ_TIMEOUT + API_WRITE_TIMEOUT
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer cancel()
			req = req.WithContext(ctx)

			done := make(chan struct{})

			startTime := time.Now()
			go func() {
				api_100000_100000.Server.Handler.ServeHTTP(recorder, req)
				close(done)
			}()

			select {
			case <-done:
				elapsed := time.Since(startTime)
				t.Logf("%s finished after %s", routeStr, elapsed)
				if recorder.Code != tt.expectedStatus {
					t.Errorf("%s did not have expected response status %d, got status: %d body: %s", routeStr, tt.expectedStatus, recorder.Code, recorder.Body)
				}
				t.Logf("completed %s, with body: %s", routeStr, recorder.Body.String())
			case <-time.After(timeout):
				t.Errorf("%s did not complete before r/w timeout %fs", routeStr, timeout.Seconds())
			}

			testUser.Logout(api_100000_100000.Server.Handler.(*http.ServeMux))
			api_100000_100000.Server.Close()
		})
	}
}
