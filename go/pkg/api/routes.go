package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
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
	ignoreFields := slices.Concat([]string{"state", "sizeCache", "unknownFields"}, handlerOpts.NoLogFields)

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

		exeTimeDefer := util.ExeTime(handlerOpts.Pattern)

		defer func() {
			if p := recover(); p != nil {
				panic(p)
			} else if deferredError != nil {
				util.RequestError(w, deferredError.Error(), ignoreFields, pbVal)
			}
		}()

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

		results := []reflect.Value{}

		err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
			results = handlerFunc.Call([]reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(req),
				reflect.ValueOf(pb),
				reflect.ValueOf(session),
				reflect.ValueOf(tx),
			})

			if len(results) != 2 {
				return util.ErrCheck(errors.New("badly formed handler"))
			}

			if err, ok := results[1].Interface().(error); ok {
				return util.ErrCheck(err)
			}

			return nil
		}, session.UserSub, session.GroupId, strings.Join(session.AvailableUserGroupRoles, " "))
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}

		resLen, err := responseHandler(w, results)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}

		defer exeTimeDefer(fmt.Sprintf("response len %d", resLen))
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
