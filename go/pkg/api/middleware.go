package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func ApplyMiddleware(h http.HandlerFunc, middlewares []Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

func (a *API) CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("APP_HOST_URL"))

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (a *API) SocketAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		var userSub, userEmail string

		ticket := req.URL.Query().Get("ticket")

		if ticket == "" {
			util.ErrCheck(errors.New("no ticket during socket auth"))
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		subscriber, err := a.Handlers.Socket.GetSubscriberByTicket(ticket)
		if err != nil {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		if subscriber.UserSub == "" {
			util.ErrCheck(errors.New("err getting user sub during socket auth"))
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		kcUser, err := a.Handlers.Keycloak.GetUserInfoById(subscriber.UserSub)

		if err != nil {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		userSub = kcUser.Sub
		userEmail = kcUser.Email

		if userSub == "" || userEmail == "" {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		ctx = context.WithValue(ctx, "UserSession", &clients.UserSession{UserSub: userSub, UserEmail: userEmail})
		ctx = context.WithValue(ctx, "SourceIp", req.RemoteAddr)

		req = req.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

func (a *API) SessionAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		token, ok := req.Header["Authorization"]

		if !ok {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		tokenParts := strings.Split(token[0], " ")

		if len(tokenParts) != 2 {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		_, payload, err := util.ParseJWT(tokenParts[1])
		if err != nil {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		var userSub string
		if us, ok := payload["sub"]; ok {
			userSub = fmt.Sprint(us)
		} else {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		buildSession := true

		gidSelect := req.Header.Get("X-Gid-Select")

		session, err := a.Handlers.Redis.GetSession(ctx, userSub)

		if err == nil && gidSelect == "" {
			groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(ctx, session.GroupId)
			if err != nil {
				http.Error(w, util.InternalErrorResponse, http.StatusInternalServerError)
				return
			}

			if session.GroupSessionVersion == groupVersion && session.GroupId != "" {
				buildSession = false
			} else {
				session.GroupSessionVersion = groupVersion
			}
		}

		if buildSession {

			session = &clients.UserSession{
				UserSub: userSub,
			}

			if em, ok := payload["email"]; ok {
				session.UserEmail = fmt.Sprint(em)
			} else {
				util.ErrCheck(errors.New("no email "))
				http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
				return
			}

			if g, ok := payload["groups"]; ok {

				switch t := g.(type) {
				case []interface{}:
					if groups, ok := util.CastSlice[string](t); ok {
						session.SubGroups = groups
					} else {
						util.ErrCheck(errors.New("bad group slice " + fmt.Sprint(g)))
						http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
						return
					}
				default:
					util.ErrCheck(errors.New("groups in bad format " + fmt.Sprint(g)))
					http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
					return
				}

				if len(session.SubGroups) > 0 {

					if gidSelect != "" {
						session.SubGroupName = gidSelect
					} else {
						session.SubGroupName = fmt.Sprint(session.SubGroups[0])
					}

					if session.SubGroupName == "" {
						util.ErrCheck(errors.New("no group names found "))
						http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
						return
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
						util.ErrCheck(errors.New("no group ext id found " + session.GroupName))
						http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
						return
					}

					sgRoles := a.Handlers.Keycloak.GetGroupSiteRoles(session.SubGroupExternalId)
					for _, mapping := range sgRoles {
						session.AvailableUserGroupRoles = append(session.AvailableUserGroupRoles, mapping.Name)
					}

					err = a.Handlers.Database.Client().QueryRow(`
						SELECT g.id, g.ai, u.sub
						FROM dbtable_schema.groups g
						JOIN dbtable_schema.users u ON u.username = CONCAT('system_group_', g.id)
						WHERE g.external_id = $1
					`, session.GroupExternalId).Scan(&session.GroupId, &session.GroupAi, &session.GroupSub)
					if err != nil {
						util.ErrCheck(err)
						http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
						return
					}

					groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(ctx, session.GroupId)
					if err != nil {
						http.Error(w, util.InternalErrorResponse, http.StatusInternalServerError)
						return
					} else {
						session.GroupSessionVersion = groupVersion
					}
				}
			}

			a.Handlers.Redis.SetSession(ctx, userSub, session)
		}

		if session.UserSub == "" || session.UserEmail == "" {
			http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
			return
		}

		ctx = context.WithValue(ctx, "UserSession", session)
		ctx = context.WithValue(ctx, "SourceIp", req.RemoteAddr)

		req = req.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {

		if opts.SiteRole == "" || opts.SiteRole == types.SiteRoles_UNRESTRICTED.String() {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req)
			})
		} else {

			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

				session := a.Handlers.Redis.ReqSession(req)
				hasSiteRole := slices.Contains(session.AvailableUserGroupRoles, opts.SiteRole)

				fmt.Println(fmt.Sprintf("access of %s, request allowed: %v", req.URL, hasSiteRole))

				if !hasSiteRole {
					http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, req)
			})
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

func (a *API) CacheMiddleware(opts *util.HandlerOptions) func(http.HandlerFunc) http.HandlerFunc {

	duration180, _ := time.ParseDuration("180s")
	shouldStore := types.CacheType_STORE == opts.CacheType

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {

			session := a.Handlers.Redis.ReqSession(req)
			cacheKey := session.UserSub + strings.TrimLeft(req.URL.String(), os.Getenv("API_PATH")) // gives a cache key like absd-asff-asff-asfdgroup/users

			ctx := req.Context()

			if shouldStore || req.Method == http.MethodGet && types.CacheType_SKIP != opts.CacheType {
				cachedRes, _ := a.Handlers.Redis.Client().Get(ctx, cacheKey).Bytes()
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

				next.ServeHTTP(cacheWriter, req)

				resBytes := cacheWriter.Buffer.Bytes()

				if len(resBytes) > 0 {
					if shouldStore {
						a.Handlers.Redis.Client().Set(ctx, cacheKey, resBytes, 0)
					} else {
						duration := duration180
						if opts.CacheDuration > 0 {
							var err error
							duration, err = time.ParseDuration(fmt.Sprintf("%ds", opts.CacheDuration))
							if err != nil {
								util.ErrCheck(err)
								duration = duration180
							}
						}

						a.Handlers.Redis.Client().SetEx(ctx, cacheKey, resBytes, duration)
					}
				}
			} else {
				next.ServeHTTP(w, req)
				if types.CacheType_STORE != opts.CacheType {
					a.Handlers.Redis.Client().Del(ctx, cacheKey)
				}
			}
		}
	}
}
