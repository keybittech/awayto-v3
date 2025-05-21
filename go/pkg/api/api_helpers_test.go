package api

import (
	"bytes"
	"crypto/tls"
	json "encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func init() {
	util.MakeLoggers()
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

var testApi *API

func getTestApi(limit rate.Limit, burst int) *API {
	var api *API
	if testApi != nil {
		api = testApi
	} else {
		httpsPort, err := strconv.Atoi(os.Getenv("GO_HTTPS_PORT"))
		if err != nil {
			log.Fatalf("error getting test api %v", err)
		}
		httpPort, err := strconv.Atoi(os.Getenv("GO_HTTP_PORT"))
		if err != nil {
			log.Fatalf("error getting test api %v", err)
		}
		unixPath := "/tmp/test.sock"

		api = NewAPI(httpsPort)

		go api.RedirectHTTP(httpPort)

		go api.InitUnixServer(unixPath)
	}

	api.Server.Handler = http.NewServeMux()
	api.InitProtoHandlers()
	api.InitAuthProxy()
	api.InitSockServer()
	api.InitStatic()
	rateLimiter := NewRateLimit("api", limit, burst, time.Duration(5*time.Minute))
	limitMiddleware := api.LimitMiddleware(rateLimiter)(api.Server.Handler)
	api.Server.Handler = api.AccessRequestMiddleware(limitMiddleware)

	testApi = api

	return api
}

// func getProtoTestReq(token, method, url string, bodyMessage proto.Message) *http.Request {
// 	var bodyReader io.Reader
// 	if bodyMessage != nil {
// 		marshaledBody, err := proto.Marshal(bodyMessage)
// 		if err != nil {
// 			log.Fatalf("Failed to marshal proto message: %v", err)
// 		}
// 		bodyReader = bytes.NewReader(marshaledBody)
// 	}
//
// 	testReq, err := http.NewRequest(method, url, bodyReader)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	testReq.RemoteAddr = "127.0.0.1:9999"
// 	testReq.Header.Set("Authorization", "Bearer "+token)
// 	testReq.Header.Set("Accept", "application/x-protobuf")
// 	testReq.Header.Set("X-Tz", "America/Los_Angeles")
// 	if bodyMessage != nil {
// 		testReq.Header.Set("Content-Type", "application/x-protobuf")
// 	}
// 	return testReq
// }

func getTestReq(token, method, url string, body io.Reader) *http.Request {
	testReq, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatal(err)
	}
	testReq.RemoteAddr = "127.0.0.1:9999"
	testReq.Header.Set("Authorization", "Bearer "+token)
	testReq.Header.Set("Accept", "application/json")
	testReq.Header.Set("X-Tz", "America/Los_Angeles")
	if body != nil {
		testReq.Header.Set("Content-Type", "application/json")
	}
	return testReq
}

func getKeycloakToken(user *types.TestUser) (string, *types.UserSession, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("KC_CLIENT"))
	data.Set("username", "1@"+user.TestUserId)
	data.Set("password", "1")
	data.Set("grant_type", "password")

	req, err := http.NewRequest(
		"POST",
		"http://localhost:8080/auth/realms/"+os.Getenv("KC_REALM")+"/protocol/openid-connect/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return "", nil, err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", nil, fmt.Errorf("access token not found in response")
	}

	session, err := ValidateToken(publicKey, token, "America/Los_Angeles", "0.0.0.0")
	if err != nil {
		log.Fatalf("error validating auth token: %v", err)
	}

	return token, session, nil
}

func checkResponseFor(buf []byte, items []byte) bool {
	if len(buf) == 0 {
		return false
	}

	if !bytes.Contains(items, buf[:1]) {
		return false
	}

	return true
}

func setupRouteRequest(userId int32, limit rate.Limit, burst int, method, path, contentType string) (*API, *http.Request, *httptest.ResponseRecorder) {
	token, _, err := getKeycloakToken(integrationTest.TestUsers[userId])
	if err != nil {
		panic(util.ErrCheck(err))
	}
	req := getTestReq(token, method, path, nil)
	req.RemoteAddr = "127.0.0.1:9999"
	req.Header.Set("X-Tz", "America/Los_Angeles")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", contentType)
	req.Header.Set("Content-Type", contentType)

	return getTestApi(limit, burst), req, httptest.NewRecorder()
}

func setRouteRequestBody(req *http.Request, body proto.Message, contentType string) {
	var bodyReader io.Reader
	switch contentType {
	case "application/x-protobuf":
		marshaledBody, _ := proto.Marshal(body)
		bodyReader = bytes.NewReader(marshaledBody)
	case "application/json":
		marshaledBody, _ := protojson.Marshal(body)
		bodyReader = bytes.NewReader(marshaledBody)
	default:
		marshaledBody, _ := protojson.Marshal(body)
		bodyReader = bytes.NewReader(marshaledBody)
	}
	req.Body = io.NopCloser(bodyReader)
}

func checkRouteRequest(recorder *httptest.ResponseRecorder, byteSet []byte) {
	good := checkResponseFor(recorder.Body.Bytes(), byteSet)
	if !good {
		panic("Response body (status %d) did not start with '{'. Got: %s" + strconv.Itoa(recorder.Code) + string(recorder.Body.Bytes()))
	}
}

func doApiBenchmark(b *testing.B, api *API, req *http.Request, recorder *httptest.ResponseRecorder, checkBytes []byte) {
	reset(b)
	for b.Loop() {
		b.StopTimer()
		recorder.Body.Reset()
		b.StartTimer()
		api.Server.Handler.ServeHTTP(recorder, req)
	}

	checkRouteRequest(recorder, checkBytes)
}

func doApiBenchmarkWithBody(b *testing.B, api *API, req *http.Request, recorder *httptest.ResponseRecorder, contentType string, checkBytes []byte, body proto.Message, update func(proto.Message)) {

	reset(b)
	for b.Loop() {
		b.StopTimer()
		update(body)
		setRouteRequestBody(req, body, contentType)
		recorder.Body.Reset()
		b.StartTimer()

		api.Server.Handler.ServeHTTP(recorder, req)
	}

	checkRouteRequest(recorder, checkBytes)
}

func getApiProtoSchedule() *types.PostScheduleRequest {
	return proto.Clone(&types.PostScheduleRequest{
		AsGroup:            true,
		StartDate:          integrationTest.MasterSchedule.GetStartDate(),
		EndDate:            integrationTest.MasterSchedule.GetEndDate(),
		ScheduleTimeUnitId: integrationTest.MasterSchedule.GetScheduleTimeUnitId(),
		BracketTimeUnitId:  integrationTest.MasterSchedule.GetBracketTimeUnitId(),
		SlotTimeUnitId:     integrationTest.MasterSchedule.GetSlotTimeUnitId(),
		SlotDuration:       integrationTest.MasterSchedule.GetSlotDuration(),
	}).(*types.PostScheduleRequest)
}
