package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/keybittech/awayto-v3/go/pkg/api"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
)

var (
	httpPort               = 7080
	httpsPort              = 7443
	turnListenerPort       = 7788
	turnInternalPort       = 3478
	unixPath               = "/tmp/goapp.sock"
	requestsPerSecond      = 50
	requestsPerSecondBurst = 50
)

var (
	httpPortFlag               = flag.Int("httpPort", httpPort, "Server HTTP port")
	httpsPortFlag              = flag.Int("httpsPort", httpsPort, "Server HTTPS port")
	turnListenerPortFlag       = flag.Int("turnListenerPort", turnListenerPort, "Turn listener port")
	turnInternalPortFlag       = flag.Int("turnInternalPort", turnInternalPort, "Turn internal port")
	unixPathFlag               = flag.String("unixPath", unixPath, "Unix socket path")
	requestsPerSecondFlag      = flag.Int("requestsPerSecond", requestsPerSecond, "Turn internal port")
	requestsPerSecondBurstFlag = flag.Int("requestsPerSecondBurst", requestsPerSecondBurst, "Turn internal port")
)

func init() {
	httpPort = *httpPortFlag
	httpsPort = *httpsPortFlag
	turnListenerPort = *turnListenerPortFlag
	turnInternalPort = *turnInternalPortFlag
	unixPath = *unixPathFlag
	requestsPerSecond = *requestsPerSecondFlag
	requestsPerSecondBurst = *requestsPerSecondBurstFlag

	err := godotenv.Load(os.Getenv("GO_ENVFILE_LOC"))
	if err != nil {
		log.Fatal(err)
	}

	httpPortEnv := os.Getenv("GO_HTTP_PORT")
	if httpPortEnv != "" && *httpPortFlag == httpPort {
		httpPortEnvI, err := strconv.Atoi(httpPortEnv)
		if err != nil {
			fmt.Printf("please set GO_HTTP_PORT as int %s", err.Error())
		} else {
			httpPort = httpPortEnvI
			fmt.Printf("set custom port %d\n", httpPort)
		}
	}

	httpsPortEnv := os.Getenv("GO_HTTPS_PORT")
	if httpsPortEnv != "" && *httpsPortFlag == httpsPort {
		httpsPortEnvI, err := strconv.Atoi(httpsPortEnv)
		if err != nil {
			fmt.Printf("please set GO_HTTPS_PORT as int %s", err.Error())
		} else {
			httpsPort = httpsPortEnvI
			fmt.Printf("set custom port %d\n", httpsPort)
		}
	}

	unixSockDir, unixSockFile := os.Getenv("UNIX_SOCK_DIR"), os.Getenv("UNIX_SOCK_FILE")
	if unixPath == unixPath && unixSockDir != "" && unixSockFile != "" {
		unixPath = fmt.Sprintf("%s/%s", unixSockDir, unixSockFile)
		fmt.Printf("set custom path %s\n", unixPath)
	}
}

func main() {
	util.MakeLoggers()

	server := api.NewAPI(httpsPort)

	go server.RedirectHTTP(httpPort)

	go server.InitUnixServer(unixPath)

	server.InitProtoHandlers()
	server.InitAuthProxy()
	server.InitSockServer()
	server.InitStatic()

	rateLimiter := api.NewRateLimit("api", rate.Limit(requestsPerSecond), requestsPerSecondBurst, time.Duration(5*time.Minute))
	limitMiddleware := server.LimitMiddleware(rateLimiter)(server.Server.Handler)
	server.Server.Handler = server.AccessRequestMiddleware(limitMiddleware)

	stopChan := make(chan struct{}, 1)
	go setupGc(server, stopChan)

	// go func() {
	// 	ticker := time.NewTicker(time.Duration(5 * time.Second))
	// 	defer ticker.Stop()
	// 	var m runtime.MemStats
	// 	for range ticker.C {
	// 		runtime.ReadMemStats(&m)
	// 		fmt.Printf("[%s] Runtime Stats: Goroutines: %d, Alloc: %v MiB, TotalAlloc: %v MiB, Sys: %v MiB, NumGC: %v\n",
	// 			time.Now().Format(time.RFC3339),
	// 			runtime.NumGoroutine(),
	// 			m.Alloc/1024/1024,
	// 			m.TotalAlloc/1024/1024,
	// 			m.Sys/1024/1024,
	// 			m.NumGC,
	// 		)
	// 	}
	// }()

	defer func() {
		close(stopChan)
		err := server.Server.Close()
		if err != nil {
			log.Fatal(err)
		}
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
