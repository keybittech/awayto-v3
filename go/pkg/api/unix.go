package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (a *API) InitUnixServer(unixPath string) {
	os.Remove(unixPath)

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

		go a.HandleUnixConnection(conn)
	}
}

func (a *API) HandleUnixConnection(conn net.Conn) {
	defer conn.Close()

	var authEvent *types.AuthEvent

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		json.Unmarshal(scanner.Bytes(), &authEvent)
	}

	handler := reflect.ValueOf(a.Handlers).MethodByName("AuthWebhook_" + authEvent.WebhookName)

	if handler.IsValid() {
		var deferredError error

		// Create fake context so we can use our regular http handlers
		fakeReq, err := http.NewRequest("GET", "unix://auth", nil)
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}

		session := &types.UserSession{
			UserSub:   authEvent.UserId,
			UserEmail: authEvent.Email,
			AnonIp:    util.AnonIp(authEvent.IpAddress),
			Timezone:  authEvent.Timezone,
		}

		defer func() {
			if p := recover(); p != nil {
				panic(p)
			} else if deferredError != nil {
				util.ErrorLog.Println(deferredError)
				fmt.Fprint(conn, fmt.Sprintf(`{ "success": false, "reason": "%s" }`, deferredError))
			}
		}()

		results := []reflect.Value{}
		err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
			results = handler.Call([]reflect.Value{
				reflect.ValueOf(fakeReq),
				reflect.ValueOf(authEvent),
				reflect.ValueOf(session),
				reflect.ValueOf(tx),
			})
			return nil
		}, "worker", "", "")
		if err != nil {
			deferredError = util.ErrCheck(err)
			return
		}

		if len(results) != 2 {
			deferredError = util.ErrCheck(errors.New("incorrectly structured auth webhook: " + authEvent.WebhookName))
			return
		}

		if !results[1].IsNil() {
			deferredError = util.ErrCheck(results[1].Interface().(error))
			return
		}

		resStr := results[0].Interface().(string)

		fmt.Fprint(conn, resStr)
	}
}
