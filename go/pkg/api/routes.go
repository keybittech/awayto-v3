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

			exeTimeDefer := util.ExeTime(handlerOpts.Pattern)

			session, err := a.GetAuthorizedSession(req)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				http.Error(w, util.ForbiddenResponse, http.StatusForbidden)
				return
			}

			requestId := uuid.NewString()

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
					util.ErrorLog.Println(util.ErrCheck(err))
					http.Error(w, "could not read proto handler body", http.StatusUnprocessableEntity)
					return
				}
				defer req.Body.Close()

				if len(body) > 0 {
					err = json.Unmarshal(body, pb)
					if err != nil {
						util.ErrorLog.Println(util.ErrCheck(err))
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

			var reqErr error

			tx, err := a.Handlers.Database.Client().Begin()
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			// SET doesn't support parameters, and UserSub is sourced from the auth token for safety
			_, err = tx.Exec(fmt.Sprintf("SET SESSION app_session.user_sub = '%s'", session.UserSub))
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}
			if session.GroupId != "" {
				_, err = tx.Exec(fmt.Sprintf("SET SESSION app_session.group_id = '%s'", session.GroupId))
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
					return
				}
			}

			defer func() {
				_, err = tx.Exec(`SET SESSION app_session.user_sub = ''`)
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(reqErr))
				}
				if session.GroupId != "" {
					_, err = tx.Exec(`SET SESSION app_session.group_id = ''`)
					if err != nil {
						util.ErrorLog.Println(util.ErrCheck(reqErr))
					}
				}

				if p := recover(); p != nil {
					tx.Rollback()
					panic(p)
				} else if reqErr != nil {
					tx.Rollback()
				} else {
					reqErr = tx.Commit()
					if reqErr != nil {
						util.ErrorLog.Println(util.ErrCheck(reqErr))
					}
				}
			}()

			results := handlerFunc.Call([]reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(req),
				reflect.ValueOf(pb),
				reflect.ValueOf(session),
				reflect.ValueOf(tx),
			})

			if len(results) != 2 || !results[1].IsNil() {
				if len(results) != 2 {
					util.ErrorLog.Println(util.ErrCheck(errors.New(fmt.Sprintf("bad api result for %s", pbVal.Type().Name()))))
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

				loggedErr := errors.New(fmt.Sprintf("\n  RequestId: %s\n  Error: %s\n  Params:%s\n", requestId, errStr, reqParams))

				util.ErrorLog.Println(loggedErr)

				var errRes string

				if strings.Contains(errStr, util.ErrorForUser) {
					errRes = fmt.Sprintf("Request Id: %s\n%s", requestId, util.SnipUserError(errStr))
				} else {
					errRes = fmt.Sprintf("Request Id: %s\nAn error occurred. Please try again later or contact your administrator with the request id provided.", requestId)
				}

				reqErr = loggedErr
				http.Error(w, errRes, http.StatusInternalServerError)
				return
			}

			if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "files/content") {
				w.Header().Add("Content-Type", "application/octet-stream")
				w.Write(*results[0].Interface().(*[]byte))
			} else {
				pbJsonBytes, err := protojson.Marshal(results[0].Interface().(protoreflect.ProtoMessage))
				if err != nil {
					reqErr = err
					util.ErrorLog.Println(util.ErrCheck(err))
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
			a.CorsMiddleware,
			// a.SessionAuthMiddleware,
		})

		mux.HandleFunc(handlerOpts.Pattern, middlewareHandler)
	}
}
