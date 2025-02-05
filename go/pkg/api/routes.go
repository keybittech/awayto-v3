package api

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"context"
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

		serviceMethod.Input().FullName()

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
			exeTimeDefer := util.ExeTime(handlerOpts.Pattern)

			requestId := uuid.NewString()

			ctx := context.WithValue(req.Context(), "LogId", requestId)

			req = req.WithContext(ctx)

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
					util.ErrCheck(err)
					http.Error(w, "could not read proto handler body", http.StatusUnprocessableEntity)
					return
				}
				defer req.Body.Close()

				if len(body) > 0 {
					err = json.Unmarshal(body, pb)
					if err != nil {
						util.ErrCheck(err)
						http.Error(w, "could not unmarshal proto handler body", http.StatusUnprocessableEntity)
						return
					}
				}
			}

			pbVal := reflect.ValueOf(pb).Elem()

			util.ParseProtoQueryParams(pbVal, req.URL.Query())
			util.ParseProtoPathParams(
				pbVal,
				strings.Split(handlerOpts.ServiceMethodURL, "/"),
				strings.Split(strings.TrimPrefix(req.URL.Path, "/api"), "/"),
			)

			results := handlerFunc.Call([]reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(req),
				reflect.ValueOf(pb),
			})

			if len(results) != 2 || !results[1].IsNil() {
				if len(results) != 2 {
					util.ErrCheck(errors.New(fmt.Sprintf("bad api result for %s", pbVal.Type().Name())))
				}

				var reqParams string
				pbValType := pbVal.Type()
				for i = 0; i < pbVal.NumField(); i++ {
					field := pbVal.Field(i)

					fName := pbValType.Field(i).Name

					if !slices.Contains(ignoreFields, fName) {
						reqParams += " " + fmt.Sprintf("%s=%v", fName, field.Interface())
					}
				}

				errStr := results[1].Interface().(error).Error()

				loggedErr := errors.New(fmt.Sprintf("RequestId: %s Error: %s Params:%s", requestId, errStr, reqParams))

				util.ErrorLog.Println(loggedErr)

				if *util.DebugMode {
					fmt.Println(fmt.Sprintf("DEBUG: %s", loggedErr))
				}

				var errRes string

				if strings.Contains(errStr, util.ErrorForUser) {
					errRes = fmt.Sprintf("Request Id: %s\n%s", requestId, util.SnipUserError(errStr))
				} else {
					errRes = fmt.Sprintf("Request Id: %s\nAn error occurred. Please try again later or contact your administrator with the request id provided.", requestId)
				}

				http.Error(w, errRes, http.StatusInternalServerError)
				return
			}

			if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "files/content") {
				w.Header().Add("Content-Type", "application/octet-stream")
				w.Write(*results[0].Interface().(*[]byte))
			} else {
				pbJsonBytes, err := protojson.Marshal(results[0].Interface().(protoreflect.ProtoMessage))
				if err != nil {
					util.ErrCheck(err)
					http.Error(w, "Response parse failure", http.StatusInternalServerError)
					return
				}

				defer exeTimeDefer("response len " + fmt.Sprint(len(pbJsonBytes)))

				w.Write(pbJsonBytes)
			}

		})

		middlewareHandler := ApplyMiddleware(protoHandler, []Middleware{
			a.CacheMiddleware(handlerOpts),
			a.SiteRoleCheckMiddleware(handlerOpts),
			a.SessionAuthMiddleware,
			a.CorsMiddleware,
		})

		mux.HandleFunc(handlerOpts.Pattern, middlewareHandler)
	}
}
