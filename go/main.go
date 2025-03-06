package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	a "github.com/keybittech/awayto-v3/go/pkg/api"
	c "github.com/keybittech/awayto-v3/go/pkg/clients"
	h "github.com/keybittech/awayto-v3/go/pkg/handlers"
)

var (
	httpPort         int
	httpsPort        int
	turnListenerPort int
	turnInternalPort int
	unixPath         string
)

var (
	httpPortDefault         = 7080
	httpsPortDefault        = 7443
	turnListenerPortDefault = 7788
	turnInternalPortDefault = 3478
	unixPathDefault         = "/tmp/goapp.sock"
)

var (
	httpPortFlag         = flag.Int("httpPort", httpPortDefault, "Server HTTP port")
	httpsPortFlag        = flag.Int("httpsPort", httpsPortDefault, "Server HTTPS port")
	turnListenerPortFlag = flag.Int("turnListenerPort", turnListenerPortDefault, "Turn listener port")
	turnInternalPortFlag = flag.Int("turnInternalPort", turnInternalPortDefault, "Turn internal port")
	unixPathFlag         = flag.String("unixPath", unixPathDefault, "Unix socket path")
)

func init() {
	httpPort = *httpPortFlag
	httpsPort = *httpsPortFlag
	turnListenerPort = *turnListenerPortFlag
	turnInternalPort = *turnInternalPortFlag
	unixPath = *unixPathFlag

	godotenv.Load(os.Getenv("GO_ENVFILE_LOC"))

	if httpPortEnv := os.Getenv("GO_HTTP_PORT"); httpPortEnv != "" && *httpPortFlag == httpPortDefault {
		httpPortEnvI, err := strconv.Atoi(httpPortEnv)
		if err != nil {
			fmt.Printf("please set GO_HTTP_PORT as int %s", err.Error())
		} else {
			httpPort = httpPortEnvI
			fmt.Printf("set custom port %d\n", httpPort)
		}
	}

	if httpsPortEnv := os.Getenv("GO_HTTPS_PORT"); httpsPortEnv != "" && *httpsPortFlag == httpsPortDefault {
		httpsPortEnvI, err := strconv.Atoi(httpsPortEnv)
		if err != nil {
			fmt.Printf("please set GO_HTTPS_PORT as int %s", err.Error())
		} else {
			httpsPort = httpsPortEnvI
			fmt.Printf("set custom port %d\n", httpsPort)
		}
	}

	if unixSockDir, unixSockFile := os.Getenv("UNIX_SOCK_DIR"), os.Getenv("UNIX_SOCK_FILE"); unixPath == unixPathDefault && unixSockDir != "" && unixSockFile != "" {
		newUnixPath := fmt.Sprintf("%s/%s", unixSockDir, unixSockFile)
		unixPath = newUnixPath
		fmt.Printf("set custom path %s\n", unixPath)
	}
}

func main() {

	flag.Parse()

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
	keyLoc := os.Getenv("CERT_KEY_LOC")

	println("listening on ", httpsPort, "Cert Locations:", certLoc, keyLoc)

	err := api.Server.ListenAndServeTLS(certLoc, keyLoc)
	if err != nil {
		fmt.Printf("LISTEN AND SERVE ERROR: %s", err.Error())
		return
	}

}
