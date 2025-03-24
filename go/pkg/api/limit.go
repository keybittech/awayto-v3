package api

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type LimitedClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimit() (*sync.Mutex, map[string]*LimitedClient) {
	var (
		mu             sync.Mutex
		limitedClients = make(map[string]*LimitedClient)
	)

	return &mu, limitedClients
}

func LimitCleanup(mu *sync.Mutex, limitedClients map[string]*LimitedClient) {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, lc := range limitedClients {
			if time.Since(lc.lastSeen) > 3*time.Minute {
				delete(limitedClients, ip)
			}
		}
		mu.Unlock()
	}
}

func Limiter(w http.ResponseWriter, mu *sync.Mutex, limitedClients map[string]*LimitedClient, limit rate.Limit, burst int, identifier string) bool {
	mu.Lock()
	if _, found := limitedClients[identifier]; !found {
		limitedClients[identifier] = &LimitedClient{limiter: rate.NewLimiter(limit, burst)}
	}
	limitedClients[identifier].lastSeen = time.Now()
	if !limitedClients[identifier].limiter.Allow() {
		mu.Unlock()

		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(http.StatusText(http.StatusTooManyRequests)))
		return true
	}
	mu.Unlock()

	return false
}
