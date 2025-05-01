package api

import (
	"crypto/rsa"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

var tokenCache = NewTokenCache()

func CleanupApiTokenCache() {
	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()
	tokenCache.sessions = make(map[string]*types.UserSession)
}

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.UserSession)

type SessionMux struct {
	publicKey *rsa.PublicKey
	mux       *http.ServeMux
}

func NewSessionMux(pk *rsa.PublicKey) *SessionMux {
	return &SessionMux{
		publicKey: pk,
		mux:       http.NewServeMux(),
	}
}

func (sm *SessionMux) Handle(pattern string, handler SessionHandler) {
	sm.mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		auth, ok := req.Header["Authorization"]
		if !ok {
			return
		}

		if apiRl.Limit(auth[0]) {
			WriteLimit(w)
			return
		}

		session, ok := tokenCache.Get(auth[0])
		if !ok {
			var err error
			session, err = ValidateToken(sm.publicKey, auth[0], req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
			if err != nil {
				return
			}
			tokenCache.Set(auth[0], session)
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
