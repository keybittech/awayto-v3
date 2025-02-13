package api

import (
	handlers "av3api/pkg/handlers"
	"fmt"
	"net/http"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type API struct {
	Server   *http.Server
	Handlers *handlers.Handlers
}

func (a *API) InitMux() *http.ServeMux {
	mux := http.NewServeMux()

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}
		println(fmt.Sprintf("Attaching services for %s", fd.Path()))
		a.BuildProtoService(mux, fd)
		return true
	})

	a.Server.Handler = mux
	return mux
}
