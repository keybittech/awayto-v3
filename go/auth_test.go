package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/exp/rand"
)

func getKeycloakToken(cid int) (string, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("KC_CLIENT"))
	data.Set("username", "1@"+strconv.Itoa(cid+1))
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

func registerKeycloakUserViaForm(email, firstName, lastName, password string) (bool, error) {
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
	formData.Set("email", email)
	formData.Set("firstName", firstName)
	formData.Set("lastName", lastName)
	formData.Set("password", password)
	formData.Set("password-confirm", password)

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

func TestKeycloakRegistrationViaForm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping registration test")
	}
	numConnections := 5 // Number of concurrent registration processes
	var successes atomic.Int32
	var wg sync.WaitGroup

	t.Run("users can register", func(t *testing.T) {
		for c := 0; c < numConnections; c++ {
			wg.Add(1)
			go func(cid int) {
				defer wg.Done()

				// Create unique identifier for each registration
				uniqueID := fmt.Sprintf("%d_%d", cid, time.Now().UnixNano())
				email := fmt.Sprintf("testuser_%s@example.com", uniqueID)
				firstName := "Test"
				lastName := fmt.Sprintf("User_%s", uniqueID)
				password := "Password123!"

				success, err := registerKeycloakUserViaForm(email, firstName, lastName, password)
				if err != nil || !success {
					t.Logf("Registration failed: %v", err)
					return
				}

				successes.Add(1)
			}(c)
		}

		wg.Wait()

		if int(successes.Load()) != numConnections {
			t.Errorf("Registration Successes: %d", successes.Load())
		}
	})
}

func BenchmarkKeycloakAuthentication(b *testing.B) {

	numConnections := 1
	requestsPerClient := b.N / numConnections

	var wg sync.WaitGroup
	successful := atomic.Int64{}
	limited := atomic.Int64{}

	startTime := time.Now()

	for c := 0; c < numConnections; c++ {

		wg.Add(1)
		go func(cid int) {
			defer wg.Done()

			token, err := getKeycloakToken(cid)
			if err != nil {
				return
			}

			req, _ := http.NewRequest("GET", "https://localhost:7443/api/v1/profile/details", nil)

			req.Header.Add("Authorization", "Bearer "+token)
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}

			reset(b)
			for i := 0; i < requestsPerClient; i++ {
				res, err := client.Do(req)
				if err != nil {
					continue
				}

				if res.StatusCode == 429 {
					limited.Add(1)
				} else if res.StatusCode == 200 {
					successful.Add(1)
				}

				res.Body.Close()
			}
		}(c)
	}

	wg.Wait()

	duration := b.Elapsed().Nanoseconds()
	rps := float64(b.N) / float64(duration) * float64(1e9)
	fmt.Printf("Requests per second: %.2f\n", rps)

	duration1 := time.Since(startTime).Seconds()
	fmt.Printf("Successful requests: %d (%.2f/sec)\n", successful.Load(), float64(successful.Load())/duration1)
	fmt.Printf("Rate limited requests: %d (%.2f%%)\n", limited.Load(), float64(limited.Load())/float64(b.N)*100)
	fmt.Printf("Total throughput: %.2f req/sec\n", float64(b.N)/duration1)
}
