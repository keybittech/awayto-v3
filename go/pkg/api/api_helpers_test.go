package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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
		api = NewAPI(util.E_GO_HTTPS_PORT)
		go api.RedirectHTTP(util.E_GO_HTTP_PORT)
		go api.InitUnixServer(util.E_UNIX_PATH)
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

func getKeycloakToken(userId string) (string, *types.UserSession, error) {

	data := url.Values{}
	data.Set("client_id", util.E_KC_USER_CLIENT)
	data.Set("username", "1@"+userId)
	data.Set("password", "1")
	data.Set("grant_type", "password")

	var header http.Header
	header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := util.PostFormData("http://localhost:8080/auth/realms/"+util.E_KC_REALM+"/protocol/openid-connect/token", header, data)
	if err != nil {
		return "", nil, err
	}

	var tokens *types.OIDCToken
	if err := protojson.Unmarshal(resp, tokens); err != nil {
		return "", nil, err
	}

	// session, err := util.ValidateToken(publicKey, token, "America/Los_Angeles", "0.0.0.0")
	// if err != nil {
	// 	log.Fatalf("error validating auth token: %v", err)
	// }

	// if integrationTest.Group != nil {
	// 	groupId := integrationTest.Group.Id
	// 	session.GroupId = groupId
	// }

	return tokens.GetAccessToken(), nil, nil
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
	token, _, err := getKeycloakToken(testutil.IntegrationTest.TestUsers[userId].GetTestUserId())
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
		StartDate:          testutil.IntegrationTest.MasterSchedule.GetStartDate(),
		EndDate:            testutil.IntegrationTest.MasterSchedule.GetEndDate(),
		ScheduleTimeUnitId: testutil.IntegrationTest.MasterSchedule.GetScheduleTimeUnitId(),
		BracketTimeUnitId:  testutil.IntegrationTest.MasterSchedule.GetBracketTimeUnitId(),
		SlotTimeUnitId:     testutil.IntegrationTest.MasterSchedule.GetSlotTimeUnitId(),
		SlotDuration:       testutil.IntegrationTest.MasterSchedule.GetSlotDuration(),
	}).(*types.PostScheduleRequest)
}
