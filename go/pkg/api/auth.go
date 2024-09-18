package api

import (
	"av3api/pkg/util"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var authTransport = http.DefaultTransport

func (a *API) InitAuthProxy(mux *http.ServeMux) {

	kcInternal := os.Getenv("KC_INTERNAL")

	mux.Handle("/auth/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		proxyPath := fmt.Sprintf("%s%s", kcInternal, strings.TrimPrefix(r.URL.Path, "/auth"))
		proxyURL, err := url.Parse(proxyPath)
		if err != nil {
			util.ErrCheck(err)
			http.Error(w, "bad request", http.StatusInternalServerError)
			return
		}
		proxyURL.RawQuery = r.URL.RawQuery

		proxyReq, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
		if err != nil {
			http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
			return
		}

		// Copy the headers from the original request to the proxy request
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		// Send the proxy request using the custom transport
		resp, err := authTransport.RoundTrip(proxyReq)
		if err != nil {
			http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy the headers from the proxy response to the original response
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Set the status code of the original response to the status code of the proxy response
		w.WriteHeader(resp.StatusCode)

		// Copy the body of the proxy response to the original response
		io.Copy(w, resp.Body)

	}))

}
