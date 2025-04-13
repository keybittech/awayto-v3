package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/keybittech/awayto-v3/go/pkg/api"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
)

var (
	httpPort         = 7080
	httpsPort        = 7443
	turnListenerPort = 7788
	turnInternalPort = 3478
	unixPath         = "/tmp/goapp.sock"
)

var (
	httpPortFlag         = flag.Int("httpPort", httpPort, "Server HTTP port")
	httpsPortFlag        = flag.Int("httpsPort", httpsPort, "Server HTTPS port")
	turnListenerPortFlag = flag.Int("turnListenerPort", turnListenerPort, "Turn listener port")
	turnInternalPortFlag = flag.Int("turnInternalPort", turnInternalPort, "Turn internal port")
	unixPathFlag         = flag.String("unixPath", unixPath, "Unix socket path")
)

var mainApi *api.API

func init() {
	httpPort = *httpPortFlag
	httpsPort = *httpsPortFlag
	turnListenerPort = *turnListenerPortFlag
	turnInternalPort = *turnInternalPortFlag
	unixPath = *unixPathFlag

	godotenv.Load(os.Getenv("GO_ENVFILE_LOC"))

	if httpPortEnv := os.Getenv("GO_HTTP_PORT"); httpPortEnv != "" && *httpPortFlag == httpPort {
		httpPortEnvI, err := strconv.Atoi(httpPortEnv)
		if err != nil {
			fmt.Printf("please set GO_HTTP_PORT as int %s", err.Error())
		} else {
			httpPort = httpPortEnvI
			fmt.Printf("set custom port %d\n", httpPort)
		}
	}

	if httpsPortEnv := os.Getenv("GO_HTTPS_PORT"); httpsPortEnv != "" && *httpsPortFlag == httpsPort {
		httpsPortEnvI, err := strconv.Atoi(httpsPortEnv)
		if err != nil {
			fmt.Printf("please set GO_HTTPS_PORT as int %s", err.Error())
		} else {
			httpsPort = httpsPortEnvI
			fmt.Printf("set custom port %d\n", httpsPort)
		}
	}

	if unixSockDir, unixSockFile := os.Getenv("UNIX_SOCK_DIR"), os.Getenv("UNIX_SOCK_FILE"); unixPath == unixPath && unixSockDir != "" && unixSockFile != "" {
		unixPath = fmt.Sprintf("%s/%s", unixSockDir, unixSockFile)
		fmt.Printf("set custom path %s\n", unixPath)
	}
}

func main() {
	// flag.Parse()

	mainApi = &api.API{
		Server: &http.Server{
			Addr:         fmt.Sprintf("[::]:%d", httpsPort),
			ReadTimeout:  time.Minute,
			WriteTimeout: time.Minute,
			IdleTimeout:  time.Minute,
		},
		Handlers: &handlers.Handlers{
			Ai:       clients.InitAi(),
			Database: clients.InitDatabase(),
			Redis:    clients.InitRedis(),
			Keycloak: clients.InitKeycloak(),
			Socket:   clients.InitSocket(),
		},
	}

	go mainApi.InitUnixServer(unixPath)

	mux := mainApi.InitMux()

	mainApi.InitAuthProxy(mux)

	mainApi.InitSockServer(mux)

	mainApi.InitStatic(mux)

	go mainApi.RedirectHTTP(httpPort)

	defer mainApi.Server.Close()

	setupGc()

	certLoc := os.Getenv("CERT_LOC")
	keyLoc := os.Getenv("CERT_KEY_LOC")

	println("listening on ", httpsPort, "Cert Locations:", certLoc, keyLoc)

	err := mainApi.Server.ListenAndServeTLS(certLoc, keyLoc)
	if err != nil {
		fmt.Printf("LISTEN AND SERVE ERROR: %s", err.Error())
		return
	}

}
