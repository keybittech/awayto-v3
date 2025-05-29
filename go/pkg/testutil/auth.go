package testutil

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (tus *TestUsersStruct) RegisterKeycloakUserViaForm(code ...string) error {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	redirectUri := util.E_APP_HOST_URL + "/app"
	codeVerifier := util.GenerateCodeVerifier()
	codeChallenge := util.GenerateCodeChallenge(codeVerifier)
	state := util.GenerateState()
	nonce := util.GenerateState()

	registrationURL := fmt.Sprintf(
		"%s/auth/realms/%s/protocol/openid-connect/registrations?"+
			"client_id=%s&redirect_uri=%s&state=%s&response_mode=fragment&response_type=code&"+
			"scope=openid&nonce=%s&code_challenge=%s&code_challenge_method=S256",
		util.E_APP_HOST_URL, util.E_KC_REALM, util.E_KC_USER_CLIENT, redirectUri, state, nonce, codeChallenge,
	)

	resp, err := client.Get(registrationURL)
	if err != nil {
		return util.ErrCheck(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return util.ErrCheck(err)
	}

	startTag := `<form id="kc-register-form" class="form-horizontal" action="`
	startIdx := strings.Index(string(body), startTag)
	if startIdx == -1 {
		return util.ErrCheck(errors.New("couldn't find register action start"))
	}
	startIdx += len(startTag)
	endIdx := strings.Index(string(body[startIdx:]), `"`)
	if endIdx == -1 {
		return util.ErrCheck(errors.New("couldn't find register action end"))
	}
	formAction := string(body[startIdx : startIdx+endIdx])

	// Prepare and submit form data
	formData := url.Values{}

	if len(code) > 0 {
		formData.Set("groupCode", code[0])
	}

	formData.Set("email", tus.GetTestEmail())
	formData.Set("firstName", "first-name")
	formData.Set("lastName", "last-name")
	formData.Set("password", tus.GetTestPass())
	formData.Set("password-confirm", tus.GetTestPass())

	req, err := http.NewRequest("POST", formAction, strings.NewReader(formData.Encode()))
	if err != nil {
		return util.ErrCheck(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	submitResp, err := client.Do(req)
	if err != nil {
		return util.ErrCheck(err)
	}
	defer submitResp.Body.Close()

	// Registration success check
	if submitResp.StatusCode != http.StatusOK {
		return util.ErrCheck(fmt.Errorf("registration failed with code %d", submitResp.StatusCode))
	}

	println(fmt.Sprintf("Registered user %s with pass %s", tus.GetTestEmail(), tus.GetTestPass()))

	return nil
}

func (tus *TestUsersStruct) Login() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// if len(via) > 0 {
			// 	// Stop if redirecting from /auth/callback to /app/
			// 	// This allows us to capture the state after /auth/callback has run (e.g., cookie set)
			// 	// and verify that /auth/callback intends to redirect to /app/.
			// 	// via[len(via)-1] is the request that received the redirect response.
			// 	if strings.Contains(via[len(via)-1].URL.Path, "/auth/callback") && strings.Contains(req.URL.Path, "/app/") {
			// 		return http.ErrUseLastResponse
			// 	}
			// }

			return nil
		},
	}

	req, err := http.NewRequest("GET", util.E_APP_HOST_URL+"/auth/login?tz=America/Los_Angeles", nil)
	if err != nil {
		return fmt.Errorf("failed to create request for /auth/login: %v", err)
	}

	resp, err := doAndRead(client, req)
	if err != nil {
		return fmt.Errorf("failed to call /auth/login: %v", err)
	}

	actionRegex := regexp.MustCompile(`<form[^>]*action="([^"]*)"`)
	actionMatches := actionRegex.FindStringSubmatch(string(resp))
	if len(actionMatches) < 2 {
		return fmt.Errorf("could not find form action in Keycloak login page")
	}
	formActionURL := actionMatches[1]
	formActionURL = html.UnescapeString(formActionURL)

	hiddenInputRegex := regexp.MustCompile(`<input[^>]*type="hidden"[^>]*name="([^"]*)"[^>]*value="([^"]*)"`)
	hiddenMatches := hiddenInputRegex.FindAllStringSubmatch(string(resp), -1)
	formData := url.Values{}
	for _, match := range hiddenMatches {
		if len(match) >= 3 {
			formData.Set(match[1], match[2])
		}
	}

	formData.Set("username", tus.GetTestEmail())
	formData.Set("password", tus.GetTestPass())
	formData.Set("credentialId", "")

	req, err = http.NewRequest("POST", formActionURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = doAndRead(client, req)
	if err != nil {
		return fmt.Errorf("failed to post login form %s: %v", formActionURL, err)
	}

	appURL, _ := url.Parse(util.E_APP_HOST_URL)

	cookies := jar.Cookies(appURL)

	var sessionId string
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionId = cookie.Value
			break
		}
	}

	if sessionId == "" {
		return fmt.Errorf("session_id cookie not found after login")
	}

	tus.SetCookieData(cookies)

	return nil
}

func (tus *TestUsersStruct) Logout() error {
	req, err := http.NewRequest("GET", util.E_APP_HOST_URL+"/auth/logout", nil)
	if err != nil {
		return fmt.Errorf("failed to create request for /auth/logout: %v", err)
	}

	client := tus.getUserClient()
	resp, err := doAndRead(client, req)
	if err != nil {
		return fmt.Errorf("failed to call /auth/logout: %v", err)
	}

	lastUpdatedMatch := regexp.MustCompile(`<p>Last updated:`).FindAll(resp, 1)
	if lastUpdatedMatch == nil {
		return fmt.Errorf("could not find last updated after logout")
	}

	return nil
}

// func getUser(t *testing.T, userId string) (*types.UserSession, net.Conn, string, string, string) {
// 	userSession, token, ticket, connId := getSocketTicket(userId)
//
// 	connection, err := getClientSocketConnection(ticket)
// 	if err != nil {
// 		t.Fatalf("count not establish socket connection for user %s, ticket %s, connId %s", userId, ticket, connId)
// 	}
// 	connection.SetDeadline(time.Now().Add(10 * time.Second))
//
// 	return userSession, connection, token, ticket, connId
// }

// func getKeycloakToken(userId string) (string, *types.UserSession, error) {
//
// 	data := url.Values{}
// 	data.Set("client_id", util.E_KC_USER_CLIENT)
// 	data.Set("username", "1@"+userId)
// 	data.Set("password", "1")
// 	data.Set("grant_type", "password")
//
// 	var header http.Header
// 	header.Add("Content-Type", "application/x-www-form-urlencoded")
//
// 	resp, err := util.PostFormData("http://localhost:8080/auth/realms/"+util.E_KC_REALM+"/protocol/openid-connect/token", header, data)
// 	if err != nil {
// 		return "", nil, err
// 	}
//
// 	var tokens *types.OIDCToken
// 	if err := protojson.Unmarshal(resp, tokens); err != nil {
// 		return "", nil, err
// 	}
//
// 	return tokens.GetAccessToken(), nil, nil
// }

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
//
// func getUserProfileDetails(token string) (*types.IUserProfile, error) {
// 	getUserProfileDetailsResponse := &types.GetUserProfileDetailsResponse{}
// 	err := apiRequest(token, http.MethodPatch, "/api/v1/profile/details", nil, nil, getUserProfileDetailsResponse)
// 	if err != nil {
// 		return nil, errors.New(fmt.Sprintf("error get user profile details error: %v", err))
// 	}
// 	if getUserProfileDetailsResponse.UserProfile.Id == "" {
// 		return nil, errors.New("get user profile details response has no id")
// 	}
//
// 	return getUserProfileDetailsResponse.UserProfile, nil
// }

// func apiBenchRequest(token, method, path string, body []byte, queryParams map[string]string) (*http.Client, *http.Request, error) {
// 	reqURL := "https://localhost:7443" + path
// 	if queryParams != nil && len(queryParams) > 0 {
// 		values := url.Values{}
// 		for k, v := range queryParams {
// 			values.Add(k, v)
// 		}
// 		reqURL = reqURL + "?" + values.Encode()
// 	}
//
// 	var reqBody io.Reader
// 	if body != nil {
// 		reqBody = bytes.NewBuffer(body)
// 	}
// 	req, err := http.NewRequest(method, reqURL, reqBody)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("error creating request: %w", err)
// 	}
//
// 	if body != nil && len(body) > 0 {
// 		req.Header.Set("Content-Type", "application/json")
// 	}
// 	req.Header.Set("Accept", "application/json")
// 	req.Header.Add("X-Tz", "America/Los_Angeles")
//
// 	req.Header.Add("Authorization", "Bearer "+token)
//
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: true,
// 			},
// 		},
// 	}
//
// 	return client, req, nil
// }
