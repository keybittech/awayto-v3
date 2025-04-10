package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type MiddlewareTestSetup struct {
	// MockCtrl           *gomock.Controller
	// MockAi             *interfaces.MockIAi
	// MockDatabase       *interfaces.MockIDatabase
	// MockDatabaseClient *interfaces.MockIDatabaseClient
	// MockDatabaseTx     *interfaces.MockIDatabaseTx
	// MockDatabaseRows   *interfaces.MockIRows
	// MockDatabaseRow    *interfaces.MockIRow
	MockRedis       *interfaces.MockIRedis
	MockRedisClient *interfaces.MockIRedisClient
	// MockKeycloak       *interfaces.MockIKeycloak
	// MockSocket         *interfaces.MockISocket
	// UserSession        *types.UserSession
	API *API
}

type MiddlewareTestCase struct {
	name             string
	opts             *util.HandlerOptions
	method           string
	url              string
	session          *types.UserSession
	setupMock        func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string)
	expectedStatus   string
	expectedCode     int
	expectedBody     string
	verifyAfterwards func(api *MiddlewareTestSetup, ctx context.Context, cacheKey string)
}

func setupTestAPI(t *testing.T) *MiddlewareTestSetup {
	mockCtrl := gomock.NewController(t)
	mockRedis := interfaces.NewMockIRedis(mockCtrl)
	mockRedisClient := interfaces.NewMockIRedisClient(mockCtrl)
	return &MiddlewareTestSetup{
		MockRedis:       mockRedis,
		MockRedisClient: mockRedisClient,
		API: &API{
			Handlers: &handlers.Handlers{
				Redis: mockRedis,
			},
		},
	}
}

func middlewareHelper(middleware func(SessionHandler) SessionHandler, req *http.Request, session *types.UserSession, setup func(), verify func(*httptest.ResponseRecorder)) {
	// Setup
	if setup != nil {
		setup()
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Terminal handler that just writes a successful response
	terminalHandler := func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	// Apply middleware
	handlerWithMiddleware := middleware(terminalHandler)

	// Execute
	handlerWithMiddleware(w, req, session)

	// Verify
	verify(w)
}

func TestAPI_LimitMiddleware(t *testing.T) {
	type args struct {
		limit rate.Limit
		burst int
	}
	tests := []struct {
		name string
		a    *API
		args args
		want func(next http.HandlerFunc) http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.a.LimitMiddleware(tt.args.limit, tt.args.burst); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("API.LimitMiddleware(%v, %v) = %v, want %v", tt.args.limit, tt.args.burst, got, tt.want)
			// }
		})
	}
}

func TestAPI_ValidateTokenMiddleware(t *testing.T) {
	type args struct {
		limit rate.Limit
		burst int
	}
	tests := []struct {
		name string
		a    *API
		args args
		want func(next SessionHandler) http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.a.ValidateTokenMiddleware(tt.args.limit, tt.args.burst); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("API.ValidateTokenMiddleware(%v, %v) = %v, want %v", tt.args.limit, tt.args.burst, got, tt.want)
			// }
		})
	}
}

func TestAPI_GroupInfoMiddleware(t *testing.T) {
	type args struct {
		next SessionHandler
	}
	tests := []struct {
		name string
		a    *API
		args args
		want SessionHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.GroupInfoMiddleware(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.GroupInfoMiddleware(%v) = %v, want %v", tt.args.next, got, tt.want)
			}
		})
	}
}

func TestAPI_SiteRoleCheckMiddleware(t *testing.T) {
	type args struct {
		opts *util.HandlerOptions
	}
	tests := []struct {
		name string
		a    *API
		args args
		want func(SessionHandler) SessionHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.a.SiteRoleCheckMiddleware(tt.args.opts); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("API.SiteRoleCheckMiddleware(%v) = %v, want %v", tt.args.opts, got, tt.want)
			// }
		})
	}
}

func TestCacheWriter_Write(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		cw      *CacheWriter
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cw.Write(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheWriter.Write(%v) error = %v, wantErr %v", tt.args.data, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CacheWriter.Write(%v) = %v, want %v", tt.args.data, got, tt.want)
			}
		})
	}
}

func TestAPI_CacheMiddleware(t *testing.T) {
	tests := []MiddlewareTestCase{
		{
			name: "Cache miss - first GET request",
			opts: &util.HandlerOptions{
				CacheType: types.CacheType_DEFAULT,
			},
			method:  http.MethodGet,
			url:     "/api/group/users",
			session: &types.UserSession{UserSub: "test-user"},
			setupMock: func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string) {
				// Expect cache check (miss)
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					Get(gomock.Any(), cacheKey).
					DoAndReturn(func(_ context.Context, key string) *redis.StringCmd {
						return redis.NewStringResult("", redis.Nil)
					}).
					Times(1)

				// Expect cache set after handler executes
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					SetEx(gomock.Any(), cacheKey, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, key string, value interface{}, _ time.Duration) *redis.StatusCmd {
						return redis.NewStatusResult("OK", nil)
					}).
					Times(1)
			},
			expectedStatus: "MISS",
			expectedCode:   http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Cache hit - subsequent GET request",
			opts: &util.HandlerOptions{
				CacheType: types.CacheType_DEFAULT,
			},
			method:  http.MethodGet,
			url:     "/api/group/users",
			session: &types.UserSession{UserSub: "test-user"},
			setupMock: func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string) {
				// Create cached data
				cacheMeta := CacheMeta{
					Data:    []byte("cached response"),
					LastMod: time.Now(),
				}
				cacheData, _ := json.Marshal(cacheMeta)

				// Expect cache check (hit)
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					Get(gomock.Any(), cacheKey).
					Return(redis.NewStringResult(string(cacheData), nil)).
					Times(1)
			},
			expectedStatus: "HIT",
			expectedCode:   http.StatusOK,
			expectedBody:   "cached response",
		},
		{
			name: "Non-GET request - should bypass cache and delete",
			opts: &util.HandlerOptions{
				CacheType: types.CacheType_DEFAULT,
			},
			method:  http.MethodPost,
			url:     "/api/group/users",
			session: &types.UserSession{UserSub: "test-user"},
			setupMock: func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string) {
				// Expect cache deletion after handler executes
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					Del(gomock.Any(), cacheKey).
					Return(redis.NewIntResult(1, nil)).
					Times(1)
			},
			expectedCode: http.StatusOK,
			expectedBody: "success",
		},
		{
			name: "Not modified response",
			opts: &util.HandlerOptions{
				CacheType: types.CacheType_DEFAULT,
			},
			method:  http.MethodGet,
			url:     "/api/group/users",
			session: &types.UserSession{UserSub: "test-user"},
			setupMock: func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string) {
				// Create cached data with last modified time that matches If-Modified-Since
				lastMod := time.Now().Add(-time.Hour).Truncate(time.Second)
				cacheMeta := CacheMeta{
					Data:    []byte("cached response"),
					LastMod: lastMod,
				}
				cacheData, _ := json.Marshal(cacheMeta)

				// Expect cache check (hit)
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					Get(gomock.Any(), cacheKey).
					Return(redis.NewStringResult(string(cacheData), nil)).
					Times(1)
			},
			expectedStatus: "UNMODIFIED",
			expectedCode:   http.StatusNotModified,
			expectedBody:   "",
		},
		{
			name: "Cache store mode",
			opts: &util.HandlerOptions{
				CacheType: types.CacheType_STORE,
			},
			method:  http.MethodGet,
			url:     "/api/group/users",
			session: &types.UserSession{UserSub: "test-user"},
			setupMock: func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string) {
				// Expect cache check (miss)
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					Get(gomock.Any(), cacheKey).
					Return(redis.NewStringResult("", redis.Nil)).
					Times(1)

				// Expect permanent cache set after handler executes (no expiration)
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, key string, value interface{}, _ time.Duration) *redis.StatusCmd {
						return redis.NewStatusResult("OK", nil)
					}).
					Times(1)
			},
			expectedStatus: "MISS",
			expectedCode:   http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Cache skip mode",
			opts: &util.HandlerOptions{
				CacheType: types.CacheType_SKIP,
			},
			method:  http.MethodGet,
			url:     "/api/group/users",
			session: &types.UserSession{UserSub: "test-user"},
			setupMock: func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string) {
				// Expect cache set after handler executes (despite SKIP mode)
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					SetEx(gomock.Any(), cacheKey, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, key string, value interface{}, _ time.Duration) *redis.StatusCmd {
						return redis.NewStatusResult("OK", nil)
					}).
					Times(1)
			},
			expectedStatus: "MISS",
			expectedCode:   http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Custom cache duration",
			opts: &util.HandlerOptions{
				CacheType:     types.CacheType_DEFAULT,
				CacheDuration: 60, // 60 seconds
			},
			method:  http.MethodGet,
			url:     "/api/group/users",
			session: &types.UserSession{UserSub: "test-user"},
			setupMock: func(mockSetup *MiddlewareTestSetup, ctx context.Context, cacheKey string) {
				// Expect cache check (miss)
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					Get(gomock.Any(), cacheKey).
					Return(redis.NewStringResult("", redis.Nil)).
					Times(1)

				// Expect cache set with 60s duration
				mockSetup.MockRedis.EXPECT().
					Client().
					Return(mockSetup.MockRedisClient).
					Times(1)
				mockSetup.MockRedisClient.EXPECT().
					SetEx(gomock.Any(), cacheKey, gomock.Any(), gomock.Eq(time.Duration(60)*time.Second)).
					DoAndReturn(func(_ context.Context, key string, value interface{}, _ time.Duration) *redis.StatusCmd {
						return redis.NewStatusResult("OK", nil)
					}).
					Times(1)
			},
			expectedStatus: "MISS",
			expectedCode:   http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test API with mock controller
			mockSetup := setupTestAPI(t)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)
			os.Setenv("API_PATH", "/api")

			// If testing If-Modified-Since, add the header
			if tt.name == "Not modified response" {
				ifModifiedSince := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)
				req.Header.Set("If-Modified-Since", ifModifiedSince)
			}

			// Calculate cache key
			cacheKey := tt.session.UserSub + strings.TrimLeft(req.URL.String(), os.Getenv("API_PATH"))

			// Call the test helper
			middlewareHelper(
				mockSetup.API.CacheMiddleware(tt.opts),
				req,
				tt.session,
				func() {
					// Setup mock expectations
					if tt.setupMock != nil {
						tt.setupMock(mockSetup, req.Context(), cacheKey)
					}
				},
				func(w *httptest.ResponseRecorder) {
					// Verify response status
					if tt.expectedStatus != "" && w.Header().Get("X-Cache-Status") != tt.expectedStatus {
						t.Errorf("Expected cache status %s, got %s", tt.expectedStatus, w.Header().Get("X-Cache-Status"))
					}

					// Check response code
					if w.Code != tt.expectedCode {
						t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
					}

					// Check response body
					if string(w.Body.Bytes()) != tt.expectedBody {
						t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
					}

					// Run any additional verification
					if tt.verifyAfterwards != nil {
						tt.verifyAfterwards(mockSetup, req.Context(), cacheKey)
					}
				},
			)
		})
	}
}
