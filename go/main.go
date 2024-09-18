package main

import (
	a "av3api/pkg/api"
	c "av3api/pkg/clients"
	h "av3api/pkg/handlers"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {

	api := &a.API{
		Server: &http.Server{
			Addr:         fmt.Sprintf("[::]:%d", httpsPort),
			ReadTimeout:  5 * time.Minute,
			WriteTimeout: 15 * time.Second,
		},
		Handlers: &h.Handlers{
			Ai:       c.InitAi(),
			Database: c.InitDatabase(),
			Redis:    c.InitRedis(),
			Keycloak: c.InitKeycloak(),
			Socket:   c.InitSocket(),
		},
	}

	go api.InitUnixServer(unixPath)

	mux := api.InitMux()

	api.InitAuthProxy(mux)

	api.InitSockServer(mux)

	api.InitStatic(mux)

	go api.RedirectHTTP(httpPort)

	defer api.Server.Close()

	certLoc := os.Getenv("CERT_LOC")
	keyLoc := os.Getenv("KEY_LOC")

	println("listening on ", httpsPort, "Cert Locations:", certLoc, keyLoc)

	err := api.Server.ListenAndServeTLS(certLoc, keyLoc)
	if err != nil {
		fmt.Printf("LISTEN AND SERVE ERROR: %s", err.Error())
		return
	}

}
