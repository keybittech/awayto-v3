package api

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Adapted from https://blog.logrocket.com/rate-limiting-go-application

type RateLimiter struct {
	Name           string
	Mu             sync.Mutex
	LimitedClients map[string]*LimitedClient
	ExpiryDuration time.Duration
	LimitNum       rate.Limit
	Burst          int
}

var RateLimiters sync.Map

type LimitedClient struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func NewRateLimit(name string, limit rate.Limit, burst int, expiryDuration time.Duration) *RateLimiter {
	rateLimiter := &RateLimiter{
		Name:           name,
		LimitedClients: make(map[string]*LimitedClient),
		ExpiryDuration: expiryDuration,
		LimitNum:       limit,
		Burst:          burst,
	}

	RateLimiters.Store(name, rateLimiter)

	return rateLimiter
}

func (rl *RateLimiter) Limit(identifier string) bool {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	var client *LimitedClient
	if existingClient, found := rl.LimitedClients[identifier]; found {
		client = existingClient
	} else {
		client = &LimitedClient{
			Limiter: rate.NewLimiter(rl.LimitNum, rl.Burst),
		}
		rl.LimitedClients[identifier] = client
	}

	client.LastSeen = time.Now()
	return !client.Limiter.Allow()
}

func (rl *RateLimiter) Cleanup() {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	now := time.Now()
	for id, client := range rl.LimitedClients {
		if now.Sub(client.LastSeen) > rl.ExpiryDuration {
			client.LastSeen = time.Time{}
			delete(rl.LimitedClients, id)
		}
	}
}
