package api

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const (
	redisTokenPrefix   = "session_tokens:"
	redisTokenDuration = 55 * time.Second
)

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.UserSession)

type SessionMux struct {
	publicKey    *rsa.PublicKey
	mux          *http.ServeMux
	handlerCache *handlers.HandlerCache
}

func NewSessionMux(pk *rsa.PublicKey, handlerCache *handlers.HandlerCache) *SessionMux {
	return &SessionMux{
		publicKey:    pk,
		mux:          http.NewServeMux(),
		handlerCache: handlerCache,
	}
}

func (sm *SessionMux) Handle(pattern string, handler SessionHandler) {
	sm.mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if cachedSession := sm.handlerCache.GetSessionToken(token); cachedSession != nil {
			handler(w, req, cachedSession)
			return
		}

		session, err := ValidateToken(sm.publicKey, token, req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		go sm.handlerCache.SetSessionToken(token, session)

		handler(w, req, session)
	}))
}

func (sm *SessionMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	sm.mux.HandleFunc(pattern, handler)
}

func (sm *SessionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sm.mux.ServeHTTP(w, r)
}
