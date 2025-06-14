package util

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
)

type ErrRes struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

func successStatus(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func Get(url string, headers http.Header) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCheck(err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrCheck(err)
	}

	defer resp.Body.Close()

	if !successStatus(resp.StatusCode) {
		return nil, ErrCheck(errors.New(http.StatusText(resp.StatusCode)))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrCheck(err)
	}

	return respBody, nil
}

func GetWithParams(url string, headers http.Header, queryParams url.Values) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCheck(err)
	}
	req.URL.RawQuery = queryParams.Encode()

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrCheck(err)
	}
	defer resp.Body.Close()

	if !successStatus(resp.StatusCode) {
		return nil, ErrCheck(errors.New(http.StatusText(resp.StatusCode)))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrCheck(err)
	}

	return respBody, nil
}

func Mutate(method string, url string, headers http.Header, dataBody []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(dataBody))
	if err != nil {
		return nil, ErrCheck(err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrCheck(err)
	}

	defer resp.Body.Close()

	if !successStatus(resp.StatusCode) {
		return nil, ErrCheck(errors.New(http.StatusText(resp.StatusCode)))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrCheck(err)
	}

	if resp.StatusCode == 204 {
		return respBody, nil
	}

	if len(respBody) > 2 {
		var errRes ErrRes

		err = json.Unmarshal(respBody, &errRes)
		if err != nil {
			return nil, ErrCheck(err)
		}

		if errRes.Error != "" {
			return nil, errors.New(errRes.Error)
		}

		if errRes.ErrorMessage != "" {
			return nil, errors.New(errRes.ErrorMessage)
		}
	}

	return respBody, nil
}

func PostFormData(ctx context.Context, url string, headers http.Header, data io.Reader) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "POST", url, data)
	if err != nil {
		return nil, ErrCheck(err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrCheck(err)
	}

	defer resp.Body.Close()

	if !successStatus(resp.StatusCode) {
		errBytes := make([]byte, 1024)
		resp.Body.Read(errBytes)
		err := fmt.Errorf("bad status %d, err: %s", resp.StatusCode, errBytes)
		return nil, ErrCheck(err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrCheck(err)
	}

	return respBody, nil
}

func WriteNonceIntoBody(handler http.Handler, w http.ResponseWriter, req *http.Request) {
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	originalBody, err := io.ReadAll(recorder.Body)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	nonce, ok := req.Context().Value("CSP-Nonce").([]byte)
	if !ok {
		http.Error(w, "No csp nonce for index build", http.StatusInternalServerError)
		return
	}
	modifiedBody := regexp.MustCompile(`VITE_NONCE`).ReplaceAll(originalBody, nonce)

	_, err = w.Write(modifiedBody)
	if err != nil {
		ErrorLog.Println(ErrCheck(err))
	}
}
