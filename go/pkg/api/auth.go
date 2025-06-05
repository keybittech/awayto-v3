package api

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (a *API) InitAuthProxy() {
	kcRealm := util.E_KC_REALM
	kcInternal, err := url.Parse(util.E_KC_INTERNAL)
	if err != nil {
		log.Fatal("invalid keycloak url")
	}

	authMux := http.NewServeMux()
	authProxy := httputil.NewSingleHostReverseProxy(kcInternal)

	authProxy.ModifyResponse = func(resp *http.Response) error {
		if shouldAddNonce(resp.Request.URL.Path) &&
			strings.Contains(resp.Header.Get("Content-Type"), "text/html") {

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body.Close()

			if nonce, ok := resp.Request.Context().Value("CSP-Nonce").([]byte); ok {
				body = addNonceToHTML(body, nonce)
			}

			resp.Body = io.NopCloser(bytes.NewReader(body))
			resp.ContentLength = int64(len(body))
			resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
		}
		return nil
	}

	userRoutes := []string{
		"login-actions/registration",
		"login-actions/authenticate",
		"login-actions/reset-credentials",
		"protocol/openid-connect/registrations",
		"protocol/openid-connect/auth",
		"protocol/openid-connect/logout",
	}

	for _, ur := range userRoutes {
		authRoute := fmt.Sprintf("/auth/realms/%s/%s", kcRealm, ur)
		a.Server.Handler.(*http.ServeMux).Handle(authRoute, http.StripPrefix("/auth",
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

	a.Server.Handler.(*http.ServeMux).Handle("/auth/register", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		redirectParams := a.Handlers.GenerateLoginOrRegisterParams(req)
		if redirectParams == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		http.Redirect(w, req, util.E_KC_OPENID_REGISTER_URL+redirectParams, http.StatusFound)
	}))

	// auth/login prepares the user's challenge codes, and sends the user to login
	a.Server.Handler.(*http.ServeMux).Handle("/auth/login", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		redirectParams := a.Handlers.GenerateLoginOrRegisterParams(req)
		if redirectParams == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		http.Redirect(w, req, util.E_KC_OPENID_AUTH_URL+redirectParams, http.StatusFound)
	}))

	// The browser checks /auth/status, sees the user is not logged in, then forwards to /auth/login
	a.Server.Handler.(*http.ServeMux).Handle("/auth/status", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		session, err := a.Handlers.GetSession(req)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		checkedSession := a.Handlers.CheckSessionExpiry(req, session)
		if checkedSession == nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"authenticated": true,
		})
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
		codeVerifier := tempSession.GetCodeVerifier()

		session, err := util.GetValidTokenChallenge(req, code, codeVerifier, tempSession.GetUa(), tempSession.GetTz(), util.AnonIp(req.RemoteAddr))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		a.Handlers.StoreSession(req.Context(), session)

		signedSessionId, err := util.WriteSigned("session_id", session.Id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    signedSessionId,
			Path:     "/",
			MaxAge:   int(time.Until(time.Unix(0, session.GetRefreshExpiresAt())).Seconds()),
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
			util.ErrorLog.Println(errors.New("no sessionId during logout"))

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

func shouldAddNonce(path string) bool {
	noncePages := []string{
		"login-actions/authenticate",
		"login-actions/registration",
		"login-actions/reset-credentials",
		"protocol/openid-connect/auth",
		"protocol/openid-connect/registrations",
	}

	for _, page := range noncePages {
		if strings.Contains(path, page) {
			return true
		}
	}
	return false
}

func addNonceToHTML(body []byte, nonce []byte) []byte {
	modifiedBody := regexp.MustCompile(`(?i)<script\b`).ReplaceAll(body, []byte(`<script nonce="`+string(nonce)+`"`))
	modifiedBody = regexp.MustCompile(`\s*onsubmit="[^"]*"`).ReplaceAll(modifiedBody, []byte(""))
	scriptPattern := regexp.MustCompile(`(localStorage\.clear\(\);)(\s*loginForm\.submit\(\);)`)
	modifiedBody = scriptPattern.ReplaceAll(modifiedBody, []byte(`${1}
			document.getElementById('kc-login').disabled = true;${2}`))
	return modifiedBody
}
