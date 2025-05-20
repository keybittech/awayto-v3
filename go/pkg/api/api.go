package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const (
	API_READ_TIMEOUT  = 5 * time.Second
	API_WRITE_TIMEOUT = 15 * time.Second
	API_IDLE_TIMEOUT  = 15 * time.Second
)

type API struct {
	Server   *http.Server
	Cache    *util.Cache
	Handlers *handlers.Handlers
}

func NewAPI(httpsPort int) *API {
	h := handlers.NewHandlers()

	registerHandlers(h)

	// go func() {
	// 	ticker := time.NewTicker(time.Duration(5 * time.Second))
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		fmt.Printf("[%s] UserSessions: %d\n", time.Now().Format(time.RFC3339), c.UserSessions.Len())
	// 		fmt.Printf("[%s] Groups: %d\n", time.Now().Format(time.RFC3339), c.Groups.Len())
	// 		fmt.Printf("[%s] SubGroups: %d\n", time.Now().Format(time.RFC3339), c.SubGroups.Len())
	// 	}
	// }()

	return &API{
		Server: &http.Server{
			Addr:         fmt.Sprintf("[::]:%d", httpsPort),
			ReadTimeout:  API_READ_TIMEOUT,
			WriteTimeout: API_WRITE_TIMEOUT,
			IdleTimeout:  API_IDLE_TIMEOUT,
			Handler:      http.NewServeMux(),
		},
		Handlers: h,
		Cache:    h.Cache,
	}
}
