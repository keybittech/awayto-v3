package api

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func ValidateToken(publicKey *rsa.PublicKey, token, timezone, anonIp string) (*types.UserSession, error) {
	token = strings.TrimPrefix(token, "Bearer ")

	parsedToken, err := jwt.ParseWithClaims(token, &clients.KeycloakUserWithClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("bad signing method")
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if !parsedToken.Valid {
		return nil, util.ErrCheck(errors.New("invalid token during parse"))
	}

	claims, ok := parsedToken.Claims.(*clients.KeycloakUserWithClaims)
	if !ok {
		return nil, util.ErrCheck(errors.New("could not parse claims"))
	}
	var roleBits types.SiteRoles
	if clientAccess, clientAccessOk := claims.ResourceAccess[claims.Azp]; clientAccessOk {
		roleBits = util.StringsToSiteRoles(clientAccess.Roles)
	}

	session := &types.UserSession{
		UserSub:       claims.Subject,
		UserEmail:     claims.Email,
		SubGroupPaths: claims.Groups,
		RoleBits:      roleBits,
		ExpiresAt:     claims.ExpiresAt,
		Timezone:      timezone,
		AnonIp:        anonIp,
	}

	return session, nil
}

func SetForwardingHeadersAndServe(prox *httputil.ReverseProxy, w http.ResponseWriter, r *http.Request) {
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	r.Header.Add("X-Forwarded-Proto", "https")
	r.Header.Add("X-Forwarded-Host", r.Host)
	prox.ServeHTTP(w, r)
}

func (a *API) InitAuthProxy() {
	kcRealm := util.E_KC_REALM
	kcInternal, err := url.Parse(util.E_KC_INTERNAL)
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

	a.Server.Handler.(*http.ServeMux).Handle("/auth/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authMux.ServeHTTP(w, r)
	}))

	a.Server.Handler.(*http.ServeMux).Handle("/auth/resources/", http.StripPrefix("/auth", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		SetForwardingHeadersAndServe(authProxy, w, req)
	})))

	loginStatusHandler := func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
		var cookieVal string
		var cookieExpires int

		if strings.HasSuffix(req.URL.Path, "login") {
			if !util.CookieExpired(req) {
				return
			}

			cookieVal, err = util.WriteSigned(util.LOGIN_SIGNATURE_NAME, strconv.FormatInt(session.GetExpiresAt(), 10))
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			cookieExpires = 24
		} else {
			cookieExpires = -24
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "valid_signature",
			Value:    cookieVal,
			Path:     "/",
			Expires:  time.Now().Add(time.Duration(cookieExpires) * time.Hour),
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
			HttpOnly: true,
		})
	}

	a.Server.Handler.(*http.ServeMux).Handle("/login", a.ValidateTokenMiddleware()(loginStatusHandler))
	a.Server.Handler.(*http.ServeMux).Handle("/logout", a.ValidateTokenMiddleware()(loginStatusHandler))
}
