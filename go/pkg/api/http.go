package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (a *API) RedirectHTTP(httpPort int) {
	a.Redirect = &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	httpRedirectorMux := http.NewServeMux()

	httpRedirectorMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("%s%s", util.E_APP_HOST_URL, r.URL.Path), http.StatusMovedPermanently)
	})

	a.Redirect.Handler = httpRedirectorMux

	util.DebugLog.Println("listening on ", strconv.Itoa(httpPort))

	err := a.Redirect.ListenAndServe()
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}
}
