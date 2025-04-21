package api

import (
	"errors"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func (a *API) HandleRequest(serviceMethod protoreflect.MethodDescriptor) SessionHandler {
	var serviceType protoreflect.MessageType
	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		if mt.Descriptor().FullName() == serviceMethod.Input().FullName() {
			serviceType = mt
			return false
		}
		return true
	})

	serviceName := serviceMethod.Name()
	handlerFunc := reflect.ValueOf(a.Handlers).MethodByName(string(serviceName))

	if !handlerFunc.IsValid() {
		log.Printf("Service Method Not Implemented: %s", serviceName)
		return func(w http.ResponseWriter, r *http.Request, session *types.UserSession) {
			w.WriteHeader(501)
			return
		}
	}

	var bodyParser BodyParser
	var responseHandler ResponseHandler
	handlerOpts := util.ParseHandlerOptions(serviceMethod)
	ignoreFields := slices.Concat(util.DEFAULT_IGNORED_PROTO_FIELDS, handlerOpts.NoLogFields)

	if handlerOpts.MultipartRequest {
		bodyParser = MultipartBodyParser
	} else {
		bodyParser = ProtoBodyParser
	}

	if handlerOpts.MultipartResponse {
		responseHandler = MultipartResponseHandler
	} else {
		responseHandler = ProtoResponseHandler
	}

	return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
		var deferredError error
		var pbVal reflect.Value

		defer func(us string) {
			clients.GetGlobalWorkerPool().CleanUpClientMapping(us)
			if p := recover(); p != nil {
				panic(p)
			} else if deferredError != nil {
				util.RequestError(w, deferredError.Error(), ignoreFields, pbVal)
			}
		}(session.UserSub)

		pb, err := bodyParser(w, req, handlerOpts, serviceType)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}

		pbVal = reflect.ValueOf(pb).Elem()

		// Parse query and path parameters
		util.ParseProtoQueryParams(pbVal, req.URL.Query())
		util.ParseProtoPathParams(
			pbVal,
			strings.Split(handlerOpts.ServiceMethodURL, "/"),
			strings.Split(strings.TrimPrefix(req.URL.Path, "/api"), "/"),
		)

		tx, err := a.Handlers.Database.DatabaseClient.Pool.Begin(req.Context())
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}
		defer tx.Rollback(req.Context())

		poolTx := &clients.PoolTx{Tx: tx}

		poolTx.SetSession(session)

		results := handlerFunc.Call([]reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(req),
			reflect.ValueOf(pb),
			reflect.ValueOf(session),
			reflect.ValueOf(poolTx),
		})

		poolTx.UnsetSession()
		poolTx.Commit()

		if len(results) != 2 {
			deferredError = util.ErrCheck(errors.New("badly formed handler"))
			return
		}

		if err, ok := results[1].Interface().(error); ok {
			deferredError = util.ErrCheck(err)
			return
		}

		_, err = responseHandler(w, results)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}
	}
}

// func (a *API) BuildProtoService(mux *http.ServeMux, fd protoreflect.FileDescriptor) map[string]SessionHandler {
//
// 		mux.HandleFunc(handlerOpts.Pattern,
// 			a.ValidateTokenMiddleware(
// 				a.GroupInfoMiddleware(
// 					,
// 				),
// 			),
// 		)
// 	}
// }
