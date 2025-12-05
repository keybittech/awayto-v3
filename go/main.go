package main

import (
	"crypto/tls"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/api"
	"github.com/keybittech/awayto-v3/go/pkg/crypto"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
)

func main() {
	util.ParseEnv()

	crypto.InitVault()

	util.DebugLog.Printf(
		"started with flags -httpPort=%d -httpsPort=%d -unixPath=%s -rateLimit=%d -rateLimitBurst=%d",
		util.E_GO_HTTP_PORT,
		util.E_GO_HTTPS_PORT,
		util.E_UNIX_AUTH_PATH,
		util.E_RATE_LIMIT,
		util.E_RATE_LIMIT_BURST,
	)

	server := api.NewAPI(util.E_GO_HTTPS_PORT)

	server.Server.TLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{ // keep if tls 1.2 needed in some future situation
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		NextProtos: []string{"h2", "http/1.1"},
	}

	go server.RedirectHTTP(util.E_GO_HTTP_PORT)

	go server.InitUnixServer(util.E_UNIX_AUTH_PATH)

	server.InitProtoHandlers()
	server.InitAuthProxy()
	server.InitSockServer()
	server.InitStatic()

	rateLimiter := api.NewRateLimit("api", rate.Limit(util.E_RATE_LIMIT), util.E_RATE_LIMIT_BURST, time.Duration(5*time.Minute))

	handler := server.Server.Handler
	handler = server.SecurityHeadersMiddleware(handler)
	handler = server.LimitMiddleware(rateLimiter)(handler)
	handler = server.VaultMiddleware(handler)
	handler = server.AccessRequestMiddleware(handler)
	server.Server.Handler = handler

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
	// 		fmt.Printf("[%s] App Stats: Groups: %d, SubGroups: %d, UserSessions: %d\n",
	// 			time.Now().Format(time.RFC3339),
	// 			server.Cache.Groups.Len(),
	// 			server.Cache.SubGroups.Len(),
	// 			server.Cache.UserSessions.Len(),
	// 		)
	// 	}
	// }()

	defer func() {
		server.Close()
	}()

	util.DebugLog.Printf("Listening on %d ", util.E_GO_HTTPS_PORT)
	util.DebugLog.Printf("Cert Locations: %s %s", util.E_CERT_LOC, util.E_CERT_KEY_LOC)

	err := server.Server.ListenAndServeTLS(util.E_CERT_LOC, util.E_CERT_KEY_LOC)
	if err != nil {
		util.ErrorLog.Println("LISTEN AND SERVE ERROR: ", err.Error())
		return
	}
}
