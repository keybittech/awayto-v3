package api

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func BenchmarkRateLimit(b *testing.B) {
	rl := NewRateLimit("limit", 0, 0, time.Duration(5*time.Second))
	reset(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rl.Limit("test")
		}
	})
}

func TestNewRateLimit(t *testing.T) {
	type args struct {
		name           string
		limit          rate.Limit
		burst          int
		expiryDuration time.Duration
	}
	tests := []struct {
		name string
		args args
		want *RateLimiter
	}{
		{
			name: "Create API rate limiter",
			args: args{
				name:           "api",
				limit:          5,
				burst:          10,
				expiryDuration: time.Duration(5 * time.Minute),
			},
			want: &RateLimiter{
				Name:           "api",
				LimitedClients: make(map[string]*LimitedClient),
				ExpiryDuration: time.Duration(5 * time.Minute),
				LimitNum:       5,
				Burst:          10,
			},
		},
		{
			name: "Create DB rate limiter",
			args: args{
				name:           "db",
				limit:          10,
				burst:          20,
				expiryDuration: time.Duration(10 * time.Minute),
			},
			want: &RateLimiter{
				Name:           "db",
				LimitedClients: make(map[string]*LimitedClient),
				ExpiryDuration: time.Duration(10 * time.Minute),
				LimitNum:       10,
				Burst:          20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRateLimit(tt.args.name, tt.args.limit, tt.args.burst, tt.args.expiryDuration)

			// Custom comparison since we can't compare the mutex directly
			if got.Name != tt.want.Name {
				t.Errorf("NewRateLimit() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.ExpiryDuration != tt.want.ExpiryDuration {
				t.Errorf("NewRateLimit() ExpiryDuration = %v, want %v", got.ExpiryDuration, tt.want.ExpiryDuration)
			}
			if got.LimitNum != tt.want.LimitNum {
				t.Errorf("NewRateLimit() LimitNum = %v, want %v", got.LimitNum, tt.want.LimitNum)
			}
			if got.Burst != tt.want.Burst {
				t.Errorf("NewRateLimit() Burst = %v, want %v", got.Burst, tt.want.Burst)
			}

			// Verify it's stored in RateLimiters map
			storedRL, ok := RateLimiters.Load(tt.args.name)
			if !ok {
				t.Errorf("NewRateLimit() did not store rate limiter with name %v in RateLimiters map", tt.args.name)
			}
			if storedRL != got {
				t.Errorf("NewRateLimit() stored rate limiter %v is not the same as returned %v", storedRL, got)
			}
		})
	}
}

func TestRateLimiter_Limit(t *testing.T) {
	type args struct {
		identifier string
	}

	rl := NewRateLimit("test", 2, 2, time.Minute)

	tests := []struct {
		name string
		rl   *RateLimiter
		args args
		want bool
	}{
		{
			name: "First request not limited",
			rl:   rl,
			args: args{identifier: "client1"},
			want: false, // Allow() returns true, so Limit() returns !Allow()
		},
		{
			name: "Second request not limited",
			rl:   rl,
			args: args{identifier: "client1"},
			want: false,
		},
		{
			name: "Third request limited",
			rl:   rl,
			args: args{identifier: "client1"},
			want: true, // Should be limited now as we exceeded burst
		},
		{
			name: "Different client not limited",
			rl:   rl,
			args: args{identifier: "client2"},
			want: false, // New client should not be limited
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rl.Limit(tt.args.identifier); got != tt.want {
				t.Errorf("RateLimiter.Limit(%v) = %v, want %v", tt.args.identifier, got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimit("cleanup-test", 5, 5, 10*time.Millisecond)

	// Add a client
	rl.Limit("client1")

	// Verify client exists
	rl.Mu.Lock()
	if _, exists := rl.LimitedClients["client1"]; !exists {
		t.Errorf("Client should exist before cleanup")
	}
	rl.Mu.Unlock()

	// Wait for expiry
	time.Sleep(20 * time.Millisecond)

	// Run cleanup
	rl.Cleanup()

	// Verify client was removed
	rl.Mu.Lock()
	if _, exists := rl.LimitedClients["client1"]; exists {
		t.Errorf("Client should have been removed during cleanup")
	}
	rl.Mu.Unlock()

	tests := []struct {
		name string
		rl   *RateLimiter
	}{
		{
			name: "Cleanup expired clients",
			rl:   rl,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.rl.Cleanup()
		})
	}
}

// TestParallelRateLimiters tests multiple rate limiters running in parallel
func TestParallelRateLimiters(t *testing.T) {
	apiRl := NewRateLimit("api-parallel", 2, 2, time.Minute)
	dbRl := NewRateLimit("db-parallel", 3, 3, time.Minute)
	cacheRl := NewRateLimit("cache-parallel", 5, 5, time.Minute)

	var wg sync.WaitGroup

	numClients := 5
	requestsPerClient := 10

	apiLimited := int32(0)
	dbLimited := int32(0)
	cacheLimited := int32(0)

	for i := range numClients {
		clientID := fmt.Sprintf("client-%d", i)
		wg.Add(1)

		go func(clientID string) {
			defer wg.Done()

			for range requestsPerClient {
				if apiRl.Limit(clientID) {
					atomic.AddInt32(&apiLimited, 1)
				}

				if dbRl.Limit(clientID) {
					atomic.AddInt32(&dbLimited, 1)
				}

				if cacheRl.Limit(clientID) {
					atomic.AddInt32(&cacheLimited, 1)
				}

				time.Sleep(time.Millisecond)
			}
		}(clientID)
	}

	wg.Wait()

	t.Logf("API limited requests: %d", apiLimited)
	t.Logf("DB limited requests: %d", dbLimited)
	t.Logf("Cache limited requests: %d", cacheLimited)

	if apiLimited <= dbLimited {
		t.Errorf("Expected API limiter (%d) to limit more requests than DB limiter (%d)", apiLimited, dbLimited)
	}

	if dbLimited <= cacheLimited {
		t.Errorf("Expected DB limiter (%d) to limit more requests than Cache limiter (%d)", dbLimited, cacheLimited)
	}

	if cacheLimited == 0 {
		t.Errorf("Expected Cache limiter to limit some requests, but it limited 0")
	}
}

func runLimit(i int, rl *RateLimiter, wg *sync.WaitGroup) {
	wg.Add(2)

	go func(id int) {
		defer wg.Done()
		rl.Limit(fmt.Sprintf("client-%d", id%10))
	}(i)

	go func() {
		defer wg.Done()
		rl.Cleanup()
	}()

}

func TestRateLimiterRace(t *testing.T) {
	rl := NewRateLimit("race-test", 10, 10, time.Minute)

	concurrency := 100
	var wg sync.WaitGroup

	for i := range 5 {
		rl.Limit(fmt.Sprintf("client-%d", i))
	}

	for i := range concurrency {
		runLimit(i, rl, &wg)
	}

	wg.Wait()

	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	if len(rl.LimitedClients) > 10 {
		t.Errorf("Expected at most 10 clients, got %d", len(rl.LimitedClients))
	}
}

func BenchmarkRateLimiterRace(b *testing.B) {
	rl := NewRateLimit("race-benchmark", 10, 10, time.Minute)

	concurrency := 100
	var wg sync.WaitGroup

	for i := range 5 {
		rl.Limit(fmt.Sprintf("client-%d", i))
	}

	reset(b)
	for i := range b.N / concurrency {
		runLimit(i, rl, &wg)
	}

	wg.Wait()
}
