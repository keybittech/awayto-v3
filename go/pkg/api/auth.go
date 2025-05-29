package api

import (
	json "encoding/json"
	"errors"
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
		"protocol/openid-connect/registrations",
		"protocol/openid-connect/auth",
		"protocol/openid-connect/token",
		"protocol/openid-connect/logout",
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

	// The browser checks /auth/status, sees the user is not logged in, then forwards to /auth/login
	a.Server.Handler.(*http.ServeMux).Handle("/auth/status", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sessionId := util.GetSessionIdFromCookie(req)
		if sessionId == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		session, ok := a.Cache.UserSessions.Get(sessionId)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if time.Now().After(time.Unix(0, session.GetAccessExpiresAt()).Add(-30 * time.Second)) {
			if _, err := a.Handlers.RefreshSession(req); err != nil {
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

	// auth/login prepares the user's challenge codes, and sends the user to login
	a.Server.Handler.(*http.ServeMux).Handle("/auth/login", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		codeVerifier := util.GenerateCodeVerifier()
		codeChallenge := util.GenerateCodeChallenge(codeVerifier)
		state := util.GenerateState()

		tempSession := types.NewConcurrentTempAuthSession(&types.TempAuthSession{
			CodeVerifier: codeVerifier,
			State:        state,
			CreatedAt:    time.Now().UnixNano(),
			Tz:           req.URL.Query().Get("tz"),
			Ua:           req.Header.Get("User-Agent"),
		})

		a.Cache.TempAuthSessions.Store(state, tempSession)

		params := url.Values{
			"response_type":         {"code"},
			"client_id":             {util.E_KC_USER_CLIENT},
			"redirect_uri":          {util.E_APP_HOST_URL + "/auth/callback"},
			"scope":                 {"openid profile email groups"},
			"state":                 {state},
			"code_challenge":        {codeChallenge},
			"code_challenge_method": {"S256"},
		}

		redirectURL := util.E_KC_OPENID_AUTH_URL + "?" + params.Encode()
		http.Redirect(w, req, redirectURL, http.StatusFound)
	}))

	// After logging in the user's code is verified and token validated
	a.Server.Handler.(*http.ServeMux).Handle("/auth/callback", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		state := req.URL.Query().Get("state")

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

		util.SetForwardingHeaders(req)

		resp, err := util.PostFormData(util.E_KC_OPENID_TOKEN_URL, req.Header, data)
		if err != nil {
			util.ErrorLog.Printf("Token exchange failed post: %v", err)
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		tokens := &types.OIDCToken{}
		if err := protojson.Unmarshal(resp, tokens); err != nil {
			util.ErrorLog.Printf("Token exchange failed token decode: %v", err)
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		session, err := util.ValidateToken(tokens, util.GetUA(tempSession.GetUa()), tempSession.GetTz(), util.AnonIp(req.RemoteAddr))
		if err != nil {
			util.ErrorLog.Printf("Token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		a.Handlers.StoreSession(req.Context(), session)

		signedSessionId, err := util.WriteSigned("session_id", session.Id)
		if err != nil {
			util.ErrorLog.Printf("Failed to sign session ID: %v", err)
			http.Error(w, "Session creation failed", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    signedSessionId,
			Path:     "/",
			MaxAge:   int(tokens.RefreshExpiresIn),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, req, "/app/", http.StatusFound)
	}))

	a.Server.Handler.(*http.ServeMux).Handle("/auth/logout", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		sessionId := util.GetSessionIdFromCookie(req)
		if sessionId == "" {
			util.ErrorLog.Println(errors.New("no sessionid during logout"))

			return
		}

		session, ok := a.Handlers.Cache.UserSessions.Get(sessionId)
		if !ok {
			util.ErrorLog.Println(errors.New("no cached session during logout"))
			http.Redirect(w, req, "/", http.StatusOK)
			return
		}

		if err := a.Handlers.DeleteSession(req.Context(), sessionId); err != nil {
			util.ErrorLog.Println(errors.New("could not delete session during logout"))
			return
		}

		params := url.Values{
			"client_id":                {util.E_KC_USER_CLIENT},
			"post_logout_redirect_uri": {util.E_APP_HOST_URL},
			"id_token_hint":            {session.GetIdToken()},
		}

		redirectURL := util.E_KC_OPENID_LOGOUT_URL + "?" + params.Encode()

		http.Redirect(w, req, redirectURL, http.StatusFound)
	}))
}
