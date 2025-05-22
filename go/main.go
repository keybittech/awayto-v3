package main

import (
	"log"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/api"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
)

func main() {
	util.ParseEnv()

	util.DebugLog.Printf(
		"started with flags -httpPort=%d -httpsPort=%d -unixPath=%s -rateLimit=%d -rateLimitBurst=%d",
		util.E_GO_HTTP_PORT,
		util.E_GO_HTTPS_PORT,
		util.E_UNIX_PATH,
		util.E_RATE_LIMIT,
		util.E_RATE_LIMIT_BURST,
	)

	server := api.NewAPI(util.E_GO_HTTPS_PORT)

	go server.RedirectHTTP(util.E_GO_HTTP_PORT)

	go server.InitUnixServer(util.E_UNIX_PATH)

	server.InitProtoHandlers()
	server.InitAuthProxy()
	server.InitSockServer()
	server.InitStatic()

	rateLimiter := api.NewRateLimit("api", rate.Limit(util.E_RATE_LIMIT), util.E_RATE_LIMIT_BURST, time.Duration(5*time.Minute))
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

	util.DebugLog.Printf("Listening on %d ", util.E_GO_HTTPS_PORT)
	util.DebugLog.Printf("Cert Locations: %s %s", util.E_CERT_LOC, util.E_CERT_KEY_LOC)

	err := server.Server.ListenAndServeTLS(util.E_CERT_LOC, util.E_CERT_KEY_LOC)
	if err != nil {
		util.ErrorLog.Println("LISTEN AND SERVE ERROR: ", err.Error())
		return
	}
}
