package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func SetForwardingHeadersAndServe(prox *httputil.ReverseProxy, w http.ResponseWriter, r *http.Request) {
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	r.Header.Add("X-Forwarded-Proto", "https")
	r.Header.Add("X-Forwarded-Host", r.Host)
	prox.ServeHTTP(w, r)
}

func (a *API) InitAuthProxy(mux *http.ServeMux) {
	kcRealm := os.Getenv("KC_REALM")
	kcInternal, err := url.Parse(os.Getenv("KC_INTERNAL"))
	if err != nil {
		log.Fatal("invalid keycloak url")
	}

	authMux := http.NewServeMux()
	authProxy := httputil.NewSingleHostReverseProxy(kcInternal)

	userRoutes := []string{
		"login-actions/registration",
		"login-actions/authenticate",
		"login-actions/reset-credentials",
		"protocol/openid-connect/3p-cookies/step1.html",
		"protocol/openid-connect/3p-cookies/step2.html",
		"protocol/openid-connect/login-status-iframe.html",
		"protocol/openid-connect/login-status-iframe.html/init",
		"protocol/openid-connect/registrations",
		"protocol/openid-connect/auth",
		"protocol/openid-connect/token",
		"protocol/openid-connect/logout",
	}

	for _, ur := range userRoutes {
		authRoute := fmt.Sprintf("/auth/realms/%s/%s", kcRealm, ur)
		authMux.Handle(authRoute, http.StripPrefix("/auth",
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				SetForwardingHeadersAndServe(authProxy, w, req)
			}),
		))
	}

	mux.Handle("/auth/", a.LimitMiddleware(10, 10)(func(w http.ResponseWriter, r *http.Request) {
		authMux.ServeHTTP(w, r)
	}))

	mux.Handle("/auth/resources/", http.StripPrefix("/auth",
		a.LimitMiddleware(10, 20)(
			func(w http.ResponseWriter, req *http.Request) {
				SetForwardingHeadersAndServe(authProxy, w, req)
			},
		),
	))

	mux.Handle("/login",
		a.ValidateTokenMiddleware(1, 1)(
			func(w http.ResponseWriter, r *http.Request, session *clients.UserSession) {
				http.SetCookie(w, &http.Cookie{
					Name:     "valid_signature",
					Value:    util.WriteSigned(util.LOGIN_SIGNATURE_NAME, fmt.Sprint(session.ExpiresAt)),
					Path:     "/",
					Expires:  time.Now().Add(24 * time.Hour),
					SameSite: http.SameSiteStrictMode,
					Secure:   true,
					HttpOnly: true,
				})
			},
		),
	)

	mux.Handle("/logout",
		a.ValidateTokenMiddleware(1, 1)(
			func(w http.ResponseWriter, r *http.Request, session *clients.UserSession) {
				http.SetCookie(w, &http.Cookie{
					Name:     "valid_signature",
					Value:    "",
					Path:     "/",
					Expires:  time.Now().Add(-24 * time.Hour),
					SameSite: http.SameSiteStrictMode,
					Secure:   true,
					HttpOnly: true,
				})
			},
		),
	)
}
