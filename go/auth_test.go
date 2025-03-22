package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
)

var err error

func BenchmarkKeycloakAuthentication(b *testing.B) {
	req, err := http.NewRequest("GET", "https://localhost:7443/api/v1/profile/details", nil)
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		client := &http.Client{Transport: tr}
		b.StartTimer()
		client.Do(req)
	}

	duration := b.Elapsed().Nanoseconds()
	rps := float64(b.N) / float64(duration) * float64(1e9)
	fmt.Printf("Requests per second: %.2f\n", rps)

}
