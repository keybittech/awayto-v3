package api

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
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

func getTestApi(limit rate.Limit, burst int) *API {
	api := NewAPI(util.E_GO_HTTPS_PORT)
	go api.RedirectHTTP(util.E_GO_HTTP_PORT)
	go api.InitUnixServer(util.E_UNIX_AUTH_PATH)
	api.InitProtoHandlers()
	api.InitAuthProxy()
	api.InitSockServer()
	api.InitStatic()

	rateLimiter := NewRateLimit("api", limit, burst, time.Duration(5*time.Minute))
	limitMiddleware := api.LimitMiddleware(rateLimiter)(api.Server.Handler)
	wrappedHandler := api.AccessRequestMiddleware(limitMiddleware)

	finalHandler := http.NewServeMux()
	finalHandler.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		wrappedHandler.ServeHTTP(w, req)
	}))

	api.Server.Handler = finalHandler

	return api
}

func getTestUser(userId int32) *testutil.TestUsersStruct {
	testUser, ok := testutil.IntegrationTest.GetTestUsers()[userId]
	if !ok {
		log.Fatalf("could not find an integration user for id %d", userId)
	}
	return testUser
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

func setupRouteRequest(cookies []*http.Cookie, method, path, contentType string) (*http.Request, *httptest.ResponseRecorder) {
	req := testutil.GetTestReq(method, path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	req.Header.Set("Accept", contentType)
	req.Header.Set("Content-Type", contentType)
	return req, httptest.NewRecorder()
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
		panic(fmt.Sprintf("Response body (status %d) did not start with '{'. Got: %s", recorder.Code, recorder.Body.String()))
	}
}

func doApiBenchmark(b *testing.B, api *API, req *http.Request, recorder *httptest.ResponseRecorder, checkBytes []byte) {
	reset(b)
	for b.Loop() {
		recorder.Body.Reset()
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
