package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/api"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var publicKey *rsa.PublicKey

func init() {

	kc := &clients.KeycloakClient{
		Server: os.Getenv("KC_INTERNAL"),
		Realm:  os.Getenv("KC_REALM"),
	}

	oidcToken, err := kc.DirectGrantAuthentication()
	if err != nil {
		log.Fatal(err)
	}

	kc.Token = oidcToken

	pk, err := kc.FetchPublicKey()
	if err != nil {
		log.Fatal(err)
	}

	publicKey = pk
}

func failCheck(t *testing.T) {
	if t.Failed() {
		t.Logf("Integration Test: %+v", integrationTest)
	}
}

func apiRequest(token, method, path string, body []byte, queryParams map[string]string, responseObj proto.Message) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	reqURL := "https://localhost:7443" + path
	if queryParams != nil && len(queryParams) > 0 {
		values := url.Values{}
		for k, v := range queryParams {
			values.Add(k, v)
		}
		reqURL = reqURL + "?" + values.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if body != nil && len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Add("X-TZ", "America/Los_Angeles")

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if len(respBody) > 0 && responseObj != nil {
		if err := protojson.Unmarshal(respBody, responseObj); err != nil {
			return fmt.Errorf("error unmarshaling into responseObj: %w", err)
		}
	}

	return nil
}

func getKeycloakToken(userId int) (string, *types.UserSession, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("KC_CLIENT"))
	data.Set("username", "1@"+strconv.Itoa(userId+1))
	data.Set("password", "1")
	data.Set("grant_type", "password")

	req, err := http.NewRequest(
		"POST",
		"https://localhost:7443/auth/realms/"+os.Getenv("KC_REALM")+"/protocol/openid-connect/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", nil, err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", nil, fmt.Errorf("access token not found in response")
	}

	session, err := api.ValidateToken(publicKey, token, "America/Los_Angeles", "0.0.0.0")
	if err != nil {
		log.Fatalf("error validating auth token: %v", err)
	}

	return token, session, nil
}

func registerKeycloakUserViaForm(userId int, code ...string) (bool, error) {
	// Setup client with cookie jar and TLS skip
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// PKCE and state setup
	clientID := "devel-client"
	redirectURI := url.QueryEscape("https://localhost:7443/app")
	state := generateRandomString(36)
	nonce := generateRandomString(36)
	codeChallenge := generateCodeChallenge()

	// Load registration page
	registrationURL := fmt.Sprintf(
		"https://localhost:7443/auth/realms/"+os.Getenv("KC_REALM")+"/protocol/openid-connect/registrations?"+
			"client_id=%s&redirect_uri=%s&state=%s&response_mode=fragment&response_type=code&"+
			"scope=openid&nonce=%s&code_challenge=%s&code_challenge_method=S256",
		clientID, redirectURI, state, nonce, codeChallenge,
	)

	resp, err := client.Get(registrationURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	formAction := extractFormAction(body)
	if formAction == "" {
		return false, fmt.Errorf("failed to extract form action")
	}

	// Prepare and submit form data
	formData := url.Values{}

	if len(code) > 0 {
		formData.Set("groupCode", code[0])
	}

	formData.Set("email", "1@"+strconv.Itoa(userId+1))
	formData.Set("firstName", "first-name")
	formData.Set("lastName", "last-name")
	formData.Set("password", "1")
	formData.Set("password-confirm", "1")

	req, err := http.NewRequest("POST", formAction, strings.NewReader(formData.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	submitResp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer submitResp.Body.Close()

	// Registration success check

	return submitResp.StatusCode == http.StatusOK, nil
}

func extractFormAction(html []byte) string {
	startTag := `<form id="kc-register-form" class="form-horizontal" action="`
	startIdx := strings.Index(string(html), startTag)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startTag)
	endIdx := strings.Index(string(html[startIdx:]), `"`)
	if endIdx == -1 {
		return ""
	}
	return string(html[startIdx : startIdx+endIdx])
}

func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		panic(err)
	}
	n := nBig.Int64()
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[n]
	}
	return string(b)
}

func generateCodeChallenge() string {
	verifier := generateRandomString(43)
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func getSocketTicket(userId int) (*types.UserSession, string, string, string) {
	// Get authentication token first
	token, session, err := getKeycloakToken(userId)
	if err != nil {
		log.Fatalf("error getting auth token: %v", err)
	}

	req, err := http.NewRequest("GET", "https://localhost:7443/api/v1/sock/ticket", nil)
	if err != nil {
		log.Fatalf("could not make ticket request %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed ticket transport %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed ticket body read %v", err)
	}

	// Debug the response if it's not JSON
	if len(body) > 0 && body[0] != '{' {
		log.Fatalf("unexpected ticket response: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("ticket JSON parse error: %v, body: %s", err, string(body))
	}

	ticket, ok := result["ticket"].(string)
	if !ok {
		log.Fatal("ticket not found in response")
	}

	ticketParts := strings.Split(ticket, ":")
	_, connId := ticketParts[0], ticketParts[1]

	return session, token, ticket, connId
}

// func getUserSocketSession(userId int) (*types.SocketResponseParams, *types.UserSession, string, string) {
// 	userSession, _, ticket, connId := getSocketTicket(userId)
//
// 	subscriberSocketResponse, err := mainApi.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
// 		UserSub: userSession.UserSub,
// 		Ticket:  ticket,
// 	})
//
// 	if err = clients.ChannelError(err, subscriberSocketResponse.Error); err != nil {
// 		log.Fatal("failed to get subscriber socket command in setup", err.Error())
// 	}
//
// 	return subscriberSocketResponse.SocketResponseParams, userSession, ticket, connId
// }

func getClientSocketConnection(ticket string) net.Conn {
	u, err := url.Parse("wss://localhost:7443/sock?ticket=" + ticket)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	sockConn, err := tls.Dial("tcp", u.Host, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	// Generate WebSocket key
	keyBytes := make([]byte, 16)
	rand.Read(keyBytes)
	secWebSocketKey := base64.StdEncoding.EncodeToString(keyBytes)

	// Send HTTP Upgrade request
	fmt.Fprintf(sockConn, "GET %s HTTP/1.1\r\n", u.RequestURI())
	fmt.Fprintf(sockConn, "Host: %s\r\n", u.Host)
	fmt.Fprintf(sockConn, "Upgrade: websocket\r\n")
	fmt.Fprintf(sockConn, "Connection: Upgrade\r\n")
	fmt.Fprintf(sockConn, "Sec-WebSocket-Key: %s\r\n", secWebSocketKey)
	fmt.Fprintf(sockConn, "Sec-WebSocket-Version: 13\r\n")
	fmt.Fprintf(sockConn, "\r\n")

	reader := bufio.NewReader(sockConn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			return nil
		}
		if line == "\r\n" {
			break
		}
	}

	return sockConn
}

func getUser(userId int) (*types.UserSession, net.Conn, string, string, string) {

	userSession, token, ticket, connId := getSocketTicket(userId)

	connection := getClientSocketConnection(ticket)
	if connection == nil {
		log.Fatal("count not establish socket connection for user ", userId, ticket, connId)
	}
	connection.SetReadDeadline(time.Now().Add(120 * time.Second))

	return userSession, connection, token, ticket, connId
}

// func getSubscribers(numUsers int) (map[int]*types.Subscriber, map[string]net.Conn) {
// 	subscribers := make(map[int]*types.Subscriber, numUsers)
// 	connections := make(map[string]net.Conn, numUsers)
//
// 	for c := 0; c < numUsers; c++ {
// 		registerKeycloakUserViaForm(c)
//
// 		userSession, _, ticket, connId := getSocketTicket(c)
//
// 		connection := getClientSocketConnection(ticket)
// 		if connection == nil {
// 			println("count not establish socket connection for user ", strconv.Itoa(c), ticket, connId)
// 			return nil, nil
// 		}
// 		connection.SetReadDeadline(time.Now().Add(120 * time.Second))
// 		connections[userSession.UserSub] = connection
//
// 		subscriberRequest, err := mainApi.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
// 			UserSub: userSession.UserSub,
// 			Ticket:  ticket,
// 		}, connection)
// 		if err != nil {
// 			log.Fatalf("failed subscriber request %v", err)
// 		}
//
// 		subscribers[c] = &types.Subscriber{
// 			UserSub:      subscriberRequest.UserSub,
// 			GroupId:      subscriberRequest.GroupId,
// 			Roles:        subscriberRequest.Roles,
// 			ConnectionId: connId,
// 		}
//
// 		println("got subscribers ", len(subscribers), "connections", len(connections))
// 	}
//
// 	return subscribers, connections
// }

func getUserProfileDetails(token string) (*types.IUserProfile, error) {
	getUserProfileDetailsResponse := &types.GetUserProfileDetailsResponse{}
	err := apiRequest(token, http.MethodPatch, "/api/v1/profile/details", nil, nil, getUserProfileDetailsResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get user profile details error: %v", err))
	}
	if getUserProfileDetailsResponse.UserProfile.Id == "" {
		return nil, errors.New("get user profile details response has no id")
	}

	return getUserProfileDetailsResponse.UserProfile, nil
}

func getServiceById(token, serviceId string) (*types.IService, error) {
	getServiceByIdResponse := &types.GetServiceByIdResponse{}
	err := apiRequest(token, http.MethodGet, "/api/v1/services/"+serviceId, nil, nil, getServiceByIdResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get service by id error: %v", err))
	}
	if getServiceByIdResponse.Service.Id == "" {
		return nil, errors.New("get service by id response has no id")
	}

	return getServiceByIdResponse.Service, nil
}

func getScheduleById(token, scheduleId string) (*types.ISchedule, error) {
	getScheduleByIdResponse := &types.GetScheduleByIdResponse{}
	err := apiRequest(token, http.MethodGet, "/api/v1/schedules/"+scheduleId, nil, nil, getScheduleByIdResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get schedule by id error: %v", err))
	}
	if getScheduleByIdResponse.Schedule.Id == "" {
		return nil, errors.New("get schedule by id response has no id")
	}

	return getScheduleByIdResponse.Schedule, nil
}

func getMasterScheduleById(token, groupScheduleId string) (*types.IGroupSchedule, error) {
	getMasterScheduleByIdResponse := &types.GetGroupScheduleMasterByIdResponse{}
	err := apiRequest(token, http.MethodGet, "/api/v1/group/schedules/master/"+groupScheduleId, nil, nil, getMasterScheduleByIdResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get master schedule by id error: %v", err))
	}
	if getMasterScheduleByIdResponse.GroupSchedule.ScheduleId == "" {
		return nil, errors.New("get master schedule by id response has no schedule id")
	}

	return getMasterScheduleByIdResponse.GroupSchedule, nil
}

func patchGroupUser(token, userSub, roleId string) error {
	patchGroupUserRequestBytes, err := protojson.Marshal(&types.PatchGroupUserRequest{
		UserSub: userSub,
		RoleId:  roleId,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error marshalling patch group user %s %s %v", userSub, roleId, err))
	}

	patchGroupUserResponse := &types.PatchGroupUserResponse{}
	err = apiRequest(token, http.MethodPatch, "/api/v1/group/users", patchGroupUserRequestBytes, nil, patchGroupUserResponse)
	if err != nil {
		return errors.New(fmt.Sprintf("error patch group user request, sub: %s error: %v", userSub, err))
	}
	if !patchGroupUserResponse.Success {
		return errors.New("attach user internal was unsuccessful")
	}

	return nil
}

func patchGroupAssignments(token, roleFullName, actionName string) error {
	actions := make([]*types.IAssignmentAction, 1)
	actions[0] = &types.IAssignmentAction{
		Name: actionName,
	}
	assignmentActions := make(map[string]*types.IAssignmentActions)
	assignmentActions[roleFullName] = &types.IAssignmentActions{
		Actions: actions,
	}

	patchGroupAssignmentsBytes, err := protojson.Marshal(&types.PatchGroupAssignmentsRequest{
		Assignments: assignmentActions,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error marshalling patch group assignments %v %v", err, assignmentActions))
	}
	patchGroupAssignmentsResponse := &types.PatchGroupAssignmentsResponse{}
	err = apiRequest(token, http.MethodPatch, "/api/v1/group/assignments", patchGroupAssignmentsBytes, nil, patchGroupAssignmentsResponse)
	if err != nil {
		return errors.New(fmt.Sprintf("error patch group assignments request: %v", err))
	}
	if !patchGroupAssignmentsResponse.Success {
		return errors.New(fmt.Sprintf("patch group assignments  was unsuccessful %v", patchGroupAssignmentsResponse))
	}

	return nil
}

func postSchedule(token string) (*types.ISchedule, error) {
	brackets := make(map[string]*types.IScheduleBracket, 1)
	services := make(map[string]*types.IService, 1)
	slots := make(map[string]*types.IScheduleBracketSlot, 2)

	bracketId := strconv.Itoa(int(time.Now().UnixMilli()))
	time.Sleep(time.Millisecond)

	serviceId := integrationTest.MasterService.Id
	services[serviceId] = integrationTest.MasterService

	slot1Id := strconv.Itoa(int(time.Now().UnixMilli()))
	time.Sleep(time.Millisecond)

	slots[slot1Id] = &types.IScheduleBracketSlot{
		Id:                slot1Id,
		StartTime:         "P2DT1H",
		ScheduleBracketId: bracketId,
	}

	slot2Id := strconv.Itoa(int(time.Now().UnixMilli()))
	time.Sleep(time.Millisecond)

	slots[slot2Id] = &types.IScheduleBracketSlot{
		Id:                slot2Id,
		StartTime:         "P3DT4H",
		ScheduleBracketId: bracketId,
	}

	brackets[bracketId] = &types.IScheduleBracket{
		Id:         bracketId,
		Automatic:  false,
		Duration:   15,
		Multiplier: 100,
		Services:   services,
		Slots:      slots,
	}

	userScheduleRequestBytes, err := protojson.Marshal(&types.PostScheduleRequest{
		Brackets:           brackets,
		GroupScheduleId:    integrationTest.MasterSchedule.Id,
		Name:               integrationTest.MasterSchedule.Name,
		StartTime:          integrationTest.MasterSchedule.StartTime,
		EndTime:            integrationTest.MasterSchedule.EndTime,
		ScheduleTimeUnitId: integrationTest.MasterSchedule.ScheduleTimeUnitId,
		BracketTimeUnitId:  integrationTest.MasterSchedule.BracketTimeUnitId,
		SlotTimeUnitId:     integrationTest.MasterSchedule.SlotTimeUnitId,
		SlotDuration:       integrationTest.MasterSchedule.SlotDuration,
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error marshalling user schedule request: %v", err))
	}

	userScheduleResponse := &types.PostScheduleResponse{}
	err = apiRequest(token, http.MethodPost, "/api/v1/schedules", userScheduleRequestBytes, nil, userScheduleResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error post schedule request error: %v", err))
	}

	return getScheduleById(token, userScheduleResponse.Id)
}

func postQuote(token, serviceTierId string, slot *types.IGroupScheduleDateSlots, serviceForm, tierForm *types.IProtoFormVersionSubmission) (*types.IQuote, error) {
	if serviceForm == nil {
		serviceForm = &types.IProtoFormVersionSubmission{}
	}
	if tierForm == nil {
		tierForm = &types.IProtoFormVersionSubmission{}
	}

	postQuoteRequest := &types.PostQuoteRequest{
		SlotDate:                     slot.StartDate,
		ScheduleBracketSlotId:        slot.ScheduleBracketSlotId,
		ServiceTierId:                serviceTierId,
		ServiceFormVersionSubmission: serviceForm,
		TierFormVersionSubmission:    tierForm,
	}

	postQuoteBytes, err := protojson.Marshal(postQuoteRequest)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error marshalling post quote request %v", err))
	}

	postQuoteResponse := &types.PostQuoteResponse{}
	err = apiRequest(token, http.MethodPost, "/api/v1/quotes", postQuoteBytes, nil, postQuoteResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error post quote request error: %v", err))
	}
	if postQuoteResponse.Quote.Id == "" {
		return nil, errors.New("no post quote id")
	}

	return postQuoteResponse.Quote, nil
}

func postBooking(token string, bookingRequests []*types.IBooking) ([]*types.IBooking, error) {
	postBookingRequest := &types.PostBookingRequest{
		Bookings: bookingRequests,
	}

	postBookingBytes, err := protojson.Marshal(postBookingRequest)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error marshalling post booking request %v", err))
	}

	postBookingResponse := &types.PostBookingResponse{}
	err = apiRequest(token, http.MethodPost, "/api/v1/bookings", postBookingBytes, nil, postBookingResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error post booking request error: %v", err))
	}
	if len(postBookingResponse.Bookings) == 0 {
		return nil, errors.New("no bookings were created")
	}

	return postBookingResponse.Bookings, nil
}
