package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

var token string

func TestMain(t *testing.T) {
	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}

	go main()

	time.Sleep(5 * time.Second)
	println("did setup main")

	// Get token from Keycloak
	token, err = getKeycloakToken()
	if err != nil {
		log.Fatalf("Failed to get Keycloak token: %v", err)
	}

	// Validate that we got a non-empty token
	if token == "" {
		log.Fatalf("Empty token received")
	}
}

func getKeycloakToken() (string, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("KC_CLIENT"))
	data.Set("username", "1@1")
	data.Set("password", "1")
	data.Set("grant_type", "password")

	req, err := http.NewRequest(
		"POST",
		"https://localhost:7443/auth/realms/awaytoexchange_realm/protocol/openid-connect/token",
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
