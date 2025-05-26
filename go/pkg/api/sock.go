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
	"golang.org/x/time/rate"
)

const (
	pingTimeout             = 30 * time.Second
	socketEventTimeoutAfter = 150 * time.Millisecond
	socketCleanupTimeout    = 5 * time.Second
	maxUserConns            = 3
	messagesPerSecond       = 20
	messagesBurst           = 10
)

var (
	connLimit          = make(chan struct{}, 500)
	userConns          sync.Map // map[string]int
	pings              sync.Map // map[string]time.Time
	socketPingTicker   *time.Ticker
	socketEventTimeout = errors.New("socket request timed out")
	excessiveMessages  = errors.New("dropping excessive message")
	nilMessageBytes    = errors.New("sock message bytes nil with no error")
	sockRl             = NewRateLimit("sock", rate.Limit(messagesPerSecond), messagesBurst, time.Duration(time.Minute))
	pingMsg            = &types.SocketMessage{
		Action:  types.SocketActions_PING_PONG,
		Payload: "PING",
	}
)

// The sock handler takes a typical HTTP request and upgrades it to a websocket connection
// The connection is identified via the ticket, setup in redis and the DB
// The handler then becomes a long lived process where messages are read from the connection
// and then parsed into an expected format, then sent through the websocket event router
func (a *API) InitSockServer() {
	a.Server.Handler.(*http.ServeMux).HandleFunc("GET /sock", func(w http.ResponseWriter, req *http.Request) {
		// Limit total
		select {
		case connLimit <- struct{}{}:
			defer func() { <-connLimit }()
		default:
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

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

		if req.URL.Query().Get("ticket") == "" {
			return
		}

		_, connId, err := util.SplitColonJoined(req.URL.Query().Get("ticket"))
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		// Use the ticket against the socket client to get session details
		var session *types.ConcurrentUserSession
		{
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketEventTimeoutAfter))
			session, err = a.Handlers.Socket.StoreConn(ctx, req.URL.Query().Get("ticket"), conn)
			if err != nil {
				cancel()
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}
			cancel()
		}

		userSub := session.GetUserSub()
		groupId := session.GetGroupId()

		// Cleanup socket client connection mapping
		defer func() {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketCleanupTimeout))
			defer cancel()
			if _, err := a.Handlers.Socket.SendCommand(ctx, clients.DeleteSocketConnectionSocketCommand, &types.SocketRequestParams{
				UserSub: userSub,
				GroupId: groupId,
				ConnId:  connId,
			}); err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
			}
		}()

		// Make sure user doesn't have too many connections
		// A new browser tab is a new connection, for example
		current, _ := userConns.LoadOrStore(userSub, 0)
		if current.(int) >= maxUserConns {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		// Add a connection for the user
		userConns.Store(userSub, current.(int)+1)

		// Decrement or remove the connection on cleanup
		defer func() {
			if curr, ok := userConns.Load(userSub); ok {
				if newCount := curr.(int) - 1; newCount <= 0 {
					userConns.Delete(userSub)
				} else {
					userConns.Store(userSub, newCount)
				}
			}
		}()

		socketId := util.GetColonJoined(userSub, connId)

		// Initialize first ping time
		pings.Store(connId, time.Now())

		// Cleanup database and redis data on close
		defer func() {
			pings.Delete(connId)
			a.TearDownSocketConnection(socketId, connId, clients.DbSession{
				Pool:                  a.Handlers.Database.DatabaseClient.Pool,
				ConcurrentUserSession: session,
			})
		}()

		// Initialize database and redis info for the connection
		{
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketCleanupTimeout))

			if err := (clients.DbSession{
				Pool:                  a.Handlers.Database.DatabaseClient.Pool,
				ConcurrentUserSession: session,
			}).InitDbSocketConnection(ctx, connId); err != nil {
				cancel()
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			if err := a.Handlers.Redis.InitRedisSocketConnection(ctx, socketId); err != nil {
				cancel()
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			cancel()
		}

		// Sock connection is setup and should be logged
		var userSockInfo strings.Builder
		userSockInfo.WriteString(" /sock sub:")
		userSockInfo.WriteString(userSub)
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

		messages := make(chan []byte, messagesPerSecond+messagesBurst)
		errs := make(chan error)

		// Start listening for messages over the socket
		go func() {
			for {
				if sockRl.Limit(userSub) {
					continue
				}

				data, err := util.ReadSocketConnectionMessage(conn)
				if err != nil {
					if io.EOF == err {
						return
					}
					errs <- err
					continue
				}

				if data == nil {
					// Shouldn't get here based on read error handling
					errs <- nilMessageBytes
					continue
				}

				// Put the data in the messages queue as long as there's space
				select {
				case messages <- data:
				case <-time.After(socketEventTimeoutAfter):
					errs <- socketEventTimeout
					continue
				default:
					errs <- excessiveMessages
					continue
				}
			}
		}()

		pingTimer := time.NewTicker(pingTimeout)
		defer pingTimer.Stop()
		errorFlag := false

		// Handle messages which were successfully parsed from the socket
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
				err := a.Handlers.Socket.SendMessage(ctx, userSub, connId, pingMsg)
				if err != nil {
					cancel()
					partSockMessage.WriteString("returned due to ping error ")
					partSockMessage.WriteString(err.Error())
					return
				}
				cancel()
			case data := <-messages:
				if len(data) == 2 {
					// EOF
					partSockMessage.WriteString("returned due to messages EOF")
					return
				} else {
					// handle received message bytes
					ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketEventTimeoutAfter))

					smResultChan := make(chan *types.SocketMessage, 1)

					go func() {
						smResultChan <- a.SocketMessageReceiver(data)
					}()

					select {
					case <-ctx.Done():
						cancel()
						errs <- socketEventTimeout
						return
					case sm := <-smResultChan:
						if sm.Action == types.SocketActions_PING_PONG {
							pings.Store(connId, time.Now())
						} else {
							a.SocketMessageRouter(ctx, connId, socketId, sm, clients.DbSession{
								Topic:                 sm.Topic,
								Pool:                  a.Handlers.Database.DatabaseClient.Pool,
								ConcurrentUserSession: session,
							})
							clients.GetGlobalWorkerPool().CleanUpClientMapping(userSub)
						}
						cancel()
					}
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

func (a *API) TearDownSocketConnection(socketId, connId string, ds clients.DbSession) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(socketCleanupTimeout))
	defer cancel()

	userSub := ds.ConcurrentUserSession.GetUserSub()

	var tearDownFailures strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		topics, err := a.Handlers.Redis.HandleUnsub(ctx, socketId)
		if err != nil {
			tearDownFailures.WriteString("handle_unsub ")
			return
		}

		for topic, targets := range topics {
			err := a.Handlers.Socket.SendMessage(ctx, userSub, targets, &types.SocketMessage{
				Action:  types.SocketActions_UNSUBSCRIBE_TOPIC,
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

	go func() {
		defer wg.Done()
		if err := (clients.DbSession{
			Pool: a.Handlers.Database.DatabaseClient.Pool,
			ConcurrentUserSession: types.NewConcurrentUserSession(&types.UserSession{
				UserSub: "worker",
			}),
		}).RemoveDbSocketConnection(ctx, connId); err != nil {
			tearDownFailures.WriteString("rem_sock_conn ")
		}
	}()

	wg.Wait()

	if tearDownFailures.Len() > 0 {
		var sb strings.Builder
		sb.WriteString("TEARDOWN /sock sub:")
		sb.WriteString(userSub)
		sb.WriteString(" socketid:")
		sb.WriteString(socketId)
		sb.WriteString(" connid:")
		sb.WriteString(connId)
		sb.WriteString(" FAILURES ")
		sb.WriteString(tearDownFailures.String())
		util.ErrorLog.Println(sb.String())
	}

	clients.GetGlobalWorkerPool().CleanUpClientMapping(userSub)
}
