package api

import (
	"bufio"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var DEFAULT_UNIX_IGNORED_PROTO_FIELDS = []protoreflect.Name{
	protoreflect.Name("firstName"),
	protoreflect.Name("lastName"),
}

type UnixResponseWriter struct {
	body       []byte
	statusCode int
	header     http.Header
}

func NewUnixResponseWriter() *UnixResponseWriter {
	return &UnixResponseWriter{
		header: http.Header{},
	}
}

func (w *UnixResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (w *UnixResponseWriter) Header() http.Header {
	return w.header
}

func (w *UnixResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (a *API) InitUnixServer(unixPath string) {

	DEFAULT_UNIX_IGNORED_PROTO_FIELDS = append(DEFAULT_UNIX_IGNORED_PROTO_FIELDS, util.DEFAULT_IGNORED_PROTO_FIELDS...)

	_, err := os.Stat(unixPath)
	if err == nil {
		err = os.Remove(unixPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	unixListener, err := net.Listen("unix", unixPath)
	if err != nil {
		log.Fatal(err)
	}

	defer unixListener.Close()

	util.DebugLog.Println("Listening on", unixPath)

	for {
		conn, err := unixListener.Accept()
		if err != nil {
			util.ErrorLog.Println("Error accepting unix connection:", err)
			continue
		}

		a.HandleUnixConnection(conn)
	}
}

func (a *API) HandleUnixConnection(conn net.Conn) {
	defer conn.Close()

	var deferredError error
	var authEvent *types.AuthEvent
	var authResponseValue string

	defer func() {
		reqId := uuid.NewString()
		var sb strings.Builder
		sb.WriteString("Backchannel: ")
		sb.WriteString(reqId)
		sb.WriteByte(' ')
		sb.WriteString(authEvent.WebhookName)
		sb.WriteByte(' ')
		if authEvent == nil {
			sb.WriteString(" AUTH_EVENT_NIL ")
		} else if authEvent.GetUserId() != "" {
			clients.GetGlobalWorkerPool().CleanUpClientMapping(authEvent.GetUserId())
		}

		if p := recover(); p != nil {
			sb.WriteString(" PANIC ")
			sb.WriteString(fmt.Sprint(p))

			_, err := fmt.Fprint(conn, `{ "success": false, "reason": "UNIX_PANIC" }`)
			if err != nil {
				sb.WriteString(" PANIC_REPLY_ERROR ")
				sb.WriteString(err.Error())
			}
		}

		if deferredError != nil {
			sb.WriteString(" DEFERRED_ERROR ")
			sb.WriteString(deferredError.Error())
		}

		if authEvent != nil {
			util.MaskNologFields(authEvent, DEFAULT_UNIX_IGNORED_PROTO_FIELDS)
			sb.WriteString(" Event: ")
			authEventBytes, err := json.Marshal(authEvent)
			if err != nil {
				sb.WriteString(" NO_EVENT_BYTES ")
			} else {
				sb.WriteString(string(authEventBytes))
			}

			if authResponseValue != "" {
				sb.WriteString(" Response: ")
				sb.WriteString(authResponseValue)
			} else {
				sb.WriteString(" NO_RESPONSE ")
			}

			sb.WriteString(" CLIENT ")
			sb.WriteString(util.AnonIp(authEvent.GetIpAddress()))
		} else {
			sb.WriteString(" NO_EVENT ")
		}

		util.AuthLog.Println(sb.String())
	}()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &authEvent)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}
	}

	// Create fake context so we can use our regular http handlers
	fakeReq, err := http.NewRequest("GET", "unix://auth", nil)
	if err != nil {
		deferredError = util.ErrCheck(err)
		return
	}

	session := types.NewConcurrentUserSession(&types.UserSession{
		UserSub:   authEvent.UserId,
		UserEmail: authEvent.Email,
		AnonIp:    util.AnonIp(authEvent.IpAddress),
		Timezone:  authEvent.Timezone,
	})

	ctx, cancel := context.WithTimeout(fakeReq.Context(), 5*time.Second)
	defer cancel()

	reqInfo := handlers.ReqInfo{
		Ctx:     ctx,
		W:       nil,
		Req:     fakeReq,
		Session: session,
	}

	result, err := a.Handlers.Functions["AuthWebhook_"+authEvent.WebhookName](reqInfo, authEvent)
	if err != nil {
		deferredError = util.ErrCheck(err)
		return
	}

	authResponse, ok := result.(*types.AuthWebhookResponse)
	if !ok {
		deferredError = util.ErrCheck(errors.New("bad auth response object during backchannel"))
		return
	}

	authResponseValue = authResponse.GetValue()

	fmt.Fprint(conn, authResponseValue)
}
