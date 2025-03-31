package api

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
)

// Adapted from https://blog.logrocket.com/rate-limiting-go-application

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *clients.UserSession)

type SessionMux struct {
	mux *http.ServeMux
}

func NewSessionMux() *SessionMux {
	return &SessionMux{
		mux: http.NewServeMux(),
	}
}

func validateToken(token string) (*clients.KeycloakUser, error) {
	if strings.Contains(token, "Bearer") {
		token = strings.Split(token, " ")[1]
	}

	parsedToken, err := jwt.ParseWithClaims(token, &clients.KeycloakUser{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("bad signing method")
		}
		return clients.KeycloakPublicKey, nil
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if !parsedToken.Valid {
		return nil, util.ErrCheck(errors.New("invalid token during parse"))
	}

	if claims, ok := parsedToken.Claims.(*clients.KeycloakUser); ok {
		return claims, nil
	}

	return nil, nil
}

var sessionHandlerLimit = 2
var sessionHandlerBurst = 20

func (sm *SessionMux) Handle(pattern string, handler SessionHandler) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var deferredErr error

		// if deferred, auth token is bad so return unauth
		defer func() {
			if deferredErr != nil {
				util.ErrorLog.Println(deferredErr)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
		}()

		token, ok := req.Header["Authorization"]
		if !ok {

			// If we can't get auth token from header, use ip to rate limit
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			limited := Limiter(apiLimitMu, apiLimited, rate.Limit(sessionHandlerLimit), sessionHandlerBurst, ip)
			if limited {
				WriteLimit(w)
				return
			}
			deferredErr = util.ErrCheck(errors.New("no auth token"))
			return
		}

		// validate provided token to return a user struct
		kcUser, err := validateToken(token[0])
		if err != nil {
			deferredErr = util.ErrCheck(err)
			return
		}

		// rate limit authenticated user
		limited := Limiter(apiLimitMu, apiLimited, rate.Limit(sessionHandlerLimit), sessionHandlerBurst, kcUser.Sub)
		if limited {
			WriteLimit(w)
			return
		}

		session := &clients.UserSession{
			UserSub:                 kcUser.Sub,
			UserEmail:               kcUser.Email,
			SubGroups:               kcUser.Groups,
			AvailableUserGroupRoles: kcUser.ResourceAccess[kcUser.Azp].Roles,
			Timezone:                req.Header.Get("X-TZ"),
			ExpiresAt:               kcUser.ExpiresAt,
			AnonIp:                  util.AnonIp(req.RemoteAddr),
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
