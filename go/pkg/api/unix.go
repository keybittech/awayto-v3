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
		var deferralError error

		// Create fake context so we can use our regular http handlers
		fakeReq, err := http.NewRequest("GET", "unix://auth", nil)
		if err != nil {
			deferralError = util.ErrCheck(err)
			return
		}

		session := &clients.UserSession{
			UserSub:   authEvent.UserId,
			UserEmail: authEvent.Email,
			AnonIp:    util.AnonIp(authEvent.IpAddress),
			Timezone:  authEvent.Timezone,
		}

		tx, err := a.Handlers.Database.Client().Begin()
		if err != nil {
			deferralError = util.ErrCheck(err)
			return
		}

		err = tx.SetDbVar("user_sub", "worker")
		if err != nil {
			deferralError = util.ErrCheck(err)
			return
		}

		err = tx.SetDbVar("group_id", "")
		if err != nil {
			deferralError = util.ErrCheck(err)
			return
		}

		defer func() {
			var deferredError error

			err = tx.SetDbVar("user_sub", "")
			if err != nil {
				deferredError = util.ErrCheck(err)
			}

			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			} else if deferralError != nil {
				tx.Rollback()
			} else {
				err = tx.Commit()
				if err != nil {
					deferralError = util.ErrCheck(err)
				}
			}

			var loggedError string
			if deferredError != nil {
				loggedError = deferredError.Error()
			}
			if deferralError != nil {
				loggedError = fmt.Sprintf("%s %s", loggedError, deferralError.Error())
			}
			if loggedError != "" {
				util.ErrorLog.Println(loggedError)
				fmt.Fprint(conn, fmt.Sprintf(`{ "success": false, "reason": "%s" }`, loggedError))
			}
		}()

		results := handler.Call([]reflect.Value{
			reflect.ValueOf(fakeReq),
			reflect.ValueOf(authEvent),
			reflect.ValueOf(session),
			reflect.ValueOf(tx),
		})

		if len(results) != 2 {
			deferralError = util.ErrCheck(errors.New("incorrectly structured auth webhook: " + authEvent.WebhookName))
			return
		}

		if !results[1].IsNil() {
			deferralError = util.ErrCheck(results[1].Interface().(error))
			return
		}

		resStr := results[0].Interface().(string)

		fmt.Fprint(conn, resStr)
	}
}
