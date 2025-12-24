package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"google.golang.org/protobuf/proto"
)

func (a *API) HandleRequest(handlerOpts *util.HandlerOptions) SessionHandler {
	requestHandler, ok := a.Handlers.Functions[handlerOpts.ServiceMethodName]
	if !ok {
		util.DebugLog.Println("Service Method Not Implemented:", handlerOpts.ServiceMethodName)
		return func(w http.ResponseWriter, r *http.Request, session *types.ConcurrentUserSession) {
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
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
		var requestBody proto.Message
		var executor handlers.ReqInfo
		var done func(error) error

		defer func() {
			if p := recover(); p != nil {
				// if done was not disarmed, a genuine panic occurred somewhere before/during the handler
				if done != nil {
					_ = done(fmt.Errorf("panic occurred: %v", p))
				}

				// ErrCheck handled errors will already have a trace
				// If there is no trace file hint, print the full stack
				var sb strings.Builder
				sb.WriteString("Service: ")
				sb.WriteString(methodName)
				sb.WriteByte(' ')
				sb.WriteString(fmt.Sprint(p))
				errStr := sb.String()

				util.RequestError(w, errStr, noLogFields, requestBody)

				// if this is a genuine panic not handled by ErrCheck-ing, which adds the ".go:" text
				if !strings.Contains(errStr, ".go:") {
					util.ErrorLog.Println(string(debug.Stack()))
				}
			}

			clients.GetGlobalWorkerPool().CleanUpClientMapping(session.GetUserSub())
		}()

		requestBody = bodyParser(w, req, msgType)
		queryParser(requestBody, req)
		pathParser(requestBody, req)

		ctx := req.Context()

		executor, done = requestExecutor(ctx, w, req, session, a.Handlers.Database.DatabaseClient)

		handlerResponse, handlerErr := requestHandler(executor, requestBody)

		// if handler errors, the done fn will return that error and roll back the tx
		// otherwise this will return errors during unsetting session and tx commit
		doneErr := done(handlerErr)

		// disarm for defer
		done = nil

		// if error occurs during rollback
		if doneErr != nil {
			panic(doneErr)
		}

		resetGroup(req, session.GetGroupId())

		responseHandler(w, req, handlerResponse)
	}
}
