package api

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func (a *API) BuildProtoService(mux *http.ServeMux, fd protoreflect.FileDescriptor) {
	services := fd.Services().Get(0)

	for i := 0; i <= services.Methods().Len()-1; i++ {
		serviceMethod := services.Methods().Get(i)

		var serviceType protoreflect.MessageType
		protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
			if mt.Descriptor().FullName() == serviceMethod.Input().FullName() {
				serviceType = mt
				return false
			}
			return true
		})

		if serviceType == nil {
			log.Printf("failed service type %s", serviceMethod.Name())
			continue
		}

		serviceName := serviceMethod.Name()
		handlerFunc := reflect.ValueOf(a.Handlers).MethodByName(string(serviceName))

		if !handlerFunc.IsValid() {
			log.Printf("Service Method Not Implemented: %s", serviceName)
			continue
		}

		handlerOpts := util.ParseHandlerOptions(serviceMethod)

		ignoreFields := slices.Concat([]string{"state", "sizeCache", "unknownFields"}, handlerOpts.NoLogFields)

		protoHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var deferredError error
			var pbVal reflect.Value

			exeTimeDefer := util.ExeTime(handlerOpts.Pattern)

			requestId := uuid.NewString()

			// Setup handler transaction

			tx, err := a.Handlers.Database.Client().Begin()
			if err != nil {
				util.RequestError(w, requestId, util.ErrCheck(err).Error(), ignoreFields, pbVal)
				return
			}

			// Setup handler deferrals

			defer func() {
				var deferralError error

				err = tx.SetDbVar("user_sub", "")
				if err != nil {
					deferralError = util.ErrCheck(err)
				}

				err = tx.SetDbVar("group_id", "")
				if err != nil {
					deferralError = util.ErrCheck(err)
				}

				if p := recover(); p != nil {
					tx.Rollback()
					panic(p)
				} else if deferredError != nil {
					tx.Rollback()
				} else {
					err = tx.Commit()
					if err != nil {
						deferralError = util.ErrCheck(err)
					}
				}

				if deferredError != nil {
					loggedError := deferredError.Error()
					if deferralError != nil {
						loggedError = fmt.Sprintf("%s %s", deferredError.Error(), deferralError.Error())
					}
					util.RequestError(w, requestId, loggedError, ignoreFields, pbVal)
				}
			}()

			// Authorize the request

			session, err := a.GetAuthorizedSession(req, tx)
			if err != nil {
				util.RequestError(w, requestId, util.ErrCheck(err).Error(), ignoreFields, pbVal)
				return
			}

			// Transform the request body to a protobuf struct

			pb := serviceType.New().Interface()

			if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "files/content") {
				req.ParseMultipartForm(20480000)

				files := req.MultipartForm.File["contents"]

				pbFiles := &types.PostFileContentsRequest{}

				for _, f := range files {
					fileBuf := make([]byte, f.Size)

					fileData, _ := f.Open()
					fileData.Read(fileBuf)
					fileData.Close()

					pbFiles.Contents = append(pbFiles.Contents, &types.FileContent{
						Name:    f.Filename,
						Content: fileBuf,
					})
				}

				pb = pbFiles
			} else {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					deferredError = util.ErrCheck(err)
					return
				}
				defer req.Body.Close()

				if len(body) > 0 {
					err = json.Unmarshal(body, pb)
					if err != nil {
						deferredError = util.ErrCheck(err)
						return
					}
				}
			}

			pbVal = reflect.ValueOf(pb).Elem()

			// Parse query and path parameters

			util.ParseProtoQueryParams(pbVal, req.URL.Query())
			util.ParseProtoPathParams(
				pbVal,
				strings.Split(handlerOpts.ServiceMethodURL, "/"),
				strings.Split(strings.TrimPrefix(req.URL.Path, "/api"), "/"),
			)

			// Perform the handler function

			results := handlerFunc.Call([]reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(req),
				reflect.ValueOf(pb),
				reflect.ValueOf(session),
				reflect.ValueOf(tx),
			})

			// Handle errors

			if len(results) != 2 {
				deferredError = util.ErrCheck(errors.New("badly formed handler"))
				return
			}

			if err, ok := results[1].Interface().(error); ok {
				deferredError = err
				return
			}

			// Transform the response

			if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "files/content") {
				w.Header().Add("Content-Type", "application/octet-stream")
				_, err := w.Write(*results[0].Interface().(*[]byte))
				if err != nil {
					deferredError = util.ErrCheck(err)
					return
				}
			} else {
				pbJsonBytes, err := protojson.Marshal(results[0].Interface().(protoreflect.ProtoMessage))
				if err != nil {
					deferredError = util.ErrCheck(err)
					return
				}

				defer exeTimeDefer("response len " + fmt.Sprint(len(pbJsonBytes)))

				w.Write(pbJsonBytes)
			}

		})

		middlewareHandler := ApplyMiddleware(protoHandler, []Middleware{
			a.CacheMiddleware(handlerOpts),
			a.SiteRoleCheckMiddleware(handlerOpts),
			a.CorsMiddleware,
			// a.SessionAuthMiddleware,
		})

		mux.HandleFunc(handlerOpts.Pattern, middlewareHandler)
	}
}
