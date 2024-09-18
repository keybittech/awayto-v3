package api

import (
	handlers "av3api/pkg/handlers"
	"fmt"
	"net/http"
	"os"

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

	mux.HandleFunc(fmt.Sprintf("OPTIONS %s*", os.Getenv("API_PATH")), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("APP_HOST_URL"))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH")
		w.WriteHeader(http.StatusOK)
	}))

	a.Server.Handler = mux
	return mux
}
