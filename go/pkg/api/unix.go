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
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

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
		} else if deferredError != nil {
			util.ErrorLog.Println(deferredError)
			fmt.Fprint(conn, fmt.Sprintf(`{ "success": false, "reason": "%s" }`, deferredError))
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := a.Handlers.Database.DatabaseClient.Pool.Begin(ctx)
	if err != nil {
		deferredError = util.ErrCheck(err)
		return
	}
	defer tx.Rollback(ctx)

	poolTx := &clients.PoolTx{
		Tx: tx,
	}

	workerSession := types.NewConcurrentUserSession(&types.UserSession{
		UserSub: "worker",
	})

	err = poolTx.SetSession(ctx, workerSession)
	if err != nil {
		deferredError = util.ErrCheck(err)
		return
	}

	reqInfo := handlers.ReqInfo{
		Ctx:     fakeReq.Context(),
		W:       nil,
		Req:     fakeReq,
		Session: session,
		Tx:      poolTx,
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

	err = poolTx.UnsetSession(ctx)
	if err != nil {
		deferredError = util.ErrCheck(err)
		return
	}

	err = poolTx.Commit(ctx)
	if err != nil {
		deferredError = util.ErrCheck(err)
		return
	}

	fmt.Fprint(conn, authResponse.Value)
}
