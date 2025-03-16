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
	"strings"

	"github.com/google/uuid"
	"github.com/keybittech/awayto-v3/go/pkg/util"
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

func (a *API) InitStatic(mux *http.ServeMux) {
	staticDir := os.Getenv("PROJECT_DIR")

	devServerUrl, err := url.Parse(os.Getenv("TS_DEV_SERVER_URL"))
	if err != nil {
		fmt.Printf("please set TS_DEV_SERVER_URL %s", err.Error())
	}

	// Attach landing/ to domain url root /
	landingFiles := http.FileServer(http.Dir(fmt.Sprintf("%s/landing/public/", staticDir)))
	mux.Handle("/", http.StripPrefix("/",
		a.LimitMiddleware(10, 10)(
			func(w http.ResponseWriter, r *http.Request) {
				landingFiles.ServeHTTP(w, r)
			},
		),
	))

	// Attach demos
	demoFiles := http.FileServer(http.Dir(fmt.Sprintf("%s/demos/final/", staticDir)))
	mux.Handle("/demos/", http.StripPrefix("/demos/",
		a.LimitMiddleware(.1, 1)(
			func(w http.ResponseWriter, req *http.Request) {
				cookieValidation, err := req.Cookie("valid_signature")
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}

				err = util.VerifySigned(util.LOGIN_SIGNATURE_NAME, cookieValidation.Value)
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}

				demoFiles.ServeHTTP(w, req)
			},
		),
	))

	// use dev server or built for /app
	_, err = http.Get(devServerUrl.String())
	if err != nil && !strings.Contains(err.Error(), "failed to verify certificate") {
		util.ErrCheck(err)
		println("Using build folder")

		fileServer := http.FileServer(http.Dir(fmt.Sprintf("%s/ts/build/", staticDir)))

		mux.Handle("GET /app/", http.StripPrefix("/app",
			a.LimitMiddleware(10, 20)(
				func(w http.ResponseWriter, req *http.Request) {
					redirect := &StaticRedirect{ResponseWriter: w}

					if strings.HasSuffix(req.URL.Path, ".js") || strings.HasSuffix(req.URL.Path, ".css") || strings.HasSuffix(req.URL.Path, ".mjs") {
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

						w.Write(modifiedBody)
						return
					}

					if redirect.StatusCode == http.StatusNotFound {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(""))
					}
				},
			),
		))
	} else {
		println("Using live reload")
		var proxy *httputil.ReverseProxy
		proxy = httputil.NewSingleHostReverseProxy(devServerUrl)
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		mux.Handle("GET /app/", http.StripPrefix("/app/",
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				proxy.ServeHTTP(w, req)
			}),
		))
	}
}
