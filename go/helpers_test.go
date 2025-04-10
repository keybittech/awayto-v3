package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/api"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func apiRequest(token, method, path string, body []byte, queryParams map[string]string, response proto.Message) error {
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

	if len(respBody) > 0 {
		if err := protojson.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("error unmarshaling response: %w", err)
		}
	}

	return nil
}

func getKeycloakToken(userId int) (string, error) {
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
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return token, nil
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
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
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
	token, err := getKeycloakToken(userId)
	if err != nil {
		log.Fatalf("error getting auth token: %v", err)
	}

	session, err := api.ValidateToken(token, "America/Los_Angeles", "0.0.0.0")
	if err != nil {
		log.Fatalf("error validating auth token: %v", err)
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

func getUserSocketSession(userId int) (*types.SocketResponseParams, *types.UserSession, string, string) {
	userSession, _, ticket, connId := getSocketTicket(userId)

	subscriberSocketResponse, err := mainApi.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
		UserSub: userSession.UserSub,
		Ticket:  ticket,
	})

	if err = clients.ChannelError(err, subscriberSocketResponse.Error); err != nil {
		log.Fatal("failed to get subscriber socket command in setup", err.Error())
	}

	return subscriberSocketResponse.SocketResponseParams, userSession, ticket, connId
}

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

func getSubscribers(numUsers int) (map[int]*types.Subscriber, map[string]net.Conn) {
	subscribers := make(map[int]*types.Subscriber, numUsers)
	connections := make(map[string]net.Conn, numUsers)

	for c := 0; c < numUsers; c++ {
		registerKeycloakUserViaForm(c)

		userSession, _, ticket, connId := getSocketTicket(c)

		connection := getClientSocketConnection(ticket)
		if connection == nil {
			println("count not establish socket connection for user ", strconv.Itoa(c), ticket, connId)
			return nil, nil
		}
		connection.SetReadDeadline(time.Now().Add(120 * time.Second))
		connections[userSession.UserSub] = connection

		subscriberRequest, err := mainApi.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
			UserSub: userSession.UserSub,
			Ticket:  ticket,
		}, connection)
		if err != nil {
			log.Fatalf("failed subscriber request %v", err)
		}

		subscribers[c] = &types.Subscriber{
			UserSub:      subscriberRequest.UserSub,
			GroupId:      subscriberRequest.GroupId,
			Roles:        subscriberRequest.Roles,
			ConnectionId: connId,
		}

		println("got subscribers ", len(subscribers), "connections", len(connections))
	}

	return subscribers, connections
}
