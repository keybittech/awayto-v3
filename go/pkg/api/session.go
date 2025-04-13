package api

import (
	"errors"
	"net"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.UserSession)

type SessionMux struct {
	mux *http.ServeMux
}

func NewSessionMux() *SessionMux {
	return &SessionMux{
		mux: http.NewServeMux(),
	}
}

func (sm *SessionMux) Handle(pattern string, handler SessionHandler) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var deferredErr error
		var didLimit bool
		userSub := "unknown client"

		startTime, exeTimeDefer := util.ExeTime(pattern)

		defer func() {
			// if deferredErr, auth token is bad so return unauth
			if deferredErr != nil {
				util.ErrorLog.Println(deferredErr)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			if !didLimit {
				exeTimeDefer(startTime, "sub:"+userSub)
			}
		}()

		token, ok := req.Header["Authorization"]
		if !ok {
			// If we can't get auth token from header, use ip to rate limit
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if apiRl.Limit(ip) {
				didLimit = true
				WriteLimit(w)
				return
			}
			deferredErr = util.ErrCheck(errors.New("no auth token"))
			return
		}

		// validate provided token to return a user struct
		session, err := ValidateToken(token[0], req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
		if err != nil {
			deferredErr = util.ErrCheck(err)
			return
		}

		userSub = session.UserSub

		// rate limit authenticated user
		if apiRl.Limit(session.UserSub) {
			didLimit = true
			WriteLimit(w)
			return
		}

		handler(w, req, session)
	})

	sm.mux.Handle(pattern, handlerFunc)
}

func (sm *SessionMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	sm.mux.HandleFunc(pattern, handler)
}

func (sm *SessionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sm.mux.ServeHTTP(w, r)
}
