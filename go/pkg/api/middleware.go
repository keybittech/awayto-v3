package api

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"bytes"
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
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

func (a *API) SiteRoleCheckMiddleware(opts *util.HandlerOptions) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {

		if opts.SiteRole == "" || opts.SiteRole == types.SiteRoles_UNRESTRICTED.String() {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req)
			})
		} else {

			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				session, err := a.Handlers.Redis.ReqSession(req)
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
					return
				}

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
			session, err := a.Handlers.Redis.ReqSession(req)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

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
								util.ErrorLog.Println(util.ErrCheck(err))
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

// From https://blog.logrocket.com/rate-limiting-go-application
func (a *API) LimitMiddleware(limit rate.Limit, burst int) func(next http.HandlerFunc) http.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(next http.HandlerFunc) http.HandlerFunc {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the IP address from the request.
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(limit, burst)}
			}
			clients[ip].lastSeen = time.Now()
			if !clients[ip].limiter.Allow() {
				mu.Unlock()

				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			mu.Unlock()
			next(w, r)
		})
	}
}
