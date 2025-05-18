package api

import (
	"errors"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"google.golang.org/protobuf/proto"
)

func registerProtoHandler[ReqMsg, ResMsg proto.Message](handler handlers.TypedProtoHandler[ReqMsg, ResMsg]) handlers.ProtoHandler {
	return func(info handlers.ReqInfo, message proto.Message) (proto.Message, error) {
		msg, ok := message.(ReqMsg)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return handler(info, msg)
	}
}

func (a *API) InitProtoHandlers() {
	sessionMux := NewSessionMux(a.Handlers.Keycloak.Client.PublicKey, a.Cache.UserSessions)

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
