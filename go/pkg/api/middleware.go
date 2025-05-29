package api

import (
	"bytes"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.ConcurrentUserSession)

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

func (a *API) ValidateSessionMiddleware() func(next SessionHandler) http.HandlerFunc { // This converts a regular HandlerFunc into a SessionHandler
	return func(next SessionHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			session := a.Cache.GetSessionFromCookie(req)
			if session == nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			if time.Now().After(time.Unix(0, session.GetAccessExpiresAt()).Add(-30 * time.Second)) {
				var err error
				session, err = a.Handlers.RefreshSession(req)
				if err != nil {
					http.Error(w, "Token refresh failed in middleware", http.StatusUnauthorized)
					return
				}
			}

			next(w, req, session)
		}
	}
}

func (a *API) GroupInfoMiddleware(next SessionHandler) SessionHandler {
	return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
		sessionSubGroups := session.GetSubGroupPaths()
		if len(sessionSubGroups) == 0 {
			next(w, req, session)
			return
		}

		sessionGroupId := session.GetGroupId()
		refresh := sessionGroupId == ""

		if !refresh {
			currentGroupVersion, found := a.Cache.GroupSessionVersions.Load(sessionGroupId)
			if !found || currentGroupVersion > session.GetGroupSessionVersion() {
				refresh = true
			}
		}

		if refresh {
			ctx := req.Context()

			subGroupPath := sessionSubGroups[0]

			concurrentCachedGroup, concurrentCachedSubGroup, err := a.Handlers.CheckGroupInfo(ctx, subGroupPath)
			if err != nil {
				util.ErrorLog.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			groupId := concurrentCachedGroup.GetId()

			session.SetGroupId(groupId)
			session.SetGroupPath(concurrentCachedGroup.GetPath())
			session.SetGroupCode(concurrentCachedGroup.GetCode())
			session.SetGroupName(concurrentCachedGroup.GetName())
			session.SetGroupExternalId(concurrentCachedGroup.GetExternalId())
			session.SetGroupSub(concurrentCachedGroup.GetSub())
			session.SetGroupAi(concurrentCachedGroup.GetAi())

			session.SetSubGroupPath(subGroupPath)
			session.SetSubGroupName(concurrentCachedSubGroup.GetName())
			session.SetSubGroupExternalId(concurrentCachedSubGroup.GetExternalId())

			currentGroupVersion, found := a.Cache.GroupSessionVersions.Load(groupId)
			if !found {
				session.SetGroupSessionVersion(0)
			} else {
				session.SetGroupSessionVersion(currentGroupVersion)
			}
		}

		next(w, req, session)
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
		} else {
			return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
				if session.GetRoleBits()&siteRole == 0 {
					util.WriteAuthRequest(req, session.GetUserSub(), opts.SiteRoleName)
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

			cacheKey := userSub + req.URL.String()
			cacheKeyModTime := cacheKey + clients.CacheKeySuffixModTime

			// Check redis cache for request
			cachedResponse := a.Handlers.Redis.RedisClient.MGet(ctx, cacheKey, cacheKeyModTime).Val()

			cachedData, dataOk := cachedResponse[0].(string)
			modTime, modOk := cachedResponse[1].(string)

			// cachedData will be {} if empty
			if dataOk && modOk && len(cachedData) > 2 && modTime != "" {
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
					pipe.SetEx(ctx, cacheKey, buf.Bytes(), duration)
					pipe.SetEx(ctx, cacheKeyModTime, modTimeStr, duration)
				} else {
					pipe.Set(ctx, cacheKey, buf.Bytes(), 0)
					pipe.Set(ctx, cacheKeyModTime, modTimeStr, 0)
				}

				_, err := pipe.Exec(ctx)
				if err != nil {
					util.ErrorLog.Println("failed to perform cache insert pipeline", err.Error(), cacheKey)
				}
			}
		}
	}
}
