package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func getKeycloakToken(cid int) (string, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("KC_CLIENT"))
	data.Set("username", "1@"+strconv.Itoa(cid+1))
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
