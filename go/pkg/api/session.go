package api

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	redisTokenPrefix   = "session_tokens:"
	redisTokenDuration = 55 * time.Second
)

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.UserSession)

type SessionMux struct {
	publicKey *rsa.PublicKey
	mux       *http.ServeMux
	redis     *redis.Client
	rl        *RateLimiter
}

func NewSessionMux(pk *rsa.PublicKey, redis *redis.Client, rl *RateLimiter) *SessionMux {
	return &SessionMux{
		publicKey: pk,
		mux:       http.NewServeMux(),
		redis:     redis,
		rl:        rl,
	}
}

func (sm *SessionMux) Handle(pattern string, handler SessionHandler) {
	sm.mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		auth, ok := req.Header["Authorization"]
		if !ok || len(auth) == 0 || auth[0] == "" {
			return
		}

		if sm.rl.Limit(auth[0]) {
			WriteLimit(w)
			return
		}

		ctx := req.Context()
		redisTokenKey := redisTokenPrefix + auth[0]

		session := &types.UserSession{}
		sessionBytes, err := sm.redis.Get(ctx, redisTokenKey).Bytes()
		if err == nil {
			err = protojson.Unmarshal(sessionBytes, session)
			if err == nil {
				handler(w, req, session)
				return
			}
		} else if !errors.Is(err, redis.Nil) {
			util.ErrCheck(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session, err = ValidateToken(sm.publicKey, auth[0], req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		sessionJson, err := protojson.Marshal(session)
		if err != nil {
			util.ErrCheck(err)
		} else {
			err = sm.redis.SetEx(ctx, redisTokenKey, sessionJson, redisTokenDuration).Err()
			if err != nil {
				util.ErrCheck(err)
			}
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
