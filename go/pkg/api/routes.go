package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
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
		util.DebugLog.Println("Service Method Not Implemented:", serviceName)
		return func(w http.ResponseWriter, r *http.Request, session *types.UserSession) {
			w.WriteHeader(501)
			return
		}
	}

	handlerOpts := util.ParseHandlerOptions(serviceMethod)
	ignoreFields := slices.Concat(util.DEFAULT_IGNORED_PROTO_FIELDS, handlerOpts.NoLogFields)

	var queryParser = func(msg proto.Message, req *http.Request) {}
	if handlerOpts.HasQueryParams {
		queryParser = util.ParseProtoQueryParams
	}

	var pathParser = func(msg proto.Message, req *http.Request) {}
	if handlerOpts.HasPathParams {
		pathParams := strings.Split(handlerOpts.ServiceMethodURL, "/")
		pathParser = func(msg proto.Message, req *http.Request) {
			util.ParseProtoPathParams(msg, pathParams, req)
		}
	}

	var bodyParser BodyParser = ProtoBodyParser
	if handlerOpts.MultipartRequest {
		bodyParser = MultipartBodyParser
	}

	var requestExecutor RequestExecutor = BatchExecutor
	if handlerOpts.UseTx {
		requestExecutor = TxExecutor
	}

	var responseHandler ResponseHandler = ProtoResponseHandler
	if handlerOpts.MultipartResponse {
		responseHandler = MultipartResponseHandler
	}

	return func(w http.ResponseWriter, req *http.Request, session *types.UserSession) {
		var err error
		var pb proto.Message

		defer func() {
			if p := recover(); p != nil {
				// ErrCheck handled errors will already have a trace
				// If there is no trace file hint, print the full stack
				errStr := fmt.Sprint(p)
				if !strings.Contains(errStr, ".go:") {
					util.ErrorLog.Println(string(debug.Stack()))
				}

				util.RequestError(w, errStr, ignoreFields, pb)
			}

			clients.GetGlobalWorkerPool().CleanUpClientMapping(session.UserSub)
		}()

		pb, err = bodyParser(w, req, handlerOpts, serviceType)
		if err != nil {
			panic(util.ErrCheck(err))
		}

		queryParser(pb, req)

		pathParser(pb, req)

		ctx := req.Context()

		reqInfo, done, err := requestExecutor(ctx, a.Handlers.Database.DatabaseClient, session)
		if err != nil {
			panic(util.ErrCheck(err))
		}
		defer done()

		reqInfo.Ctx = ctx
		reqInfo.W = w
		reqInfo.Req = req
		reqInfo.Session = session

		results, err := handlerFunc(reqInfo, pb)
		if err != nil {
			panic(util.ErrCheck(err))
		}

		if results != nil {
			_, err = responseHandler(w, results)
			if err != nil {
				panic(util.ErrCheck(err))
			}
		}
	}
}
