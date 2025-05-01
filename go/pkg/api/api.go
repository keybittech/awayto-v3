package api

import (
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type API struct {
	Server   *http.Server
	Handlers *handlers.Handlers
}

var (
	apiRl = NewRateLimit("api", 5, 20, time.Duration(5*time.Minute))
)

func (a *API) InitMux() *http.ServeMux {
	muxWithSession := NewSessionMux(a.Handlers.Keycloak.Client.PublicKey, a.Handlers.Redis.RedisClient)

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}

		services := fd.Services().Get(0)

		for i := 0; i <= services.Methods().Len()-1; i++ {
			serviceMethod := services.Methods().Get(i)
			handlerOpts := util.ParseHandlerOptions(serviceMethod)

			muxWithSession.Handle(handlerOpts.Pattern,
				a.GroupInfoMiddleware(
					a.SiteRoleCheckMiddleware(handlerOpts)(
						a.CacheMiddleware(handlerOpts)(
							a.HandleRequest(serviceMethod),
						),
					),
				),
			)
		}

		return true
	})

	mux := http.NewServeMux()
	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		muxWithSession.ServeHTTP(w, r)
	}))

	a.Server.Handler = mux
	return mux
}
