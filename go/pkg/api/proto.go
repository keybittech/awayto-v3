package api

import (
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func (a *API) InitProtoHandlers() {
	sessionMux := NewSessionMux(a.Handlers.Keycloak.Client.PublicKey, a.Handlers.Cache)

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}

		services := fd.Services().Get(0)

		for i := 0; i <= services.Methods().Len()-1; i++ {
			serviceMethod := services.Methods().Get(i)
			handlerOpts := util.ParseHandlerOptions(serviceMethod)

			sessionMux.Handle(handlerOpts.Pattern,
				a.CacheMiddleware(handlerOpts)(
					a.SiteRoleCheckMiddleware(handlerOpts)(
						a.GroupInfoMiddleware(
							a.HandleRequest(serviceMethod),
						),
					),
				),
			)
		}

		return true
	})

	a.Server.Handler.(*http.ServeMux).Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionMux.mux.ServeHTTP(w, r)
	}))
}
