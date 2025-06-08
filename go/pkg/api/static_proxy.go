//go:build dev

package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func setupStaticBuildOrProxy(a *API) {
	util.DebugLog.Println("Using live reload")

	devServerUrl, err := url.Parse(util.E_TS_DEV_SERVER_URL)
	if err != nil {
		fmt.Printf("please set TS_DEV_SERVER_URL %s", err.Error())
	}

	var proxy *httputil.ReverseProxy
	proxy = httputil.NewSingleHostReverseProxy(devServerUrl)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402
	}

	a.Server.Handler.(*http.ServeMux).Handle("GET /app/", http.StripPrefix("/app/",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if strings.Contains(req.Header.Get("Accept"), "text/html") {
				util.WriteNonceIntoBody(proxy, w, req)
			} else {
				proxy.ServeHTTP(w, req)
			}
		}),
	))
}
