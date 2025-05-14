package main

import (
	"context"
	"fmt"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/api"
)

func setupGc(a *api.API, stopChan chan struct{}) {
	generalCleanupTicker := time.NewTicker(5 * time.Minute)
	connLen := 0
	for {
		select {
		case <-generalCleanupTicker.C:
			api.RateLimiters.Range(func(_, value any) bool {
				rl := value.(*api.RateLimiter)
				rl.Cleanup()
				return true
			})

			// can be removed when no longer needed
			ctx := context.Background()
			defer ctx.Done()
			socketConnections, err := a.Handlers.Redis.RedisClient.SMembers(ctx, "socket_server_connections").Result()
			sockLen := len(socketConnections)
			if err != nil {
				println("reading socket connection err", err.Error())
			}
			if connLen != sockLen {
				connLen = sockLen
				fmt.Printf("got socket connection list new count :%d %+v\n", len(socketConnections), socketConnections)
			}
		case <-stopChan:
			return
		}
	}
}
