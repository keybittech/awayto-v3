package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/proto"
)

func (a *API) HandleRequest(handlerOpts *util.HandlerOptions) SessionHandler {
	handlerFunc, ok := a.Handlers.Functions[handlerOpts.ServiceMethodName]
	if !ok {
		util.DebugLog.Println("Service Method Not Implemented:", handlerOpts.ServiceMethodName)
		return func(w http.ResponseWriter, r *http.Request, session *types.ConcurrentUserSession) {
			w.WriteHeader(501)
			return
		}
	}

	msgType := handlerOpts.ServiceMethodInputType
	methodName := handlerOpts.ServiceMethodName
	noLogFields := handlerOpts.NoLogFields
	optPack := handlerOpts.Unpack()

	var queryParser = func(proto.Message, *http.Request) {}
	if optPack.HasQueryParams {
		queryParser = util.ParseProtoQueryParams
	}

	var pathParser = func(proto.Message, *http.Request) {}
	if optPack.HasPathParams {
		pathParams := strings.Split(handlerOpts.ServiceMethodURL, "/")
		pathParser = func(msg proto.Message, req *http.Request) {
			util.ParseProtoPathParams(msg, pathParams, req)
		}
	}

	var bodyParser BodyParser = ProtoBodyParser
	if optPack.MultipartRequest {
		bodyParser = MultipartBodyParser
	}

	var requestExecutor RequestExecutor = BatchExecutor
	if optPack.UseTx {
		requestExecutor = TxExecutor
	}

	var resetGroup = func(*http.Request, string) {}
	if optPack.ResetsGroup {
		resetGroup = a.Handlers.ResetGroupSession
	}

	var responseHandler ResponseHandler = ProtoResponseHandler
	if optPack.MultipartResponse {
		responseHandler = MultipartResponseHandler
	}

	return func(w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession) {
		var err error
		var pb proto.Message

		defer func() {
			if p := recover(); p != nil {
				// ErrCheck handled errors will already have a trace
				// If there is no trace file hint, print the full stack
				var sb strings.Builder
				sb.WriteString("Service: ")
				sb.WriteString(methodName)
				sb.WriteByte(' ')
				sb.WriteString(fmt.Sprint(p))
				errStr := sb.String()

				util.RequestError(w, errStr, noLogFields, pb)

				if !strings.Contains(errStr, ".go:") {
					util.ErrorLog.Println(string(debug.Stack()))
				}
			}

			clients.GetGlobalWorkerPool().CleanUpClientMapping(session.GetUserSub())
		}()

		pb = bodyParser(w, req, msgType)
		queryParser(pb, req)
		pathParser(pb, req)

		ctx := req.Context()

		reqInfo, done := requestExecutor(ctx, a.Handlers.Database.DatabaseClient, session)

		reqInfo.Ctx = ctx
		reqInfo.W = w
		reqInfo.Req = req
		reqInfo.Session = session

		results, err := handlerFunc(reqInfo, pb)
		if err != nil {
			done()
			panic(err)
		}
		done()

		resetGroup(req, session.GetGroupId())

		responseHandler(w, results)
	}
}
