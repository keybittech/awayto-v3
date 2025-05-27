package api

import (
	json "encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

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
		// "login-actions/reset-credentials",
		// "protocol/openid-connect/3p-cookies/step1.html",
		// "protocol/openid-connect/3p-cookies/step2.html",
		// "protocol/openid-connect/login-status-iframe.html",
		// "protocol/openid-connect/login-status-iframe.html/init",
		// "protocol/openid-connect/registrations",
		"protocol/openid-connect/auth",
		"protocol/openid-connect/token",
		// "protocol/openid-connect/logout",
	}

	for _, ur := range userRoutes {
		authRoute := fmt.Sprintf("/auth/realms/%s/%s", kcRealm, ur)
		authMux.Handle(authRoute, http.StripPrefix("/auth",
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				util.SetForwardingHeaders(req)
				authProxy.ServeHTTP(w, req)
			}),
		))
	}

	a.Server.Handler.(*http.ServeMux).Handle("/auth/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authMux.ServeHTTP(w, r)
	}))

	a.Server.Handler.(*http.ServeMux).Handle("/auth/resources/", http.StripPrefix("/auth", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		util.SetForwardingHeaders(req)
		authProxy.ServeHTTP(w, req)
	})))

	// loginStatusHandler := func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
	// 	var cookieVal string
	// 	var cookieExpires int
	//
	//
	// 	if strings.HasSuffix(req.URL.Path, "login") {
	// 		if !util.CookieExpired(req) {
	// 			return
	// 		}
	//
	// 		cookieVal, err = util.WriteSigned(util.LOGIN_SIGNATURE_NAME, strconv.FormatInt(session.GetExpiresAt(), 10))
	// 		if err != nil {
	// 			util.ErrorLog.Println(util.ErrCheck(err))
	// 			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	// 			return
	// 		}
	//
	// 		cookieExpires = 24
	// 	} else {
	// 		cookieExpires = -24
	// 	}
	//
	// 	http.SetCookie(w, &http.Cookie{
	// 		Name:     "valid_signature",
	// 		Value:    cookieVal,
	// 		Path:     "/",
	// 		Expires:  time.Now().Add(time.Duration(cookieExpires) * time.Hour),
	// 		SameSite: http.SameSiteStrictMode,
	// 		Secure:   true,
	// 		HttpOnly: true,
	// 	})
	// }
	//
	// a.Server.Handler.(*http.ServeMux).Handle("/login", a.ValidateTokenMiddleware()(loginStatusHandler))
	// a.Server.Handler.(*http.ServeMux).Handle("/logout", a.ValidateTokenMiddleware()(loginStatusHandler))

	a.Server.Handler.(*http.ServeMux).Handle("/auth/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := util.GetSessionIdFromCookie(r)
		if sessionId == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		session, ok := a.Cache.UserSessions.Get(sessionId)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if time.Now().After(time.Unix(0, session.GetExpiresAt()).Add(-30 * time.Second)) {
			if err := a.Cache.RefreshAccessToken(r); err != nil {
				util.ErrorLog.Printf("Token refresh failed: %v", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"authenticated": true,
		})
	}))

	a.Server.Handler.(*http.ServeMux).Handle("/auth/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		codeVerifier := util.GenerateCodeVerifier()
		codeChallenge := util.GenerateCodeChallenge(codeVerifier)
		state := util.GenerateState()

		tempSession := types.NewConcurrentTempAuthSession(&types.TempAuthSession{
			CodeVerifier: codeVerifier,
			State:        state,
			CreatedAt:    time.Now().UnixNano(),
		})

		a.Cache.TempAuthSessions.Store(state, tempSession)

		params := url.Values{
			"response_type":  {"code"},
			"client_id":      {util.E_KC_USER_CLIENT},
			"redirect_uri":   {util.E_APP_HOST_URL + "/auth/callback"},
			"scope":          {"openid profile email groups"},
			"state":          {state},
			"code_challenge": {codeChallenge},
			// "code_challenge_method": {"S256"},
		}

		redirectURL := util.E_KC_OPENID_AUTH_URL + "?" + params.Encode()
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}))

	a.Server.Handler.(*http.ServeMux).Handle("/auth/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" || state == "" {
			http.Error(w, "Missing code or state", http.StatusBadRequest)
			return
		}

		tempSession := a.Cache.TempAuthSessions.Load(state)
		a.Cache.TempAuthSessions.Delete(state)

		data := url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {util.E_KC_USER_CLIENT},
			"client_secret": {util.E_KC_USER_CLIENT_SECRET},
			"redirect_uri":  {util.E_APP_HOST_URL + "/auth/callback"},
			"code":          {code},
			"code_verifier": {tempSession.GetCodeVerifier()},
		}

		util.SetForwardingHeaders(r)

		resp, err := util.PostFormData(util.E_KC_OPENID_TOKEN_URL, r.Header, data)
		if err != nil {
			util.ErrorLog.Printf("Token exchange failed post: %v", err)
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		var tokens types.OIDCToken
		if err := protojson.Unmarshal(resp, &tokens); err != nil {
			util.ErrorLog.Printf("Token exchange failed token decode: %v", err)
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		sessionId := util.GenerateSessionId()

		err = a.Cache.ValidateToken(&tokens, sessionId, r.Header.Get("X-Tz"), util.AnonIp(r.RemoteAddr))
		if err != nil {
			util.ErrorLog.Printf("Token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionId,
			Path:     "/",
			MaxAge:   int(tokens.RefreshExpiresIn),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		http.Redirect(w, r, "/app/", http.StatusFound)
	}))

	a.Server.Handler.(*http.ServeMux).Handle("/auth/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := util.GetSessionIdFromCookie(r)
		if sessionId != "" {
			a.Cache.UserSessions.Delete(sessionId)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		w.WriteHeader(http.StatusOK)
	}))

}
