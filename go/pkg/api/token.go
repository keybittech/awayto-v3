package api

import (
	"sync"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

type TokenCache struct {
	sessions map[string]*types.UserSession
	mu       sync.RWMutex
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		sessions: make(map[string]*types.UserSession),
	}
}

func (tc *TokenCache) Set(token string, session *types.UserSession) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.sessions[token] = session
}

func (tc *TokenCache) Get(token string) (*types.UserSession, bool) {
	tc.mu.RLock()
	session, ok := tc.sessions[token]
	tc.mu.RUnlock()
	return session, ok
}
