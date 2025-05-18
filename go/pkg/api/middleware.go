package api

import (
	"bytes"
	"errors"
	"fmt"
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
	return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
		hasGroups := len(session.GetSubGroups()) > 0

		if hasGroups { // user has groups
			sgPath := session.GetSubGroups()[0]
			session.SetSubGroupPath(sgPath)
			session.SetSubGroupName(sgPath[strings.LastIndex(sgPath, "/")+1:])
		} else if session.GetGroupSessionVersion() > 0 { // user had a group but no more
			session.SetSubGroupName("")
			session.SetSubGroupPath("")
			session.SetGroupName("")
			session.SetGroupPath("")
			session.SetRoleName("")
			session.SetGroupExternalId("")
			session.SetSubGroupExternalId("")
			session.SetGroupId("")
			session.SetGroupAi(false)
			session.SetGroupSub("")
			session.SetGroupSessionVersion(0)
			// a.Cache.SessionTokens.Store(req.Header.Get("Authorization"), session) // not necessary with concurrent session
			next(w, req, session)
			return
		}

		if hasGroups && (session.GetGroupId() == "" || a.Cache.GroupSessionVersions.Load(session.GetGroupId()) != session.GetGroupSessionVersion()) {
			sgPath := session.GetSubGroupPath()
			subGroup, ok := a.Cache.SubGroups.Load(sgPath)
			if !ok {
				util.ErrorLog.Println(errors.New("could not load subgroup " + sgPath))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			groupPath := session.GetGroupPath()
			group, ok := a.Cache.Groups.Load(groupPath)
			if !ok {
				util.ErrorLog.Println(errors.New("could not load group " + groupPath))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			session.SetGroupId(group.GetId())
			session.SetGroupExternalId(group.GetExternalId())
			session.SetGroupSub(group.GetSub())
			session.SetGroupName(group.GetName())
			session.SetGroupAi(group.GetAi())
			session.SetGroupPath(subGroup.GetGroupPath())
			session.SetRoleName(subGroup.GetName())
			session.SetSubGroupExternalId(subGroup.GetExternalId())
			session.SetGroupSessionVersion(a.Cache.GroupSessionVersions.Load(group.GetId()))

			// a.Handlers.Cache.SetSessionToken(req.Header.Get("Authorization"), session) // not necessary with concurrent session
		}

		next(w, req, session)
	}
}

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	return func(next SessionHandler) SessionHandler {
		if opts.SiteRole == int64(types.SiteRoles_UNRESTRICTED) {
			return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
				next(w, req, session)
			}
		} else {
			return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
				if session.GetRoleBits()&opts.SiteRole == 0 {
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
		return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
			if opts.ShouldSkip {
				next(w, req, session)
				return
			}

			ctx := req.Context()

			// Any non-GET processed normally, and deletes cache key unless being stored
			if !opts.ShouldStore && req.Method != http.MethodGet {
				next(w, req, session)

				iLen := len(opts.Invalidations)
				if iLen == 0 {
					return
				}

				scanKeys := make([]string, iLen)

				for _, val := range opts.Invalidations {
					invCacheKey := session.GetUserSub() + val

					if strings.Index(val, "*") == -1 {
						scanKeys = append(scanKeys, invCacheKey)
						scanKeys = append(scanKeys, invCacheKey+cacheKeySuffixModTime)
						continue
					}

					var cursor uint64
					var err error

					for {
						var pathKeys []string
						pathKeys, cursor, err = a.Handlers.Redis.RedisClient.Scan(ctx, 0, invCacheKey, 10).Result()
						if err != nil {
							util.ErrorLog.Println(util.ErrCheck(err))
							break
						}

						scanKeys = append(scanKeys, pathKeys...)

						if cursor == 0 {
							break
						}
					}
				}

				if len(scanKeys) > 0 {
					deleted, err := a.Handlers.Redis.RedisClient.Del(ctx, scanKeys...).Result()
					if err != nil {
						util.ErrorLog.Println("failed to perform cache mutation cleanup pipeline", err.Error(), fmt.Sprint(opts.Invalidations))
					}

					println("DID DELETE CACHE RECORDS", deleted, fmt.Sprint(scanKeys))
				}

				return
			}

			cacheKey := session.GetUserSub() + req.URL.String()
			cacheKeyModTime := cacheKey + cacheKeySuffixModTime

			// Check redis cache for request
			cachedResponse := a.Handlers.Redis.RedisClient.MGet(ctx, cacheKey, cacheKeyModTime).Val()

			cachedData, dataOk := cachedResponse[0].(string)
			modTime, modOk := cachedResponse[1].(string)

			// cachedData will be {} if empty
			if dataOk && modOk && len(cachedData) > 2 && modTime != "" {
				println("USING cache for ", cacheKey)

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
				println("STORING cache for ", cacheKey)

				duration := duration180

				if hasDuration {
					duration = parsedDuration
				} else if opts.ShouldStore {
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
