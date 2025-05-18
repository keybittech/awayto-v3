package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
)

type API struct {
	*http.Server
	*Cache
	*handlers.Handlers
}

func NewAPI(httpsPort int) *API {
	h := handlers.NewHandlers()

	registerHandlers(h)

	c := NewCache()

	return &API{
		Server: &http.Server{
			Addr:         fmt.Sprintf("[::]:%d", httpsPort),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  15 * time.Second,
			Handler:      http.NewServeMux(),
		},
		Handlers: h,
		Cache:    c,
	}
}
