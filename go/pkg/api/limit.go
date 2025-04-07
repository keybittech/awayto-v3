package api

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Adapted from https://blog.logrocket.com/rate-limiting-go-application

type RateLimiter struct {
	Name           string
	Mu             *sync.Mutex
	LimitedClients map[string]*LimitedClient
}

var RateLimiters sync.Map

type LimitedClient struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func NewRateLimit(name string) (*sync.Mutex, map[string]*LimitedClient) {
	var (
		mu             sync.Mutex
		limitedClients = make(map[string]*LimitedClient)
	)

	RateLimiters.Store(name, &RateLimiter{name, &mu, limitedClients})

	return &mu, limitedClients
}

func Limiter(mu *sync.Mutex, limitedClients map[string]*LimitedClient, limit rate.Limit, burst int, identifier string) bool {
	mu.Lock()
	if _, found := limitedClients[identifier]; !found {
		limitedClients[identifier] = &LimitedClient{Limiter: rate.NewLimiter(limit, burst)}
	}
	limitedClients[identifier].LastSeen = time.Now()
	if !limitedClients[identifier].Limiter.Allow() {
		mu.Unlock()
		return true
	}
	mu.Unlock()
	return false
}

func WriteLimit(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(http.StatusText(http.StatusTooManyRequests)))
}
