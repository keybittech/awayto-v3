package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func getTestApi(rateLimiter *RateLimiter) *API {
	httpsPort, err := strconv.Atoi(os.Getenv("GO_HTTPS_PORT"))
	if err != nil {
		log.Fatalf("error getting test api %v", err)
	}
	a := NewAPI(httpsPort)
	a.InitProtoHandlers()
	a.Server.Handler = a.LimitMiddleware(rateLimiter)(a.Server.Handler)
	return a
}

func getTestReq(b *testing.B, token, method, url string, body io.Reader) *http.Request {
	testReq, err := http.NewRequest(method, url, body)
	if err != nil {
		b.Fatal(err)
	}
	testReq.RemoteAddr = "127.0.0.1:9999"
	testReq.Header.Set("Authorization", "Bearer "+token)
	testReq.Header.Set("Accept", "application/json")
	testReq.Header.Set("X-TZ", "America/Los_Angeles")
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

	var result map[string]interface{}
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
