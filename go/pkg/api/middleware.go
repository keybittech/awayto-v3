package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type SessionHandler func(w http.ResponseWriter, r *http.Request, session *clients.UserSession)

// From https://blog.logrocket.com/rate-limiting-go-application
func (a *API) LimitMiddleware(limit rate.Limit, burst int) func(next SessionHandler) http.HandlerFunc {
	type limitedClient struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu             sync.Mutex
		limitedClients = make(map[string]*limitedClient)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			for ip, lc := range limitedClients {
				if time.Since(lc.lastSeen) > 3*time.Minute {
					delete(limitedClients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(next SessionHandler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Extract the IP address from the request.
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			ip = util.AnonIp(ip)

			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			if _, found := limitedClients[ip]; !found {
				limitedClients[ip] = &limitedClient{limiter: rate.NewLimiter(limit, burst)}
			}
			limitedClients[ip].lastSeen = time.Now()
			if !limitedClients[ip].limiter.Allow() {
				mu.Unlock()

				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(http.StatusText(http.StatusTooManyRequests)))
				return
			}
			mu.Unlock()

			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("APP_HOST_URL"))
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-TZ")
			w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH")

			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, req, nil)
		})
	}
}

func (a *API) ValidateTokenMiddleware(next SessionHandler) SessionHandler {
	return func(w http.ResponseWriter, req *http.Request, session *clients.UserSession) {
		var deferredErr error

		defer func() {
			if deferredErr != nil {
				util.ErrorLog.Println(deferredErr)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
		}()

		token, ok := req.Header["Authorization"]
		if !ok {
			deferredErr = errors.New("no auth token")
			return
		}

		kcUser, err := a.Handlers.Keycloak.GetUserTokenValid(token[0])
		if err != nil {
			deferredErr = err
			return
		}

		session = &clients.UserSession{
			UserSub:                 kcUser.Sub,
			UserEmail:               kcUser.Email,
			SubGroups:               kcUser.Groups,
			AvailableUserGroupRoles: kcUser.ResourceAccess[kcUser.Azp].Roles,
			Timezone:                req.Header.Get("X-TZ"),
			AnonIp:                  util.AnonIp(req.RemoteAddr),
		}

		next(w, req, session)
	}
}

func (a *API) GroupInfoMiddleware(next SessionHandler) SessionHandler {
	return func(w http.ResponseWriter, req *http.Request, session *clients.UserSession) {
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

		// If the group is not changing and there's an existing session
		if gidSelect == "" && existingSession != nil {

			// Get the most recent group version (role changes, etc)
			groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(req.Context(), existingSession.GroupId)
			if err != nil {
				deferredErr = util.ErrCheck(err)
				return
			}

			// If the group version has not changed
			if existingSession.GroupSessionVersion == groupVersion {

				// Rebuild the group info
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

		// Keep separate so above checks can take place
		if !skipRebuild {

			// If user belongs to a group, get info about it
			if len(session.SubGroups) > 0 {

				if existingSession != nil {
					// Update contextually relevant info into the existing session
					existingSession.Timezone = session.Timezone
					existingSession.AnonIp = session.AnonIp

					// re-use existing session info
					session = existingSession
				}

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

				for _, gr := range *kcGroups {
					if gr.Path == kcGroupName {
						session.GroupExternalId = gr.Id
						break
					}
				}

				kcSubgroups, err := a.Handlers.Keycloak.GetGroupSubgroups(session.GroupExternalId)

				for _, gr := range *kcSubgroups {
					if gr.Path == session.SubGroupName {
						session.SubGroupExternalId = gr.Id
						break
					}
				}

				if session.GroupExternalId == "" || session.SubGroupExternalId == "" {
					deferredErr = util.ErrCheck(errors.New("could not describe group or subgroup external ids"))
					return
				}

				err = a.Handlers.Database.TxExec(func(tx clients.IDatabaseTx) error {
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
			return func(w http.ResponseWriter, req *http.Request, session *clients.UserSession) {
				next(w, req, session)
			}
		} else {
			return func(w http.ResponseWriter, req *http.Request, session *clients.UserSession) {
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

func (cw *CacheWriter) Write(data []byte) (int, error) {
	defer cw.Buffer.Write(data)
	return cw.ResponseWriter.Write(data)
}

func (a *API) CacheMiddleware(opts *util.HandlerOptions) func(SessionHandler) SessionHandler {

	duration180, _ := time.ParseDuration("180s")
	shouldStore := types.CacheType_STORE == opts.CacheType

	return func(next SessionHandler) SessionHandler {
		return func(w http.ResponseWriter, req *http.Request, session *clients.UserSession) {
			cacheKey := session.UserSub + strings.TrimLeft(req.URL.String(), os.Getenv("API_PATH")) // gives a cache key like absd-asff-asff-asfdgroup/users

			if shouldStore || req.Method == http.MethodGet && types.CacheType_SKIP != opts.CacheType {
				cachedRes, _ := a.Handlers.Redis.Client().Get(req.Context(), cacheKey).Bytes()
				if cachedRes != nil {
					w.Header().Set("X-Cache-Status", "HIT")
					w.Write(cachedRes)
					return
				}
			}

			w.Header().Set("X-Cache-Status", "MISS")

			if shouldStore || req.Method == http.MethodGet {

				cacheWriter := &CacheWriter{
					ResponseWriter: w,
					Buffer:         new(bytes.Buffer),
				}

				next(cacheWriter, req, session)

				resBytes := cacheWriter.Buffer.Bytes()

				if len(resBytes) > 0 {
					if shouldStore {
						a.Handlers.Redis.Client().Set(req.Context(), cacheKey, resBytes, 0)
					} else {
						duration := duration180
						if opts.CacheDuration > 0 {
							var err error
							duration, err = time.ParseDuration(fmt.Sprintf("%ds", opts.CacheDuration))
							if err != nil {
								util.ErrorLog.Println(util.ErrCheck(err))
								duration = duration180
							}
						}

						a.Handlers.Redis.Client().SetEx(req.Context(), cacheKey, resBytes, duration)
					}
				}
			} else {
				next(w, req, session)
				if types.CacheType_STORE != opts.CacheType {
					a.Handlers.Redis.Client().Del(req.Context(), cacheKey)
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
