package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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
				success, err := registerKeycloakUserViaForm(int(time.Now().UnixNano()))
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
	time.Sleep(time.Second)
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

			token, _, err := getKeycloakToken(cid)
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
