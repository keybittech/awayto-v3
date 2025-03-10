package api

import (
	"fmt"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/handlers"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type API struct {
	Server   *http.Server
	Handlers *handlers.Handlers
}

func (a *API) InitMux() *http.ServeMux {
	apiMux := http.NewServeMux()

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}
		println(fmt.Sprintf("Attaching services for %s", fd.Path()))
		a.BuildProtoService(apiMux, fd)
		return true
	})

	mux := http.NewServeMux()
	mux.Handle("/api/", a.LimitMiddleware(2, 20)(func(w http.ResponseWriter, r *http.Request) {
		apiMux.ServeHTTP(w, r)
	}))

	a.Server.Handler = mux
	return mux
}
