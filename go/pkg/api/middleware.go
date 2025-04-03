package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"golang.org/x/time/rate"
)

func (a *API) LimitMiddleware(limit rate.Limit, burst int) func(next http.HandlerFunc) http.HandlerFunc {

	mu, limitedClients := NewRateLimit()

	go LimitCleanup(mu, limitedClients)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			limited := Limiter(mu, limitedClients, limit, burst, ip)
			if limited {
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
	mu, limitedClients := NewRateLimit()

	go LimitCleanup(mu, limitedClients)

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

					limited := Limiter(mu, limitedClients, limit, burst, ip)
					if limited {
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

			session, err := validateToken(token[0], req.Header.Get("X-TZ"), util.AnonIp(req.RemoteAddr))
			if err != nil {
				deferredErr = util.ErrCheck(err)
				return
			}

			limited := Limiter(mu, limitedClients, limit, burst, session.UserSub)
			if limited {
				WriteLimit(w)
				return
			}

			next(w, req, session)
		}
	}
}

func (a *API) GroupInfoMiddleware(next SessionHandler) SessionHandler {
	return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
		var deferredErr error
		var skipRebuild bool
		gidSelect := req.Header.Get("X-Gid-Select")

		defer func() {
			if deferredErr != nil {
				util.ErrorLog.Println(deferredErr)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
		}()

		existingSession, err := a.Handlers.Redis.GetSession(req.Context(), session.UserSub)
		if err != nil && err.Error() != "redis: nil" {
			deferredErr = util.ErrCheck(err)
			return
		}

		// If the group is not changing via client selection and there's an existing session
		if gidSelect == "" && existingSession != nil {

			// Get the most recent group version (role changes, etc)
			groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(req.Context(), existingSession.GroupId)
			if err != nil {
				deferredErr = util.ErrCheck(err)
				return
			}

			// If the group version has not changed
			if existingSession.GroupSessionVersion == groupVersion {

				// Ignore rebuilding
				skipRebuild = true
			}
		}

		// If we're planning to skip the rebuild, re-use existing session that we now know exists
		if skipRebuild {
			// If group membership changed, rebuild
			if slices.Compare(existingSession.SubGroups, session.SubGroups) != 0 {
				skipRebuild = false
			}

			// If role permissions changed, rebuild
			if slices.Compare(existingSession.AvailableUserGroupRoles, session.AvailableUserGroupRoles) != 0 {
				skipRebuild = false
			}
		}

		// re-use existing session info if no changes
		if skipRebuild && existingSession != nil && len(session.SubGroups) > 0 {

			session = existingSession

		} else {

			// If user belongs to a group, get info about it
			if len(session.SubGroups) > 0 {

				if gidSelect != "" && slices.Contains(session.SubGroups, gidSelect) {
					session.SubGroupName = gidSelect
				} else {
					session.SubGroupName = fmt.Sprint(session.SubGroups[0])
				}

				names := strings.Split(session.SubGroupName, "/")
				session.GroupName = names[1]
				session.RoleName = names[2]
				kcGroupName := "/" + session.GroupName

				kcGroups, err := a.Handlers.Keycloak.GetGroupByName(kcGroupName)

				for _, gr := range kcGroups {
					if gr.Path == kcGroupName {
						session.GroupExternalId = gr.Id
						break
					}
				}

				kcSubgroups, err := a.Handlers.Keycloak.GetGroupSubgroups(session.GroupExternalId)

				for _, gr := range kcSubgroups {
					if gr.Path == session.SubGroupName {
						session.SubGroupExternalId = gr.Id
						break
					}
				}

				if session.GroupExternalId == "" || session.SubGroupExternalId == "" {
					deferredErr = util.ErrCheck(errors.New("could not describe group or subgroup external ids"))
					return
				}

				err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
					var txErr error

					txErr = tx.QueryRow(`
						SELECT id, ai, sub
						FROM dbtable_schema.groups
						WHERE external_id = $1
					`, session.GroupExternalId).Scan(&session.GroupId, &session.GroupAi, &session.GroupSub)
					if txErr != nil {
						return util.ErrCheck(txErr)
					}

					return nil
				}, session.UserSub, session.GroupExternalId, strings.Join(session.AvailableUserGroupRoles, " "))
				if err != nil {
					deferredErr = util.ErrCheck(err)
					return
				}

				groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(req.Context(), session.GroupId)
				if err != nil {
					deferredErr = util.ErrCheck(err)
					return
				}

				session.GroupSessionVersion = groupVersion
			}

			err = a.Handlers.Redis.SetSession(req.Context(), session)
			if err != nil {
				deferredErr = util.ErrCheck(err)
				return
			}
		}

		next(w, req, session)
	}
}

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {
	return func(next SessionHandler) SessionHandler {

		if opts.SiteRole == "" || opts.SiteRole == types.SiteRoles_UNRESTRICTED.String() {
			return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
				next(w, req, session)
			}
		} else {
			return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
				hasSiteRole := slices.Contains(session.AvailableUserGroupRoles, opts.SiteRole)

				fmt.Println(fmt.Sprintf("access of %s, request allowed: %v", req.URL, hasSiteRole))

				if !hasSiteRole {
					http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
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

type CacheMeta struct {
	Data       []byte    `json:"data"`
	LastMod    time.Time `json:"last_modified"`
	StatusCode int       `json:"status_code"`
}

func (a *API) CacheMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {

	duration180, _ := time.ParseDuration("180s")
	shouldStore := types.CacheType_STORE == opts.CacheType

	return func(next SessionHandler) SessionHandler {
		return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
			// gives a cache key like absd-asff-asff-asfdgroup/users
			cacheKey := session.UserSub + strings.TrimLeft(req.URL.String(), os.Getenv("API_PATH"))

			// Any non-GET processed normally, and deletes cache key unless being stored
			if !shouldStore && req.Method != http.MethodGet {
				next(w, req, session)
				if types.CacheType_STORE != opts.CacheType {
					a.Handlers.Redis.Client().Del(req.Context(), cacheKey)
				}
				return
			}

			// Check if client sent If-Modified-Since header
			ifModifiedSince := req.Header.Get("If-Modified-Since")
			var clientModTime time.Time
			if ifModifiedSince != "" {
				if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
					clientModTime = t.Truncate(time.Second)
				}
			}

			if types.CacheType_SKIP != opts.CacheType {
				// Check redis cache for request
				if cacheData, err := a.Handlers.Redis.Client().Get(req.Context(), cacheKey).Bytes(); err == nil {

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

			// Perform the handler request
			cacheWriter := &CacheWriter{
				ResponseWriter: w,
				Buffer:         new(bytes.Buffer),
			}

			// Response is written out to client
			next(cacheWriter, req, session)

			// Cache any response
			if cacheWriter.Buffer.Len() > 0 {

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
					a.Handlers.Redis.Client().Set(req.Context(), cacheKey, cacheData, 0)

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

					a.Handlers.Redis.Client().SetEx(req.Context(), cacheKey, cacheData, duration)
				}
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
