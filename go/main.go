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
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/util"
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
	util.MakeLoggers()
	// flag.Parse()

	server := &api.API{
		Server: &http.Server{
			Addr:         fmt.Sprintf("[::]:%d", httpsPort),
			ReadTimeout:  time.Minute,
			WriteTimeout: time.Minute,
			IdleTimeout:  time.Minute,
		},
		Handlers: handlers.NewHandlers(),
	}

	go server.InitUnixServer(unixPath)

	mux := server.InitMux()

	server.InitAuthProxy(mux)

	server.InitSockServer(mux)

	server.InitStatic(mux)

	go server.RedirectHTTP(httpPort)

	stopChan := make(chan struct{}, 1)
	go setupGc(server, stopChan)

	defer func() {
		close(stopChan)
		server.Server.Close()
	}()

	certLoc := os.Getenv("CERT_LOC")
	keyLoc := os.Getenv("CERT_KEY_LOC")

	println("listening on ", httpsPort, "\nCert Locations:", certLoc, keyLoc)

	err := server.Server.ListenAndServeTLS(certLoc, keyLoc)
	if err != nil {
		println("LISTEN AND SERVE ERROR: ", err.Error())
		return
	}
}
