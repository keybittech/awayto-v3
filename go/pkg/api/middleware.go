package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"

	"golang.org/x/time/rate"
)

func (a *API) LimitMiddleware(limit rate.Limit, burst int) func(next http.HandlerFunc) http.HandlerFunc {

	rl := NewRateLimit("limit-middleware", limit, burst, time.Duration(5*time.Minute))

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			if rl.Limit(ip) {
				WriteLimit(w)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("APP_HOST_URL"))
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-TZ")
			w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH")

			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, req)
		}
	}
}

func (a *API) ValidateTokenMiddleware(limit rate.Limit, burst int) func(next SessionHandler) http.HandlerFunc {
	rl := NewRateLimit("validate-token", limit, burst, time.Duration(5*time.Minute))

	return func(next SessionHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			var deferredErr error

			defer func() {
				if deferredErr != nil {
					ip, _, err := net.SplitHostPort(req.RemoteAddr)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						util.ErrorLog.Println(util.ErrCheck(err))
					}

					if rl.Limit(ip) {
						WriteLimit(w)
						return
					}

					util.ErrorLog.Println(deferredErr)
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				}
			}()

			token, ok := req.Header["Authorization"]
			if !ok {
				deferredErr = util.ErrCheck(errors.New("no auth token"))
				return
			}

			session, err := ValidateToken(a.Handlers.Keycloak.Client.PublicKey, token[0], req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
			if err != nil {
				deferredErr = util.ErrCheck(err)
				return
			}

			if rl.Limit(session.UserSub) {
				WriteLimit(w)
				return
			}

			next(w, req, session)
		}
	}
}

func (a *API) GroupInfoMiddleware(next SessionHandler) SessionHandler {
	return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
		ctx := req.Context()
		needsLookup := false

		if len(session.SubGroups) > 0 {
			session.SubGroupName = session.SubGroups[0]
		} else if session.GroupSessionVersion > 0 {
			session.SubGroupName = ""
			session.GroupName = ""
			session.RoleName = ""
			session.GroupExternalId = ""
			session.SubGroupExternalId = ""
			session.GroupId = ""
			session.GroupAi = false
			session.GroupSub = ""
			session.GroupSessionVersion = 0
			next(w, req, session)
			return
		}

		if session.GroupId == "" {
			needsLookup = true
			session.GroupId = ""
			session.GroupAi = false
			session.GroupSub = ""
			session.GroupExternalId = ""
			session.SubGroupExternalId = ""
			session.GroupSessionVersion = 0
		} else {
			groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(ctx, session.GroupId)
			if err != nil && !errors.Is(err, redis.Nil) {
				util.ErrorLog.Println(util.ErrCheck(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if groupVersion != session.GroupSessionVersion {
				needsLookup = true
			}
		}

		if needsLookup && session.SubGroupName != "" {
			splitIdx := strings.LastIndex(session.SubGroupName, "/")
			if len(session.SubGroupName) < splitIdx {
				util.ErrorLog.Println(errors.New("bad subgroup name length"))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			session.GroupName = session.SubGroupName[1:splitIdx]
			session.RoleName = session.SubGroupName[splitIdx+1:]
			kcGroupName := session.SubGroupName[:splitIdx]

			kcGroups, err := a.Handlers.Keycloak.GetGroupByName(session.UserSub, kcGroupName)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			for _, gr := range kcGroups {
				if gr.Path == kcGroupName {
					session.GroupExternalId = gr.Id
					break
				}
			}

			kcSubgroups, err := a.Handlers.Keycloak.GetGroupSubgroups(session.UserSub, session.GroupExternalId)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			for _, gr := range kcSubgroups {
				if gr.Path == session.SubGroupName {
					session.SubGroupExternalId = gr.Id
					break
				}
			}

			if session.GroupExternalId == "" || session.SubGroupExternalId == "" {
				util.ErrorLog.Println(util.ErrCheck(errors.New("could not describe group or subgroup external ids" + session.GroupExternalId + "==" + session.SubGroupExternalId)))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ds := clients.DbSession{
				Pool:        a.Handlers.Database.DatabaseClient.Pool,
				UserSession: session,
			}

			row, done, err := ds.SessionBatchQueryRow(req.Context(), `
				SELECT id, ai, sub
				FROM dbtable_schema.groups
				WHERE external_id = $1
			`, session.GroupExternalId)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = row.Scan(&session.GroupId, &session.GroupAi, &session.GroupSub)
			if err != nil {
				done()
				util.ErrorLog.Println(util.ErrCheck(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			done()

			groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(req.Context(), session.GroupId)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			session.GroupSessionVersion = groupVersion

			sessionJson, err := protojson.Marshal(session)
			if err != nil {
				util.ErrCheck(err)
			} else {
				redisTokenKey := redisTokenPrefix + req.Header.Get("Authorization")
				err = a.Handlers.Redis.RedisClient.SetEx(ctx, redisTokenKey, sessionJson, redisTokenDuration).Err()
				if err != nil {
					util.ErrCheck(err)
				}
			}
		}

		next(w, req, session)
	}
}

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	siteRoleName := types.SiteRoles_name[opts.SiteRole]
	return func(next SessionHandler) SessionHandler {

		if opts.SiteRole == int32(types.SiteRoles_UNRESTRICTED) {
			return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
				next(w, req, session)
			}
		} else {
			return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
				var sb strings.Builder
				sb.WriteString(opts.Pattern)
				sb.WriteString(" sub:")
				sb.WriteString(session.UserSub)
				sb.WriteString(" role:")
				sb.WriteString(siteRoleName)

				if session.RoleBits&opts.SiteRole == 0 {
					sb.WriteString(" FAIL")
					util.AuthLog.Println(sb.String())
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}

				sb.WriteString(" PASS")
				util.AuthLog.Println(sb.String())

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

type CacheMeta struct {
	Data       []byte    `json:"data"`
	LastMod    time.Time `json:"last_modified"`
	StatusCode int       `json:"status_code"`
}

const maxCacheBuffer = 1 << 12

var (
	apiPathLen     = len(os.Getenv("API_PATH"))
	duration180, _ = time.ParseDuration("180s")
)

var cacheMiddlewareBufferPool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		return &buf
	},
}

func (a *API) CacheMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	shouldStore := types.CacheType_STORE == opts.CacheType

	return func(next SessionHandler) SessionHandler {
		return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
			// gives a cache key like absd-asff-asff-asfdgroup/users
			var cacheKey strings.Builder
			cacheKey.WriteString(session.UserSub)
			cacheKey.WriteString(req.URL.String()[apiPathLen:])

			// Any non-GET processed normally, and deletes cache key unless being stored
			if !shouldStore && req.Method != http.MethodGet {
				next(w, req, session)
				if types.CacheType_STORE != opts.CacheType {
					a.Handlers.Redis.Client().Del(req.Context(), cacheKey.String())
				}
				return
			}

			// Check if client sent If-Modified-Since header
			var clientModTime time.Time
			if req.Header.Get("If-Modified-Since") != "" {
				if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil {
					clientModTime = t.Truncate(time.Second)
				}
			}

			if types.CacheType_SKIP != opts.CacheType {
				// Check redis cache for request
				if cacheData, err := a.Handlers.Redis.Client().Get(req.Context(), cacheKey.String()).Bytes(); err == nil {

					var cacheMeta CacheMeta
					err = json.Unmarshal(cacheData, &cacheMeta)
					if err != nil {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						util.ErrorLog.Println(util.ErrCheck(err))
						return
					}

					w.Header().Set("Last-Modified", cacheMeta.LastMod.UTC().Format(http.TimeFormat))

					// If the client has current data
					if !clientModTime.IsZero() && !cacheMeta.LastMod.Truncate(time.Second).After(clientModTime) {
						w.Header().Set("X-Cache-Status", "UNMODIFIED")
						w.WriteHeader(http.StatusNotModified)
						w.Write([]byte{})
						return
					}

					// Serve cached data if no header interaction
					w.Header().Set("X-Cache-Status", "HIT")
					w.Write(cacheMeta.Data)
					return
				}
			}

			// No cached data, create it on this request
			timeNow := time.Now().UTC()

			w.Header().Set("X-Cache-Status", "MISS")
			w.Header().Set("Last-Modified", timeNow.Format(http.TimeFormat))

			buf := cacheMiddlewareBufferPool.Get().(*bytes.Buffer)
			buf.Reset()

			// Perform the handler request
			cacheWriter := &CacheWriter{
				ResponseWriter: w,
				Buffer:         buf,
			}

			// Response is written out to client
			next(cacheWriter, req, session)

			writeLen := cacheWriter.Buffer.Len()

			// Cache any response
			if writeLen > 0 {

				// Prep for redis storage
				cacheMeta := CacheMeta{
					Data:    cacheWriter.Buffer.Bytes(),
					LastMod: timeNow,
				}

				cacheData, err := json.Marshal(cacheMeta)
				if err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					util.ErrorLog.Println(util.ErrCheck(err))
					return
				}

				if shouldStore {

					// Store the response in redis until restart
					a.Handlers.Redis.Client().Set(req.Context(), cacheKey.String(), cacheData, 0)

				} else {

					// Default 3 min cache or as otherwise specified
					duration := duration180
					if opts.CacheDuration > 0 {
						if parsedDuration, err := time.ParseDuration(fmt.Sprintf("%ds", opts.CacheDuration)); err == nil {
							duration = parsedDuration
						} else {
							util.ErrorLog.Println(util.ErrCheck(err))
						}
					}

					a.Handlers.Redis.Client().SetEx(req.Context(), cacheKey.String(), cacheData, duration)
				}
			}

			if writeLen < maxCacheBuffer {
				cacheMiddlewareBufferPool.Put(buf)
			}
		}
	}
}

// func (a *API) SocketAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
// 		ctx := req.Context()
//
// 		var userSub, userEmail string
//
// 		ticket := req.URL.Query().Get("ticket")
// 		if ticket == "" {
// 			util.ErrorLog.Println(errors.New("no ticket during socket auth"))
// 			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
// 			return
// 		}
//
// 		subscriber, err := a.Handlers.Socket.GetSubscriberByTicket(ticket)
// 		if err != nil {
// 			util.ErrorLog.Println(util.ErrCheck(err))
// 			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
// 			return
// 		}
//
// 		if subscriber.UserSub == "" {
// 			util.ErrorLog.Println(util.ErrCheck(errors.New("err getting user sub during socket auth")))
// 			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
// 			return
// 		}
//
// 		kcUser, err := a.Handlers.Keycloak.GetUserInfoById(subscriber.UserSub)
// 		if err != nil {
// 			util.ErrorLog.Println(util.ErrCheck(err))
// 			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
// 			return
// 		}
//
// 		userSub = kcUser.Sub
// 		userEmail = kcUser.Email
//
// 		if userSub == "" || userEmail == "" {
// 			util.ErrorLog.Println(errors.New("change these errors"))
// 			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
// 			return
// 		}
//
// 		ctx = context.WithValue(ctx, "UserSession", &clients.UserSession{UserSub: userSub, UserEmail: userEmail})
// 		ctx = context.WithValue(ctx, "SourceIp", util.AnonIp(req.RemoteAddr))
//
// 		req = req.WithContext(ctx)
// 		next.ServeHTTP(w, req)
// 	})
// }
