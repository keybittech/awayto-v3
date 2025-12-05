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

var wasmIntegrity string

func initWasmHash() {
	path := fmt.Sprintf("%s/ts/build/lib.wasm", util.E_PROJECT_DIR)
	hash, err := util.CalcFileIntegrity(path)
	if err != nil {
		util.ErrorLog.Printf("CRITICAL: could not calc wasm integrity, %v", err)
		wasmIntegrity = ""
	} else {
		wasmIntegrity = hash
		util.DebugLog.Printf("wasm integrity pinned, %s", wasmIntegrity)
	}
}

func setupStaticBuildOrProxy(a *API) {
	util.DebugLog.Println("Using build folder")

	initWasmHash()

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
			} else if strings.HasSuffix(req.URL.Path, ".wasm") {
				w.Header().Set("Content-Type", "application/wasm")
				fileServer.ServeHTTP(redirect, req)
			} else {
				req.URL.Path = "/"

				nonce, ok := req.Context().Value("CSP-Nonce").([]byte)
				if !ok {
					http.Error(w, "no csp nonce", http.StatusInternalServerError)
					return
				}

				replacements := map[string]string{
					"VITE_NONCE":          string(nonce),
					"VITE_WASM_INTEGRITY": wasmIntegrity,
				}

				util.WriteIndexHtml(fileServer, w, req, replacements)
				return
			}

			if redirect.StatusCode == http.StatusNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
		}),
	))
}
