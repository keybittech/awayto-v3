package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/crypto"
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
		csp.WriteString("script-src 'nonce-" + nonce + "' 'strict-dynamic' 'wasm-unsafe-eval'; ")
		csp.WriteString("style-src-elem 'self' 'nonce-" + nonce + "'; ")
		csp.WriteString("style-src-attr 'unsafe-inline'; ")
		csp.WriteString("img-src 'self' data:; ")
		csp.WriteString("font-src 'self' data:; ")
		csp.WriteString("connect-src 'self' blob:; ")
		csp.WriteString("media-src 'self' blob:; ")
		csp.WriteString("object-src 'none'; ")
		csp.WriteString("child-src 'self' blob:; ")
		csp.WriteString("worker-src 'self' blob:; ")
		csp.WriteString("frame-ancestors 'none'; ")
		csp.WriteString("form-action 'self'; ")
		csp.WriteString("base-uri 'self'; ")
		csp.WriteString("upgrade-insecure-requests; ")

		cs := csp.String()

		w.Header().Set("Content-Security-Policy", cs)

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
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

type ctxKey string

const CtxVaultKey ctxKey = "vaultKeyProp"

type VaultResponseWriter struct {
	http.ResponseWriter
	buf         *bytes.Buffer
	statusCode  int
	paramHeader string
}

func (vrw *VaultResponseWriter) WriteHeader(code int) {
	vrw.statusCode = code
}

func (vrw *VaultResponseWriter) Write(b []byte) (int, error) {
	return vrw.buf.Write(b)
}

func (a *API) VaultMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(req.URL.Path, "/api") || strings.Contains(req.URL.Path, "/vault/key") {
			next.ServeHTTP(w, req)
			return
		}

		sessionId, err := util.GetSessionIdFromCookie(req)
		if sessionId == "" || err != nil {
			util.ErrorLog.Println("VaultMiddleware no session id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var sharedSecret []byte

		ct := req.Header.Get("Content-Type")

		// If content type is vault, handle body
		if ct == "application/x-awayto-vault" && !strings.HasPrefix(ct, "multipart/form-data") {
			reqBytes, readErr := io.ReadAll(req.Body)
			req.Body.Close()
			if readErr == nil && len(reqBytes) > 0 {
				var plaintext []byte
				plaintext, sharedSecret, err = crypto.ServerDecrypt(crypto.VaultKey, reqBytes, sessionId)

				if err == nil {
					req.Body = io.NopCloser(bytes.NewBuffer(plaintext))
					req.Header.Set("Content-Type", "application/json")
					if oct := req.Header.Get("X-Original-Content-Type"); oct != "" {
						req.Header.Set("Content-Type", oct)
					}

					req.ContentLength = int64(len(plaintext))
					req.Header.Set("Content-Length", strconv.Itoa(len(plaintext)))
				} else {
					err = util.ErrCheck(err)
				}
			}
		}

		// Else this is a GET so handle header
		if len(sharedSecret) == 0 {
			if vaultHeader := req.Header.Get("X-Awayto-Vault"); vaultHeader != "" {
				blob, b64Err := base64.StdEncoding.DecodeString(vaultHeader)
				if b64Err == nil {
					_, ss, dErr := crypto.ServerDecrypt(crypto.VaultKey, blob, sessionId)
					if dErr == nil {
						sharedSecret = ss
					} else {
						err = util.ErrCheck(dErr)
					}
				}
			}
		}

		if err != nil {
			util.ErrorLog.Printf("VaultMiddleware: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if len(sharedSecret) == 0 {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ctx := context.WithValue(req.Context(), CtxVaultKey, sharedSecret)
		req = req.WithContext(ctx)

		vrw := &VaultResponseWriter{
			ResponseWriter: w,
			buf:            &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(vrw, req)

		respPlaintext := vrw.buf.Bytes()

		encryptedResp, encryptErr := crypto.ServerEncrypt(sharedSecret, respPlaintext, sessionId)
		if encryptErr != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		b64Resp := make([]byte, base64.StdEncoding.EncodedLen(len(encryptedResp)))
		base64.StdEncoding.Encode(b64Resp, encryptedResp)

		w.Header().Set("Content-Type", "application/x-awayto-vault")
		w.Header().Set("Content-Length", strconv.Itoa(len(b64Resp)))

		w.WriteHeader(vrw.statusCode)
		w.Write(b64Resp)
	})
}

// Converts a regular HandlerFunc into a SessionHandler
func (a *API) ValidateSessionMiddleware() func(next SessionHandler) http.HandlerFunc {
	return func(next SessionHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			session, err := a.Handlers.GetSession(req)
			if session == nil || err != nil {
				util.ErrorLog.Printf("validate middleware get session fail, err: %v", err)

				// Clear invalid session cookie
				util.SetSessionCookie(w, -1, "")

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
				util.DebugLog.Printf("scored too high, err: %v", scorings)
				if err := a.Handlers.DeleteSession(req.Context(), session.GetId()); err != nil {
					util.DebugLog.Printf("could not delete high score session, err: %v", scorings)
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

			signedSessionId, err := util.WriteSigned("session_id", checkedSession.GetId())
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			util.SetSessionCookie(w, int64(time.Until(time.Unix(0, checkedSession.GetRefreshExpiresAt())).Seconds()), signedSessionId)

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
				util.WriteAuthRequest(req, session.GetUserSub(), string(siteRole))
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next(w, req, session)
		}
	}
}

func genETag(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

type BufResponseWriter struct {
	http.ResponseWriter
	Buf        *bytes.Buffer
	StatusCode int
}

func (bw *BufResponseWriter) Write(data []byte) (int, error) {
	return bw.Buf.Write(data)
}

func (bw *BufResponseWriter) WriteHeader(code int) {
	bw.StatusCode = code
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
				etag := genETag(cachedBytes)

				if req.Header.Get("If-None-Match") == etag {
					w.WriteHeader(http.StatusNotModified)
					return
				}

				w.Header().Set("ETag", etag)
				w.Header().Set("Content-Length", strconv.Itoa(len(cachedBytes)))

				w.WriteHeader(http.StatusOK)
				w.Write(cachedBytes)
				return
			}

			// Run the handler
			var buf bytes.Buffer
			bufWriter := &BufResponseWriter{
				ResponseWriter: w,
				Buf:            &buf,
				StatusCode:     http.StatusOK,
			}

			next(bufWriter, req, session)

			// Handle empty responses or errors
			if buf.Len() == 0 {
				if bufWriter.StatusCode != 200 {
					w.WriteHeader(bufWriter.StatusCode)
				}
				return
			}

			// Check if matching client
			responseBytes := buf.Bytes()
			newETag := genETag(responseBytes)
			w.Header().Set("ETag", newETag)

			if req.Header.Get("If-None-Match") == newETag {
				w.WriteHeader(http.StatusNotModified)
				return
			} else {
				w.WriteHeader(bufWriter.StatusCode)
				w.Write(responseBytes)
			}

			// Store in cache
			if buf.Len() > 0 {
				duration := duration180s

				if hasDuration {
					duration = parsedDuration
				} else if shouldStore {
					duration = duration86400s
				}

				_, err := a.Handlers.Redis.RedisClient.SetEx(ctx, cacheKey, responseBytes, duration).Result()
				if err != nil {
					util.ErrorLog.Println("failed to perform cache insert pipeline", err.Error(), cacheKey)
				}
			}
		}
	}
}
