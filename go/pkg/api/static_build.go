//go:build !dev

package api

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func setupStaticBuildOrProxy(a *API) {
	util.DebugLog.Println("Using build folder")

	fileServer := http.FileServer(http.Dir(fmt.Sprintf("%s/ts/build/", util.E_PROJECT_DIR)))

	a.Server.Handler.(*http.ServeMux).Handle("GET /app/", http.StripPrefix("/app",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			redirect := &StaticRedirect{ResponseWriter: w}

			if slices.Contains([]string{".js", ".css", ".mjs"}, filepath.Ext(req.URL.Path)) {
				w.Header().Set("Cache-Control", maxAgeStr)
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
				util.WriteNonceIntoBody(fileServer, w, req)
				return
			}

			if redirect.StatusCode == http.StatusNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
		}),
	))
}
