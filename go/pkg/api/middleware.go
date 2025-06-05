package api

import (
	"bytes"
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.ConcurrentUserSession)

func (a *API) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nonce := util.GenerateState()

		ctx := context.WithValue(req.Context(), "CSP-Nonce", []byte(nonce))
		req = req.WithContext(ctx)

		var csp strings.Builder

		csp.WriteString("default-src 'self'; ")
		csp.WriteString("script-src 'nonce-" + nonce + "' 'strict-dynamic'; ")
		csp.WriteString("style-src 'self' 'unsafe-inline'; ")
		csp.WriteString("img-src 'self' data: https:; ")
		csp.WriteString("font-src 'self' data:; ")
		csp.WriteString("connect-src 'self' wss: https: stun: turn:; ")
		csp.WriteString("media-src 'self' blob:; ")
		csp.WriteString("object-src 'none'; ")
		csp.WriteString("child-src 'self' blob:; ")
		csp.WriteString("worker-src 'self' blob:; ")
		csp.WriteString("frame-ancestors 'none'; ")
		csp.WriteString("form-action 'self'; ")
		csp.WriteString("base-uri 'self'; ")
		csp.WriteString("upgrade-insecure-requests;")

		cs := csp.String()

		w.Header().Set("Content-Security-Policy", cs)
		w.Header().Set("X-Content-Security-Policy", cs)

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(self), camera=(self)")

		if req.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		w.Header().Set("Server", "")

		next.ServeHTTP(w, req)
	})
}

func (a *API) AccessRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, req)
		if !rw.hijacked {
			util.WriteAccessRequest(req, time.Since(start).Milliseconds(), rw.statusCode)
		} else {
			util.WriteAccessRequest(req, time.Since(start).Milliseconds(), http.StatusSwitchingProtocols)
		}
	})
}

func (a *API) LimitMiddleware(rl *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if rl.Limit(ip) {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// Converts a regular HandlerFunc into a SessionHandler
func (a *API) ValidateSessionMiddleware() func(next SessionHandler) http.HandlerFunc {
	return func(next SessionHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			session, err := a.Handlers.GetSession(req)
			if session == nil || err != nil {
				util.ErrorLog.Printf("validate middleware get session fail, err: %v", err)

				// Clear invalid session cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				})

				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			scorings := [][]string{
				{session.GetUserAgent(), req.Header.Get("User-Agent")},
				{session.GetTimezone(), req.Header.Get("X-Tz")},
				{session.GetAnonIp(), util.AnonIp(req.RemoteAddr)},
			}

			score := util.ScoreValues(scorings)
			if score > 1 {
				util.DebugLog.Printf("failed score, err: %v", scorings)
				if err := a.Handlers.DeleteSession(req.Context(), session.GetId()); err != nil {
					util.DebugLog.Printf("failed score, err: %v", scorings)
				}
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			checkedSession := a.Handlers.CheckSessionExpiry(req, session)
			if checkedSession == nil {
				util.DebugLog.Printf("failed score, err: %v", scorings)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next(w, req, checkedSession)
		}
	}
}

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	siteRole := opts.Unpack().SiteRole
	unrestricted := int32(types.SiteRoles_UNRESTRICTED)
	return func(next SessionHandler) SessionHandler {
		if siteRole == unrestricted {
			return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
				next(w, req, session)
			}
		}

		return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
			if session.GetRoleBits()&siteRole == 0 {
				util.WriteAuthRequest(req, session.GetUserSub(), opts.SiteRoleName)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next(w, req, session)
		}
	}
}

type CacheWriter struct {
	http.ResponseWriter
	Buffer *bytes.Buffer
}

// This lets us pull data out of the writer and into the cache, after whatever handler has written to it
func (cw *CacheWriter) Write(data []byte) (int, error) {
	defer cw.Buffer.Write(data)
	return cw.ResponseWriter.Write(data)
}

func (a *API) CacheMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	shouldStore := opts.Unpack().ShouldStore
	shouldSkip := opts.Unpack().ShouldSkip

	var parsedDuration time.Duration
	var hasDuration bool
	if opts.Unpack().CacheDuration > 0 {
		hasDuration = true
		cacheDuration := strconv.FormatInt(int64(opts.Unpack().CacheDuration), 10)
		if d, err := time.ParseDuration(cacheDuration + "s"); err == nil {
			parsedDuration = d
		} else {
			log.Fatalf("incorrect cache duration parsing %s", cacheDuration+"s")
		}
	}

	return func(next SessionHandler) SessionHandler {
		return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
			if shouldSkip {
				next(w, req, session)
				return
			}

			ctx := req.Context()
			userSub := session.GetUserSub()

			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, private")
			w.Header().Set("Vary", "Cookie, User-Agent, X-Tz")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")

			// Any non-GET processed normally, and deletes cache key unless being stored
			if !shouldStore && req.Method != http.MethodGet {
				next(w, req, session)

				iLen := len(opts.Invalidations)
				if iLen == 0 {
					return
				}

				a.Handlers.Redis.ScanAndDelKeys(ctx, opts.Invalidations, userSub)
				return
			}

			// Serve from cache if possible
			cacheKey := userSub + req.URL.String()
			cachedBytes, err := a.Handlers.Redis.RedisClient.Get(ctx, cacheKey).Bytes()
			if err == nil {
				w.Header().Set("Content-Length", strconv.Itoa(len(cachedBytes)))
				_, err := w.Write(cachedBytes)
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
				}

				w.WriteHeader(http.StatusNotModified)
				return
			}

			// Add to cache
			var buf bytes.Buffer
			cacheWriter := &CacheWriter{
				ResponseWriter: w,
				Buffer:         &buf,
			}

			next(cacheWriter, req, session)

			if buf.Len() > 0 {
				duration := duration180

				if hasDuration {
					duration = parsedDuration
				} else if shouldStore {
					duration = 86400
				}

				_, err := a.Handlers.Redis.RedisClient.SetEx(ctx, cacheKey, buf.Bytes(), duration).Result()
				if err != nil {
					util.ErrorLog.Println("failed to perform cache insert pipeline", err.Error(), cacheKey)
				}
			}
		}
	}
}
