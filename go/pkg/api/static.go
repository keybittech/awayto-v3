package api

import (
	"av3api/pkg/util"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func (a *API) StaticAuthCheckMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// session := a.Handlers.Redis.ReqSession(req)
		// TODO integreate session.UserRoles
		hasRole := true

		if hasRole {
			next.ServeHTTP(w, req)
		} else {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}
	}
}

type StaticRedirect struct {
	http.ResponseWriter
	StatusCode int
}

func (sr *StaticRedirect) Write(b []byte) (int, error) {
	if sr.StatusCode == http.StatusNotFound {
		return len(b), nil
	}
	if sr.StatusCode != 0 {
		sr.WriteHeader(sr.StatusCode)
	}
	return sr.ResponseWriter.Write(b)
}

func (sr *StaticRedirect) WriteHeader(code int) {
	if code > 300 && code < 400 {
		sr.ResponseWriter.WriteHeader(code)
		return
	}
	sr.StatusCode = code
}

func (a *API) InitStatic(mux *http.ServeMux) {
	devServerUrl, err := url.Parse(os.Getenv("TS_DEV_SERVER_URL"))
	if err != nil {
		fmt.Printf("please set TS_DEV_SERVER_URL %s", err.Error())
	}

	// Attach landing/ to domain url root /
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("landing/build/"))))

	// use dev server or built for /app
	_, err = http.Get(devServerUrl.String())

	if err != nil && !strings.Contains(err.Error(), "failed to verify certificate") {
		util.ErrCheck(err)
		println("Using build folder")

		fileServer := http.FileServer(http.Dir("ts/build/"))

		mux.Handle("GET /app/", a.StaticAuthCheckMiddleware(http.StripPrefix("/app/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirect := &StaticRedirect{ResponseWriter: w}
			fileServer.ServeHTTP(redirect, r)
			if redirect.StatusCode == http.StatusNotFound {
				r.URL.Path = "/"
				redirect.Header().Set("Content-Type", "text/html")
				fileServer.ServeHTTP(w, r)
			}
		}))))
	} else {
		println("Using live reload")
		var proxy *httputil.ReverseProxy
		proxy = httputil.NewSingleHostReverseProxy(devServerUrl)
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		mux.Handle("GET /app/", a.StaticAuthCheckMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})))

		var wsDevProxy *httputil.ReverseProxy
		wsDevProxy = httputil.NewSingleHostReverseProxy(devServerUrl)
		wsDevProxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		mux.Handle("GET /ws", a.StaticAuthCheckMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsDevProxy.ServeHTTP(w, r)
		})))
	}
}
