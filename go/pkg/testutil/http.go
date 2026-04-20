package testutil

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func DoAndRead(client *http.Client, req *http.Request) ([]byte, error) {
	if client == nil {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: TestTLS,
			},
		}
	}

	resp, err := client.Do(req) // #nosec G704
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func GetTestReq(method, target string, body io.Reader) *http.Request {
	var testReq *http.Request
	var err error

	parsedTarget, err := url.Parse(target)
	if err != nil || parsedTarget.Scheme == "" {
		log.Fatalf("invalid url: %s", err)
	}

	if parsedTarget.Hostname() != "localhost" {
		log.Fatalf("test hostname is not localhost")
	}

	testReq, err = http.NewRequest(method, "http://localhost", body)
	if err != nil {
		log.Fatalf("get test req failed, %s", err)
	}

	testReq.URL = parsedTarget
	testReq.Host = parsedTarget.Host

	testReq.RemoteAddr = "127.0.0.1:9999"
	testReq.Header.Set("Accept", "application/json")
	testReq.Header.Set("X-Tz", "America/Los_Angeles")
	testReq.Header.Set("User-Agent", "api unit test")
	if body != nil {
		testReq.Header.Set("Content-Type", "application/json")
	}
	return testReq
}

func CheckServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", util.E_APP_HOST_URL, nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	_, err = DoAndRead(nil, req)
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}
