package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
)

type API struct {
	Server   *http.Server
	Handlers *handlers.Handlers
}

func NewAPI(httpsPort int) *API {
	return &API{
		Server: &http.Server{
			Addr:         fmt.Sprintf("[::]:%d", httpsPort),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  15 * time.Second,
			Handler:      http.NewServeMux(),
		},
		Handlers: handlers.NewHandlers(),
	}
}
