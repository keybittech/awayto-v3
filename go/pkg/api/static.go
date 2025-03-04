package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/util"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
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
	devServerUrl, err := url.Parse(os.Getenv("TS_DEV_SERVER_URL"))
	if err != nil {
		fmt.Printf("please set TS_DEV_SERVER_URL %s", err.Error())
	}

	// Attach landing/ to domain url root /
	landingFiles := http.FileServer(http.Dir("landing/public/"))
	mux.Handle("/", http.StripPrefix("/",
		a.LimitMiddleware(10, 10)(
			func(w http.ResponseWriter, r *http.Request, session *clients.UserSession) {
				landingFiles.ServeHTTP(w, r)
			},
		),
	))

	// Attach demos
	demoFiles := http.FileServer(http.Dir("demos/final/"))
	mux.Handle("/demos/", http.StripPrefix("/demos/",
		a.LimitMiddleware(.1, 1)(
			func(w http.ResponseWriter, req *http.Request, session *clients.UserSession) {
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

		fileServer := http.FileServer(http.Dir("ts/build/"))

		mux.Handle("GET /app/", http.StripPrefix("/app/",
			a.LimitMiddleware(10, 10)(
				func(w http.ResponseWriter, req *http.Request, session *clients.UserSession) {
					redirect := &StaticRedirect{ResponseWriter: w}
					if !strings.HasSuffix(req.URL.Path, ".js") {
						fileServer.ServeHTTP(redirect, req)
					} else {
						w.Header().Set("Content-Encoding", "gzip")
						gz := gzip.NewWriter(w)
						defer gz.Close()
						gzr := StaticGzip{Writer: gz, ResponseWriter: redirect}
						fileServer.ServeHTTP(gzr, req)
					}
					if redirect.StatusCode == http.StatusNotFound {
						req.URL.Path = "/"
						redirect.Header().Set("Content-Type", "text/html")
						fileServer.ServeHTTP(w, req)
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
			// a.LimitMiddleware(100, 200)(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				proxy.ServeHTTP(w, req)
			}),
			// ),
		))
	}
}
