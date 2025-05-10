package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const (
	PING                    = "PING"
	PONG                    = "PONG"
	socketEventTimeoutAfter = 150 * time.Millisecond
)

var (
	pingBytes            = util.GenerateMessage(util.DefaultPadding, &types.SocketMessage{Payload: PING})
	pings                sync.Map // map[string]time.Time
	socketPingTicker     *time.Ticker
	socketRequestTimeout = errors.New("socket request timed out")
	sockRl               = NewRateLimit("sock", 30, 10, time.Duration(time.Minute))
)

// The sock handler takes a typical HTTP request and upgrades it to a websocket connection
// The connection is identified via the ticket, setup in redis and the DB
// The handler then becomes a long lived process where messages are read from the connection
// and then parsed into an expected format, then sent through the websocket event router
func (a *API) InitSockServer() {
	a.Server.Handler.(*http.ServeMux).HandleFunc("GET /sock", func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Connection"), "Upgrade") || req.Header.Get("Upgrade") != "websocket" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("Bad Request: Expected WebSocket"))
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
			}
			return
		}

		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Sec-WebSocket-Accept", util.ComputeWebSocketAcceptKey(req.Header.Get("Sec-WebSocket-Key")))
		w.WriteHeader(http.StatusSwitchingProtocols)

		hijacker, ok := w.(*responseCodeWriter).ResponseWriter.(http.Hijacker)
		if !ok {
			http.Error(w, "Websocket upgrade failed", http.StatusInternalServerError)
			return
		}

		conn, bufrw, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, "Websocket upgrade failed", http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		defer bufrw.Flush()

		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		err = conn.SetWriteDeadline(time.Time{})
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		ticket := req.URL.Query().Get("ticket")
		if ticket == "" {
			return
		}

		_, connId, err := util.SplitColonJoined(ticket)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		connCtx, cancelConnCtx := context.WithDeadline(context.Background(), time.Now().Add(socketEventTimeoutAfter))
		userSession, err := a.Handlers.Socket.StoreConn(connCtx, ticket, conn)
		if err != nil {
			cancelConnCtx()
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}
		cancelConnCtx()

		socketId := util.GetColonJoined(userSession.UserSub, connId)

		// Use wait group/goroutine to clean up DbSession usage as this overall func is long-lived
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()

			ds := clients.DbSession{
				Pool:        a.Handlers.Database.Client().Pool,
				UserSession: userSession,
			}

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketEventTimeoutAfter))
			defer cancel()
			err = ds.InitDbSocketConnection(ctx, connId)
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}
		}()

		wg.Wait()

		defer func() {
			ds := clients.DbSession{
				Pool:        a.Handlers.Database.DatabaseClient.Pool,
				UserSession: userSession,
			}

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketEventTimeoutAfter))
			defer cancel()
			a.TearDownSocketConnection(ctx, socketId, connId, ds)
		}()

		err = a.Handlers.Redis.InitRedisSocketConnection(req.Context(), socketId)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		var userSockInfo strings.Builder
		userSockInfo.WriteString(" /sock sub:")
		userSockInfo.WriteString(userSession.UserSub)
		userSockInfo.WriteString(" cid:")
		userSockInfo.WriteString(connId)
		userSockInfo.WriteString(" ")

		var joinSockMessage strings.Builder
		joinSockMessage.WriteString("JOIN")
		joinSockMessage.WriteString(userSockInfo.String())
		util.SockLog.Println(joinSockMessage.String())

		var partSockMessage strings.Builder
		partSockMessage.WriteString("PART")
		partSockMessage.WriteString(userSockInfo.String())

		defer func() {
			util.SockLog.Println(partSockMessage.String())
		}()

		messages := make(chan []byte)
		errs := make(chan error)

		go func() {
			for {
				data, err := util.ReadSocketConnectionMessage(conn)
				if err != nil {
					if io.EOF == err {
						return
					}
					errs <- err
				} else {
					messages <- data
				}
			}
		}()

		pings.Store(connId, time.Now())

		pingTimer := time.NewTicker(30 * time.Second)
		errorFlag := false

		for {
			select {
			case <-req.Context().Done():
				partSockMessage.WriteString("sock context ended")
				return
			case <-pingTimer.C:
				// let errors log again
				if errorFlag {
					errorFlag = false
				}

				// Close connections not responding to pings
				pingVal, _ := pings.Load(connId)
				if lastSeen, ok := pingVal.(time.Time); ok && time.Since(lastSeen) > 90*time.Second {
					partSockMessage.WriteString("returned due to ping timeout")
					return
				}

				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketEventTimeoutAfter))
				defer cancel()
				err := a.Handlers.Socket.SendMessage(ctx, userSession.UserSub, connId, &types.SocketMessage{
					Action:  socketActionPingPong,
					Payload: PING,
				})
				if err != nil {
					partSockMessage.WriteString("returned due to ping error ")
					partSockMessage.WriteString(err.Error())
					return
				}
			case data := <-messages:
				if len(data) == 2 {
					// EOF
					partSockMessage.WriteString("returned due to messages EOF")
					return
				} else {
					if sockRl.Limit(userSession.UserSub) {
						continue
					}

					go func() {
						ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketEventTimeoutAfter))

						defer func() {
							clients.GetGlobalWorkerPool().CleanUpClientMapping(userSession.UserSub)
							cancel()
						}()

						smResultChan := make(chan *types.SocketMessage, 1)

						go func() {
							smResultChan <- a.SocketMessageReceiver(data)
						}()

						select {
						case <-ctx.Done():
							errs <- socketRequestTimeout
							return
						case sm := <-smResultChan:
							if sm == nil {
								return
							}

							if sm.Action == socketActionPingPong {
								pings.Store(connId, time.Now())
								return
							}

							a.SocketMessageRouter(ctx, connId, socketId, sm, clients.DbSession{
								Topic:       sm.Topic,
								Pool:        a.Handlers.Database.DatabaseClient.Pool,
								UserSession: userSession,
							})
						}
					}()
				}
			case err := <-errs:
				// normally shouldn't error, but if it does it could potentially be heavy load
				// so only err log once every ping timer
				if errorFlag {
					continue
				}
				errStr := err.Error()

				if strings.Index(errStr, "connection reset by peer") != -1 {
					partSockMessage.WriteString(errStr)
					return
				} else {
					util.ErrorLog.Println(util.ErrCheck(errors.New("sock read error " + connId + errStr)))
				}

				errorFlag = true
			}
		}
	})
}

func (a *API) TearDownSocketConnection(ctx context.Context, socketId, connId string, ds clients.DbSession) {
	var tearDownFailures strings.Builder
	var wg sync.WaitGroup
	// Log errors but attempt to tear down everything

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx := context.Background()

		topics, err := a.Handlers.Redis.HandleUnsub(ctx, socketId)
		if err != nil {
			tearDownFailures.WriteString("handle_unsub ")
			return
		}

		for topic, targets := range topics {
			err := a.Handlers.Socket.SendMessage(context.Background(), ds.UserSession.UserSub, targets, &types.SocketMessage{
				Action:  socketActionUnsubscribeTopic,
				Topic:   topic,
				Payload: socketId,
			})
			if err != nil {
				continue
			}

			err = a.Handlers.Redis.RemoveTopicFromConnection(ctx, socketId, topic)
			if err != nil {
				tearDownFailures.WriteString("rem_conn_topic ")
				tearDownFailures.WriteString(topic)
				tearDownFailures.WriteString(" ")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		workerDs := clients.DbSession{
			Pool: a.Handlers.Database.DatabaseClient.Pool,
			UserSession: &types.UserSession{
				UserSub: "worker",
			},
		}

		err := workerDs.RemoveDbSocketConnection(ctx, connId)
		if err != nil {
			tearDownFailures.WriteString("rem_sock_conn ")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := a.Handlers.Socket.SendCommand(context.Background(), clients.DeleteSocketConnectionSocketCommand, &types.SocketRequestParams{
			UserSub: ds.UserSession.UserSub,
			ConnId:  connId,
		})
		if err != nil {
			tearDownFailures.WriteString("del_cmd ")
		}
	}()

	wg.Wait()

	if tearDownFailures.Len() > 0 {
		var sb strings.Builder
		sb.WriteString("TEARDOWN /sock sub:")
		sb.WriteString(ds.UserSession.UserSub)
		sb.WriteString(" socketid:")
		sb.WriteString(socketId)
		sb.WriteString(" connid:")
		sb.WriteString(connId)
		sb.WriteString(" ")
		sb.WriteString(tearDownFailures.String())
		util.ErrorLog.Println(sb.String())
	}

	clients.GetGlobalWorkerPool().CleanUpClientMapping(ds.UserSession.UserSub)
}
