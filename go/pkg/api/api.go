package api

import (
	"fmt"
	"net"
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
	Server    *http.Server
	Redirect  *http.Server
	Cache     *util.Cache
	Handlers  *handlers.Handlers
	Unix      net.Listener
	CloseChan chan struct{}
}

func NewAPI(httpsPort int) *API {
	h := handlers.NewHandlers()

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
		Handlers:  h,
		Cache:     h.Cache,
		CloseChan: make(chan struct{}),
	}
}

func (a *API) Close() {
	close(a.CloseChan)
	a.Handlers.Database.DatabaseClient.Close()
	a.Handlers.Socket.Close()
	a.Handlers.Keycloak.Close()
	if err := a.Handlers.Redis.RedisClient.Close(); err != nil {
		util.ErrorLog.Printf("could not close redis client, err: %v", err)
	}

	if err := a.Unix.Close(); err != nil {
		util.ErrorLog.Printf("could not close unix listener, err: %v", err)
	}

	if err := a.Redirect.Close(); err != nil {
		util.ErrorLog.Printf("could not close redirect server, err: %v", err)
	}

	if err := a.Server.Close(); err != nil {
		util.ErrorLog.Printf("could not close primary server client, err: %v", err)
	}
}
