package api

import (
	"fmt"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type API struct {
	Server   *http.Server
	Handlers *handlers.Handlers
}

func (a *API) InitMux() *http.ServeMux {
	// apiMux := http.NewServeMux()

	mapa := make(map[string]SessionHandler)

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}
		println(fmt.Sprintf("Attaching services for %s", fd.Path()))

		services := fd.Services().Get(0)

		for i := 0; i <= services.Methods().Len()-1; i++ {
			serviceMethod := services.Methods().Get(i)
			handlerOpts := util.ParseHandlerOptions(serviceMethod)

			mapa[handlerOpts.Pattern] = a.SiteRoleCheckMiddleware(handlerOpts)(
				a.CacheMiddleware(handlerOpts)(
					a.HandleRequest(serviceMethod),
				),
			)
		}

		// a.BuildProtoService(apiMux, fd)
		return true
	})

	mux := http.NewServeMux()
	mux.Handle("/api/",
		a.ValidateTokenMiddleware(2, 20)(
			a.GroupInfoMiddleware(
				func(w http.ResponseWriter, r *http.Request, session *clients.UserSession) {
					mapa[r.Method+" "+r.URL.Path](w, r, session)
				},
			),
		),
	)

	a.Server.Handler = mux
	return mux
}
