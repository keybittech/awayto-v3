package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

// 30 days
const maxAge = 60 * 60 * 24 * 30

var (
	maxAgeStr = "public, max-age=" + strconv.Itoa(maxAge)
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
	// Attach landing/ to domain url root /
	landingFiles := http.FileServer(http.Dir(fmt.Sprintf("%s/landing/public/", util.E_PROJECT_DIR)))
	a.Server.Handler.(*http.ServeMux).Handle("/", http.StripPrefix("/",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Cache-Control", maxAgeStr)
			w.Header().Set("Expires", time.Now().Add(maxAgeDur).UTC().Format(http.TimeFormat))
			if req.URL.Path == "" {
				util.WriteNonceIntoBody(landingFiles, w, req)
			} else {
				landingFiles.ServeHTTP(w, req)
			}
		}),
	))

	// Attach demos
	demoFiles := http.FileServer(http.Dir(fmt.Sprintf("%s/demos/final/", util.E_PROJECT_DIR)))
	demoRl := NewRateLimit("demos", .1, 1, time.Duration(5*time.Minute))
	a.Server.Handler.(*http.ServeMux).Handle("GET /demos/", http.StripPrefix("/demos/",
		a.LimitMiddleware(demoRl)(
			a.ValidateSessionMiddleware()(
				SessionHandler(func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
					w.Header().Set("Cache-Control", maxAgeStr)
					w.Header().Set("Expires", time.Now().Add(maxAgeDur).UTC().Format(http.TimeFormat))
					demoFiles.ServeHTTP(w, req)
				}),
			),
		),
	))

	setupStaticBuildOrProxy(a)
}
