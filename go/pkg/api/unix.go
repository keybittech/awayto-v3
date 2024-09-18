package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/util"
	"bufio"
	"context"
	"encoding/json"
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

		bgContext := context.Background()
		bgContext = context.WithValue(bgContext, "SourceIp", authEvent.IpAddress)
		bgContext = context.WithValue(bgContext, "UserSession", &clients.UserSession{UserSub: authEvent.UserId})

		fakeReq = fakeReq.WithContext(bgContext)

		handler.Call([]reflect.Value{
			reflect.ValueOf(fakeReq),
			reflect.ValueOf(authEvent),
		})

		fmt.Fprint(conn, `{ "success": true }`)
	}
}
