//go:build !dev

package api

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func setupStaticBuildOrProxy(a *API) {
	util.DebugLog.Println("Using build folder")

	staticDir := os.Getenv("PROJECT_DIR")

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
			_, err := w.Write([]byte(""))
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
			}
		}
	})))
}
