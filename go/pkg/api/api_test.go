package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
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

func getTestApi() *API {
	api := &API{
		Server:   &http.Server{},
		Handlers: handlers.NewHandlers(),
	}
	api.InitMux()
	return api
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

func BenchmarkApiProfileDetails(b *testing.B) {
	api := getTestApi()
	user := integrationTest.TestUsers[0]
	req := getTestReq(b, user.TestToken, http.MethodGet, "/api/v1/profile/details")
	recorder := httptest.NewRecorder()
	reset(b)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		recorder.Body.Reset()
		b.StartTimer()

		api.Server.Handler.ServeHTTP(recorder, req)

		b.StopTimer()
		if recorder.Code != http.StatusOK && recorder.Code != http.StatusNotModified {
			b.Fatalf("Handler returned unexpected code %d", recorder.Code)
		}
		b.StartTimer()
	}
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
			if got := tt.a.InitMux(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.InitMux() = %v, want %v", got, tt.want)
			}
		})
	}
}
