package api

import (
	"net/http"
)

func (a *API) InitProtoHandlers() {
	for _, handlerOpts := range a.Handlers.Options {
		a.Server.Handler.(*http.ServeMux).Handle(handlerOpts.Pattern,
			a.ValidateSessionMiddleware()(
				a.CacheMiddleware(handlerOpts)(
					a.SiteRoleCheckMiddleware(handlerOpts)(
						a.GroupInfoMiddleware(
							a.HandleRequest(handlerOpts),
						),
					),
				),
			),
		)
	}
}
