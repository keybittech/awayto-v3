package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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
	// sockPort         = flag.Int("sockPort", 6677, "Socket port")
	httpPortFlag         = flag.Int("httpPort", httpPortDefault, "Server HTTP port")
	httpsPortFlag        = flag.Int("httpsPort", httpsPortDefault, "Server HTTPS port")
	turnListenerPortFlag = flag.Int("turnListenerPort", turnListenerPortDefault, "Turn listener port")
	turnInternalPortFlag = flag.Int("turnInternalPort", turnInternalPortDefault, "Turn internal port")
	unixPathFlag         = flag.String("unixPath", unixPathDefault, "Unix socket path")
)

func ParseFlags() {
	flag.Parse()

	httpPort = *httpPortFlag
	httpsPort = *httpsPortFlag
	turnListenerPort = *turnListenerPortFlag
	turnInternalPort = *turnInternalPortFlag
	unixPath = *unixPathFlag

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
