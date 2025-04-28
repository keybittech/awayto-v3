package api

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/proto"
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

	serviceName := string(serviceMethod.Name())
	handlerFunc, ok := a.Handlers.Functions[serviceName]
	if !ok {
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
		var pb proto.Message

		defer func() {
			if p := recover(); p != nil {
				util.ErrorLog.Println("ROUTE RECOVERY", fmt.Sprint(p))
			}
			go func() {
				clients.GetGlobalWorkerPool().CleanUpClientMapping(session.UserSub)

				if deferredError != nil {
					util.RequestError(w, deferredError.Error(), ignoreFields, pb)
				}
			}()
		}()

		pb, err := bodyParser(w, req, handlerOpts, serviceType)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}

		// Parse query and path parameters
		util.ParseProtoQueryParams(pb, req.URL.Query())
		util.ParseProtoPathParams(
			pb,
			strings.Split(handlerOpts.ServiceMethodURL, "/"),
			strings.Split(strings.TrimPrefix(req.URL.Path, "/api"), "/"),
		)

		ctx := req.Context()

		poolTx, err := a.Handlers.Database.DatabaseClient.OpenPoolSessionTx(ctx, session)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}
		defer poolTx.Rollback(ctx)

		reqInfo := handlers.ReqInfo{
			W:       w,
			Req:     req,
			Session: session,
			Tx:      poolTx,
		}

		results, err := handlerFunc(reqInfo, pb)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}

		err = a.Handlers.Database.DatabaseClient.ClosePoolSessionTx(ctx, poolTx)
		if err != nil {
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
