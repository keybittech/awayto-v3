package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
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

		var wg sync.WaitGroup
		wg.Add(1)
		a.HandleUnixConnection(&wg, conn)
		wg.Wait()
	}
}

func (a *API) HandleUnixConnection(wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	defer conn.Close()

	var authEvent *types.AuthEvent

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		json.Unmarshal(scanner.Bytes(), &authEvent)
	}

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
			util.ErrorLog.Println(fmt.Sprint(p))
		} else if deferredError != nil {
			util.ErrorLog.Println(deferredError)
			fmt.Fprint(conn, fmt.Sprintf(`{ "success": false, "reason": "%s" }`, deferredError))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := a.Handlers.Database.DatabaseClient.Pool.Begin(ctx)
	if err != nil {
		deferredError = util.ErrCheck(err)
		return
	}

	poolTx := &clients.PoolTx{
		Tx: tx,
	}

	workerSession := &types.UserSession{
		UserSub: "worker",
	}

	poolTx.SetSession(ctx, workerSession)

	reqInfo := handlers.ReqInfo{
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

	poolTx.UnsetSession(ctx)
	poolTx.Commit(ctx)

	fmt.Fprint(conn, authResponse.Value)
}
