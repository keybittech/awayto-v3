package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/util"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
)

func (a *API) InitUnixServer(unixPath string) {
	os.Remove(unixPath)

	unixListener, err := net.Listen("unix", unixPath)
	if err != nil {
		panic(err)
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

	var authEvent clients.AuthEvent

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		json.Unmarshal(scanner.Bytes(), &authEvent)
	}

	handler := reflect.ValueOf(a.Handlers).MethodByName("AuthWebhook_" + authEvent.WebhookName)

	if handler.IsValid() {

		// Create fake context so we can use our regular http handlers
		fakeReq, err := http.NewRequest("GET", "unix://auth", nil)
		if err != nil {
			util.ErrCheck(err)
			return
		}

		session := &clients.UserSession{
			UserSub:   authEvent.UserId,
			UserEmail: authEvent.Email,
			AnonIp:    util.AnonIp(authEvent.IpAddress),
		}

		tx, err := a.Handlers.Database.Client().Begin()
		if err != nil {
			util.ErrCheck(err)
			return
		}

		var reqErr error

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			} else if reqErr != nil {
				tx.Rollback()
			} else {
				util.ErrCheck(tx.Commit())
			}
		}()

		results := handler.Call([]reflect.Value{
			reflect.ValueOf(fakeReq),
			reflect.ValueOf(authEvent),
			reflect.ValueOf(session),
			reflect.ValueOf(tx),
		})

		if len(results) != 2 {
			reqErr = errors.New("incorrectly structured auth webhook: " + authEvent.WebhookName)
			util.ErrorLog.Println(util.ErrCheck(reqErr))
			return
		}

		if !results[1].IsNil() {
			reqErr = results[1].Interface().(error)
			util.ErrorLog.Println(util.ErrCheck(reqErr))
			return
		}

		resStr := results[0].Interface().(string)

		fmt.Fprint(conn, resStr)
	}
}
