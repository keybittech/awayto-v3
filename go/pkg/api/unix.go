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
	"runtime/debug"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

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
	err := os.Remove(unixPath)
	if err != nil {
		log.Fatal(err)
	}

	unixListener, err := net.Listen("unix", unixPath)
	if err != nil {
		log.Fatal(err)
	}

	defer unixListener.Close()

	fmt.Println("Listening on", unixPath)

	for {
		conn, err := unixListener.Accept()
		if err != nil {
			fmt.Println("Error accepting unix connection:", err)
			continue
		}

		a.HandleUnixConnection(conn)
	}
}

func (a *API) HandleUnixConnection(conn net.Conn) {
	var deferredError error
	var authEvent *types.AuthEvent

	defer conn.Close()

	defer func() {
		if p := recover(); p != nil {
			util.ErrorLog.Println(fmt.Sprint(p))
		}
	}()

	w := NewUnixResponseWriter()

	defer func() {
		if p := recover(); p != nil {
			// ErrCheck handled errors will already have a trace
			// If there is no trace file hint, print the full stack
			var sb strings.Builder
			if authEvent != nil {
				sb.WriteString("Backchannel: ")
				sb.WriteString(authEvent.WebhookName)
				sb.WriteByte(' ')
				sb.WriteString(fmt.Sprint(p))
			}
			errStr := sb.String()

			util.RequestError(w, errStr, util.DEFAULT_IGNORED_PROTO_FIELDS, authEvent)
			fmt.Fprint(conn, `{ "success": false, "reason": "UNIX_PANIC" }`)

			if !strings.Contains(errStr, ".go:") {
				util.ErrorLog.Println(string(debug.Stack()))
			}
		}

		if deferredError != nil {
			util.ErrorLog.Println(deferredError)
		}

		if authEvent != nil && authEvent.GetUserId() != "" {
			clients.GetGlobalWorkerPool().CleanUpClientMapping(authEvent.GetUserId())
		}

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

	fmt.Fprint(conn, authResponse.GetValue())
}
