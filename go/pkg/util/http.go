package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ErrRes struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

func Get(url string, headers http.Header) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func GetWithParams(url string, headers http.Header, queryParams url.Values) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, strings.NewReader(queryParams.Encode()))
	if err != nil {
		return nil, err
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func Mutate(method string, url string, headers http.Header, dataBody []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(dataBody))
	if err != nil {
		return nil, err
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 204 {
		return respBody, nil
	}

	if json.Valid(respBody) {
		var errRes ErrRes

		err = json.Unmarshal(respBody, &errRes)
		if err != nil {
			return nil, err
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

func PostFormData(url string, headers http.Header, data url.Values) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// println(req.URL.String(), data.Encode())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
