package api

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type responseCodeWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseCodeWriter {
	return &responseCodeWriter{w, http.StatusOK}
}

func (rw *responseCodeWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (a *API) AccessRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, req)
		util.WriteAccessRequest(req, time.Since(start).Milliseconds(), rw.statusCode)
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

func (a *API) ValidateTokenMiddleware() func(next SessionHandler) http.HandlerFunc {
	return func(next SessionHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			token := req.Header.Get("Authorization")
			if token == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			session, err := ValidateToken(a.Handlers.Keycloak.Client.PublicKey, token, req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
			if err != nil {
				util.ErrorLog.Println(err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next(w, req, session)
		}
	}
}

func (a *API) GroupInfoMiddleware(next SessionHandler) SessionHandler {
	return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
		hasGroups := len(session.SubGroups) > 0

		if hasGroups { // user has groups
			session.SubGroupPath = session.SubGroups[0]
			session.SubGroupName = session.SubGroups[0][strings.LastIndex(session.SubGroups[0], "/")+1:]
		} else if session.GroupSessionVersion > 0 { // user had a group but no more
			session.SubGroupName = ""
			session.SubGroupPath = ""
			session.GroupName = ""
			session.GroupPath = ""
			session.RoleName = ""
			session.GroupExternalId = ""
			session.SubGroupExternalId = ""
			session.GroupId = ""
			session.GroupAi = false
			session.GroupSub = ""
			session.GroupSessionVersion = 0
			a.Handlers.Cache.SetSessionToken(req.Header.Get("Authorization"), session)
			next(w, req, session)
			return
		}

		if hasGroups && (session.GroupId == "" || a.Handlers.Cache.GetGroupSessionVersion(session.GroupId) != session.GroupSessionVersion) {
			subGroup := a.Handlers.Cache.GetCachedSubGroup(session.SubGroupPath)
			if subGroup == nil {
				util.ErrorLog.Println(errors.New("could not load subgroup " + session.SubGroupPath))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			group := a.Handlers.Cache.GetCachedGroup(subGroup.GroupPath)
			if group == nil {
				util.ErrorLog.Println(errors.New("could not load group " + subGroup.GroupPath))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			session.GroupId = group.Id
			session.GroupExternalId = group.ExternalId
			session.GroupSub = group.Sub
			session.GroupName = group.Name
			session.GroupAi = group.Ai
			session.GroupPath = subGroup.GroupPath
			session.RoleName = subGroup.Name
			session.SubGroupExternalId = subGroup.ExternalId
			session.GroupSessionVersion = a.Handlers.Cache.GetGroupSessionVersion(session.GroupId)

			a.Handlers.Cache.SetSessionToken(req.Header.Get("Authorization"), session)
			// go func(token string, s *types.UserSession) {
			// 	s.GroupSessionVersion = a.Handlers.Cache.GetGroupSessionVersion(session.GroupId)
			// 	a.Handlers.Cache.SetSessionToken(token, s)
			// }(req.Header.Get("Authorization"), session)
		}

		next(w, req, session)
	}
}

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	return func(next SessionHandler) SessionHandler {
		if opts.SiteRole == int64(types.SiteRoles_UNRESTRICTED) {
			return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
				next(w, req, session)
			}
		} else {
			return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
				if session.RoleBits&opts.SiteRole == 0 {
					util.WriteAuthRequest(req, session.UserSub, opts.SiteRoleName)
					w.WriteHeader(http.StatusForbidden)
					return
				}

				next(w, req, session)
			}
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
	shouldStore := int64(types.CacheType_STORE) == opts.CacheType
	shouldSkip := int64(types.CacheType_SKIP) == opts.CacheType

	var parsedDuration time.Duration
	var hasDuration bool
	if opts.CacheDuration > 0 {
		hasDuration = true
		cacheDuration := strconv.FormatInt(opts.CacheDuration, 10)
		if d, err := time.ParseDuration(cacheDuration + "s"); err == nil {
			parsedDuration = d
		} else {
			log.Fatalf("incorrect cache duration parsing %s", cacheDuration+"s")
		}
	}

	return func(next SessionHandler) SessionHandler {
		return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
			if shouldSkip {
				next(w, req, session)
				return
			}

			ctx := req.Context()
			var cacheKey strings.Builder
			cacheKey.WriteString(session.UserSub)
			cacheKey.WriteString(req.URL.String())
			cacheKeyData := cacheKey.String() + cacheKeySuffixData
			cacheKeyModTime := cacheKey.String() + cacheKeySuffixModTime

			// Any non-GET processed normally, and deletes cache key unless being stored
			if !shouldStore && req.Method != http.MethodGet {
				next(w, req, session)
				if !shouldStore {
					pipe := a.Handlers.Redis.RedisClient.Pipeline()
					pipe.Del(ctx, cacheKeyData)
					pipe.Del(ctx, cacheKeyModTime)
					_, err := pipe.Exec(ctx)
					if err != nil {
						util.ErrorLog.Println("failed to perform cache mutation cleanup pipeline", err.Error(), cacheKeyData)
					}
				}
				return
			}

			// Check redis cache for request
			cachedResponse := a.Handlers.Redis.RedisClient.MGet(ctx, cacheKeyData, cacheKeyModTime).Val()

			cachedData, dataOk := cachedResponse[0].(string)
			modTime, modOk := cachedResponse[1].(string)

			if dataOk && modOk && len(cachedData) > 0 && modTime != "" {

				lastMod, err := time.Parse(time.RFC3339Nano, modTime)
				if err == nil {

					w.Header().Set("Last-Modified", lastMod.Format(http.TimeFormat))

					// Check if client sent If-Modified-Since header
					if ifModifiedSince := req.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
						if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
							if !lastMod.Truncate(time.Second).After(t.Truncate(time.Second)) {
								w.Header().Set("X-Cache-Status", "UNMODIFIED")
								w.WriteHeader(http.StatusNotModified)
								return
							}
						}
					}

					// Serve cached data if no header interaction
					w.Header().Set("X-Cache-Status", "HIT")
					w.Header().Set("Content-Length", strconv.Itoa(len(cachedData)))
					_, err := w.Write([]byte(cachedData))
					if err != nil {
						util.ErrorLog.Println(util.ErrCheck(err))
					}
					return
				}
			}

			// No cached data, create it on this request
			timeNow := time.Now()

			w.Header().Set("X-Cache-Status", "MISS")
			w.Header().Set("Last-Modified", timeNow.Format(http.TimeFormat))

			var buf bytes.Buffer

			// Perform the handler request
			cacheWriter := &CacheWriter{
				ResponseWriter: w,
				Buffer:         &buf,
			}

			// Response is written out to client
			next(cacheWriter, req, session)

			// Cache any response
			if buf.Len() > 0 {
				duration := duration180

				if hasDuration {
					duration = parsedDuration
				} else if shouldStore {
					duration = 0
				}

				pipe := a.Handlers.Redis.RedisClient.Pipeline()
				modTimeStr := timeNow.Format(time.RFC3339Nano)

				if duration > 0 {
					pipe.SetEx(ctx, cacheKeyData, buf.Bytes(), duration)
					pipe.SetEx(ctx, cacheKeyModTime, modTimeStr, duration)
				} else {
					pipe.Set(ctx, cacheKeyData, buf.Bytes(), 0)
					pipe.Set(ctx, cacheKeyModTime, modTimeStr, 0)
				}

				_, err := pipe.Exec(ctx)
				if err != nil {
					util.ErrorLog.Println("failed to perform cache insert pipeline", err.Error(), cacheKeyData)
				}
			}
		}
	}
}
