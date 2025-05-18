package api

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const (
	redisTokenPrefix   = "session_tokens:"
	redisTokenDuration = 55 * time.Second
)

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.ConcurrentUserSession)

type SessionMux struct {
	publicKey    *rsa.PublicKey
	mux          *http.ServeMux
	userSessions *types.ConcurrentUserSessionCache
}

func NewSessionMux(pk *rsa.PublicKey, userSessions *types.ConcurrentUserSessionCache) *SessionMux {
	return &SessionMux{
		publicKey:    pk,
		mux:          http.NewServeMux(),
		userSessions: userSessions,
	}
}

func (sm *SessionMux) Handle(pattern string, handler SessionHandler) {
	sm.mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		session, err := sm.userSessions.LoadOrSet(token, func() (*types.UserSession, error) {
			return ValidateToken(sm.publicKey, token, req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
		})
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		handler(w, req, session)
	}))
}

func (sm *SessionMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	sm.mux.HandleFunc(pattern, handler)
}

func (sm *SessionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sm.mux.ServeHTTP(w, r)
}
