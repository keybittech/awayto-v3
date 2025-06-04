package testutil

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
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

func handlerFollowRedirects(handler *http.ServeMux, w *httptest.ResponseRecorder, req *http.Request, cookies []*http.Cookie) (*httptest.ResponseRecorder, []*http.Cookie, error) {
	currentW := w
	currentReq := req
	for currentW.Code >= 300 && currentW.Code < 400 {
		location := currentW.Header().Get("Location")
		if location == "" {
			break
		}

		// Parse the redirect URL relative to the current request URL
		redirectURL, err := currentReq.URL.Parse(location)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse redirect URL: %v", err)
		}

		// Create new request with the resolved URL
		currentReq = GetTestReq("GET", redirectURL.String(), nil)
		for _, cookie := range cookies {
			currentReq.AddCookie(cookie)
		}
		currentW = httptest.NewRecorder()
		handler.ServeHTTP(currentW, currentReq)
		if cookies != nil {
			cookies = append(cookies, currentW.Result().Cookies()...)
		}
	}
	return currentW, cookies, nil
}

func (tus *TestUsersStruct) Login(handler ...*http.ServeMux) ([]*http.Cookie, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	sessionCookies := make([]*http.Cookie, 0)
	var resp []byte
	if handler != nil {
		h := handler[0]
		w := httptest.NewRecorder()
		req := GetTestReq("GET", util.E_APP_HOST_URL+"/auth/login?tz=America/Los_Angeles", nil)
		h.ServeHTTP(w, req)

		w, sessionCookies, err = handlerFollowRedirects(h, w, req, sessionCookies)
		if err != nil {
			return nil, fmt.Errorf("error following redirects, err: %v", err)
		}

		if w.Code != 200 {
			return nil, fmt.Errorf("GET /auth/login returned status %d", w.Code)
		}

		resp = w.Body.Bytes()

	} else {

		req := GetTestReq("GET", util.E_APP_HOST_URL+"/auth/login?tz=America/Los_Angeles", nil)
		resp, err = doAndRead(client, req)
		if err != nil {
			return nil, fmt.Errorf("failed to call /auth/login: %v", err)
		}
	}

	actionRegex := regexp.MustCompile(`<form[^>]*action="([^"]*)"`)
	actionMatches := actionRegex.FindStringSubmatch(string(resp))
	if len(actionMatches) < 2 {
		return nil, fmt.Errorf("could not find form action in Keycloak login page")
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

	if handler != nil {
		h := handler[0]
		w := httptest.NewRecorder()
		req := GetTestReq("POST", formActionURL, strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		for _, c := range sessionCookies {
			req.AddCookie(c)
		}
		h.ServeHTTP(w, req)

		w, sessionCookies, err = handlerFollowRedirects(h, w, req, sessionCookies)
		if err != nil {
			return nil, fmt.Errorf("error following redirects, err: %v", err)
		}

		if w.Code != 200 {
			return nil, fmt.Errorf("POST form action login returned status %d, path: %s", w.Code, req.URL.String())
		}
	} else {
		req := GetTestReq("POST", formActionURL, strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		for _, c := range sessionCookies {
			req.AddCookie(c)
		}
		_, err = doAndRead(client, req)
		if err != nil {
			return nil, fmt.Errorf("failed to post login form %s: %v", formActionURL, err)
		}
		appURL, _ := url.Parse(util.E_APP_HOST_URL)
		sessionCookies = append(sessionCookies, jar.Cookies(appURL)...)
	}

	var sessionId string
	for _, cookie := range sessionCookies {
		if cookie.Name == "session_id" {
			sessionId = cookie.Value
			break
		}
	}

	if sessionId == "" {
		return nil, fmt.Errorf("session_id cookie not found after login")
	}

	tus.SetCookieData(sessionCookies)

	return sessionCookies, nil
}

func (tus *TestUsersStruct) Logout(handler ...*http.ServeMux) error {
	req := GetTestReq("GET", util.E_APP_HOST_URL+"/auth/logout", nil)
	for _, c := range tus.CookieData {
		req.AddCookie(&c)
	}

	var resp []byte
	var err error
	if handler != nil {
		h := handler[0]
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		handlerFollowRedirects(h, w, req, nil)

		if w.Code != 200 {
			return fmt.Errorf("GET /auth/logout returned status %d", w.Code)
		}

		resp = w.Body.Bytes()
	} else {
		client := tus.getUserClient()
		resp, err = doAndRead(client, req)
		if err != nil {
			return fmt.Errorf("failed to call /auth/logout: %v", err)
		}
	}

	lastUpdatedMatch := regexp.MustCompile(`<p>Last updated:`).FindAll(resp, 1)
	if lastUpdatedMatch == nil {
		return fmt.Errorf("could not find last updated after logout")
	}

	tus.CookieData = nil

	return nil
}
