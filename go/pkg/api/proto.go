package api

import (
	"net/http"
)

func (a *API) InitProtoHandlers() {
	sessionMux := NewSessionMux(a.Handlers.Keycloak.Client.PublicKey, a.Handlers.Cache)

	for _, handlerOpts := range a.Handlers.Options {
		sessionMux.Handle(handlerOpts.Pattern,
			a.CacheMiddleware(handlerOpts)(
				a.SiteRoleCheckMiddleware(handlerOpts)(
					a.GroupInfoMiddleware(
						a.HandleRequest(handlerOpts),
					),
				),
			),
		)
	}

	a.Server.Handler.(*http.ServeMux).Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionMux.mux.ServeHTTP(w, r)
	}))
}
