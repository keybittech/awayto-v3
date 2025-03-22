package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// TestKeycloakAuthentication tests the Keycloak authentication flow
func TestKeycloakAuthentication(t *testing.T) {

	go main()

	time.Sleep(5 * time.Second)

	// Get token from Keycloak
	token, err := getKeycloakToken()
	if err != nil {
		t.Fatalf("Failed to get Keycloak token: %v", err)
	}

	// Validate that we got a non-empty token
	if token == "" {
		t.Fatalf("Empty token received")
	}

	t.Logf("Successfully obtained token: %s...", token[:20])

	// Now use the token to access your protected endpoint
	req, err := http.NewRequest("GET", "https://localhost:7443/api/v1/profile/details", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add the bearer token to authorization header
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the request
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to access protected endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Check that the response is successful
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	detailsBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading userinfo response: %v\n", err)
		return
	}

	fmt.Println("\nUser profile raw response:", string(detailsBody))
}

// Helper function to get a token from Keycloak
func getKeycloakToken() (string, error) {
	// Prepare the form data
	data := url.Values{}
	data.Set("client_id", os.Getenv("KC_CLIENT"))
	data.Set("username", "1@1")
	data.Set("password", "1")
	data.Set("grant_type", "password")

	// If your client is confidential and requires a secret
	// data.Set("client_secret", "your-client-secret")

	// Create the request
	req, err := http.NewRequest(
		"POST",
		"https://localhost:7443/auth/realms/awaytoexchange_realm/protocol/openid-connect/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}

	// Set the appropriate content type for form data
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Extract the access token
	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return token, nil
}
