package api

import (
	"bytes"
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

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *types.ConcurrentUserSession)

func (a *API) AccessRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		if rw == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
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

func (a *API) ValidateTokenMiddleware() func(next SessionHandler) http.HandlerFunc { // This converts a regular HandlerFunc into a SessionHandler
	return func(next SessionHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			token := req.Header.Get("Authorization")
			if token == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			session, err := a.Cache.UserSessions.LoadOrSet(token, func() (*types.UserSession, error) {
				return ValidateToken(a.Handlers.Keycloak.Client.PublicKey, token, req.Header.Get("X-Tz"), util.AnonIp(req.RemoteAddr))
			})
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
			subGroupPathIdx := strings.LastIndex(subGroupPath, "/")
			subGroupName := subGroupPath[subGroupPathIdx+1:]

			groupPath := subGroupPath[:subGroupPathIdx]
			groupName := groupPath[1:]

			concurrentCachedGroup, err := a.Cache.Groups.LoadOrSet(groupPath, func() (*types.CachedGroup, error) {
				batch := util.NewBatchable(a.Handlers.Database.DatabaseClient.Pool, "worker", "", 0)
				groupReq := util.BatchQueryRow[types.IGroup](batch, `
					SELECT id, code, "externalId", sub, ai
					FROM dbview_schema.enabled_groups
					WHERE name = $1
				`, groupName)
				batch.Send(ctx)

				group := *groupReq

				kcSubGroups, err := a.Handlers.Keycloak.GetGroupSubGroups(ctx, "worker", group.ExternalId)
				if err != nil {
					return nil, err
				}

				cachedGroup := &types.CachedGroup{
					SubGroupPaths: make([]string, 0, len(kcSubGroups)),
					Id:            group.Id,
					Code:          group.Code,
					Name:          groupName,
					ExternalId:    group.ExternalId,
					Sub:           group.Sub,
					Ai:            group.Ai,
				}

				for _, subGroup := range kcSubGroups {
					cachedGroup.SubGroupPaths = append(cachedGroup.SubGroupPaths, subGroup.Path)
				}

				if _, found := a.Cache.GroupSessionVersions.Load(group.Id); !found {
					a.Cache.GroupSessionVersions.Store(group.Id, time.Now().UnixNano())
				}

				return cachedGroup, nil
			})
			if err != nil {
				util.ErrorLog.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			concurrentCachedSubGroup, err := a.Cache.SubGroups.LoadOrSet(subGroupPath, func() (*types.CachedSubGroup, error) {
				batch := util.NewBatchable(a.Handlers.Database.DatabaseClient.Pool, "worker", "", 0)
				subGroupReq := util.BatchQueryRow[types.IGroupRole](batch, `
					SELECT egr."externalId"
					FROM dbview_schema.enabled_roles er
					JOIN dbview_schema.enabled_group_roles egr ON egr."roleId" = er.id
					WHERE er.name = $1
				`, subGroupName)

				batch.Send(ctx)

				subGroup := *subGroupReq

				cachedSubGroup := &types.CachedSubGroup{
					ExternalId: subGroup.ExternalId,
					Name:       subGroupName,
					GroupPath:  groupPath,
				}

				return cachedSubGroup, nil
			})
			if err != nil {
				util.ErrorLog.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			groupId := concurrentCachedGroup.GetId()

			session.SetGroupId(groupId)
			session.SetGroupPath(groupPath)
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

			a.Cache.UserSessions.Store(req.Header.Get("Authorization"), session)
		}

		next(w, req, session)
	}
}

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	optPack := opts.Unpack()
	return func(next SessionHandler) SessionHandler {
		if optPack.SiteRole == types.SiteRoles_UNRESTRICTED {
			return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
				next(w, req, session)
			}
		} else {
			return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
				if session.GetRoleBits()&optPack.SiteRole == 0 {
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

				scanKeys := make([]string, 0, iLen*2) // double num to make room for ModTime keys

				for _, val := range opts.Invalidations {
					invCacheKey := userSub + val

					// If it's not a dynamically generated path, we don't need to scan redis for keys
					if strings.Index(val, "*") == -1 {
						scanKeys = append(scanKeys, invCacheKey)
						scanKeys = append(scanKeys, invCacheKey+cacheKeySuffixModTime)
						continue
					}

					var cursor uint64

					for {
						var pathKeys []string
						var err error
						pathKeys, cursor, err = a.Handlers.Redis.RedisClient.Scan(ctx, cursor, invCacheKey, 10).Result()
						if err != nil {
							util.ErrorLog.Println(util.ErrCheck(err))
							break
						}

						for _, pathKey := range pathKeys {
							scanKeys = append(scanKeys, pathKey)
							scanKeys = append(scanKeys, pathKey+cacheKeySuffixModTime)
						}

						if cursor == 0 {
							break
						}
					}
				}

				if len(scanKeys) > 0 {
					_, err := a.Handlers.Redis.RedisClient.Del(ctx, scanKeys...).Result()
					if err != nil {
						util.ErrorLog.Println("failed to perform cache mutation cleanup pipeline", err.Error(), fmt.Sprint(opts.Invalidations))
					}
				}

				return
			}

			cacheKey := userSub + req.URL.String()
			cacheKeyModTime := cacheKey + cacheKeySuffixModTime

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
