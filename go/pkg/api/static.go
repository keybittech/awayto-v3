package api

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const maxAge = 60 * 60 * 24 * 30

var (
	maxAgeStr = strconv.Itoa(maxAge)
	maxAgeDur = time.Duration(maxAge) * time.Second
)

type StaticGzip struct {
	io.Writer
	http.ResponseWriter
}

func (g StaticGzip) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
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

func (a *API) InitStatic() {
	staticDir := os.Getenv("PROJECT_DIR")

	devServerUrl, err := url.Parse(os.Getenv("TS_DEV_SERVER_URL"))
	if err != nil {
		fmt.Printf("please set TS_DEV_SERVER_URL %s", err.Error())
	}

	// Attach landing/ to domain url root /
	landingFiles := http.FileServer(http.Dir(fmt.Sprintf("%s/landing/public/", staticDir)))
	a.Server.Handler.(*http.ServeMux).Handle("/", http.StripPrefix("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age="+maxAgeStr)
		w.Header().Set("Expires", time.Now().Add(maxAgeDur).UTC().Format(http.TimeFormat))
		landingFiles.ServeHTTP(w, r)
	})))

	// Attach demos
	demoFiles := http.FileServer(http.Dir(fmt.Sprintf("%s/demos/final/", staticDir)))
	demoRl := NewRateLimit("demos", .1, 1, time.Duration(5*time.Minute))
	a.Server.Handler.(*http.ServeMux).Handle("GET /demos/", http.StripPrefix("/demos/",
		a.LimitMiddleware(demoRl)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if util.CookieExpired(req) {
				util.ErrorLog.Println(util.ErrCheck(err))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}

			w.Header().Set("Cache-Control", "public, max-age="+maxAgeStr)
			w.Header().Set("Expires", time.Now().Add(maxAgeDur).UTC().Format(http.TimeFormat))

			demoFiles.ServeHTTP(w, req)
		})),
	))

	// use dev server or built for /app
	_, err = http.Get(devServerUrl.String())
	if err != nil && !strings.Contains(err.Error(), "failed to verify certificate") {
		println("Using build folder")

		fileServer := http.FileServer(http.Dir(fmt.Sprintf("%s/ts/build/", staticDir)))

		a.Server.Handler.(*http.ServeMux).Handle("GET /app/", http.StripPrefix("/app", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			redirect := &StaticRedirect{ResponseWriter: w}

			if strings.HasSuffix(req.URL.Path, ".js") || strings.HasSuffix(req.URL.Path, ".css") || strings.HasSuffix(req.URL.Path, ".mjs") {
				w.Header().Set("Cache-Control", "public, max-age="+maxAgeStr)
				w.Header().Set("Expires", time.Now().Add(maxAgeDur).UTC().Format(http.TimeFormat))
				w.Header().Set("Content-Encoding", "gzip")
				gz := gzip.NewWriter(w)
				defer gz.Close()
				gzr := StaticGzip{Writer: gz, ResponseWriter: redirect}
				fileServer.ServeHTTP(gzr, req)
			} else if strings.HasSuffix(req.URL.Path, ".png") {
				fileServer.ServeHTTP(redirect, req)
			} else {
				req.URL.Path = "/"

				nonceData := uuid.NewString()
				nonceB64 := base64.StdEncoding.EncodeToString([]byte(nonceData))

				w.Header().Set(
					"Content-Security-Policy",
					"object-src 'none';"+
						"script-src 'nonce-"+nonceB64+"' 'strict-dynamic';"+
						"base-uri 'none';")
				recorder := httptest.NewRecorder()

				fileServer.ServeHTTP(recorder, req)

				// Read the response body
				originalBody, err := io.ReadAll(recorder.Body)
				if err != nil {
					http.Error(w, "Error reading file", http.StatusInternalServerError)
					return
				}

				modifiedBody := regexp.MustCompile(`VITE_NONCE`).ReplaceAll(originalBody, []byte(nonceB64))

				_, err = w.Write(modifiedBody)
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
				}
				return
			}

			if redirect.StatusCode == http.StatusNotFound {
				w.WriteHeader(http.StatusNotFound)
				_, err = w.Write([]byte(""))
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
				}
			}
		})))
	} else {
		println("Using live reload")
		var proxy *httputil.ReverseProxy
		proxy = httputil.NewSingleHostReverseProxy(devServerUrl)
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402
		}

		a.Server.Handler.(*http.ServeMux).Handle("GET /app/", http.StripPrefix("/app/",
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				proxy.ServeHTTP(w, req)
			}),
		))
	}
}
