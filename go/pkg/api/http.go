package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (a *API) RedirectHTTP(httpPort int) {
	httpRedirector := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	httpRedirectorMux := http.NewServeMux()

	httpRedirectorMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("%s%s", os.Getenv("APP_HOST_URL"), r.URL.Path), http.StatusMovedPermanently)
	})

	httpRedirector.Handler = httpRedirectorMux

	println("listening on ", strconv.Itoa(httpPort))

	err := httpRedirector.ListenAndServe()
	if err != nil {
		util.ErrCheck(err)
		return
	}
}
