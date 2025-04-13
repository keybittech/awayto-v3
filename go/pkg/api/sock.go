package api

import (
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
	PING = "PING"
	PONG = "PONG"
)

var (
	socketPingTicker *time.Ticker
	pings            sync.Map // map[string]time.Time
	sockRl           = NewRateLimit("sock", 30, 10, time.Duration(time.Minute))
	pingBytes        = util.GenerateMessage(util.DefaultPadding, &types.SocketMessage{Payload: PING})
)

func (a *API) InitSockServer(mux *http.ServeMux) {

	sockHandler := func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Connection"), "Upgrade") || req.Header.Get("Upgrade") != "websocket" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request: Expected WebSocket"))
			return
		}

		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Sec-WebSocket-Accept", util.ComputeWebSocketAcceptKey(req.Header.Get("Sec-WebSocket-Key")))
		w.WriteHeader(http.StatusSwitchingProtocols)

		hijacker, ok := w.(http.Hijacker)
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

		conn.SetReadDeadline(time.Time{})
		conn.SetWriteDeadline(time.Time{})

		ticket := req.URL.Query().Get("ticket")
		if ticket == "" {
			return
		}

		_, connId, err := util.SplitColonJoined(ticket)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		userConnection, err := a.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
			UserSub: "worker",
			Ticket:  ticket,
		}, conn)
		if err = clients.ChannelError(err, userConnection.Error); err != nil {
			util.ErrorLog.Println(err)
			return
		}

		var exitReason string
		util.SockLog.Println("JOIN /sock sub:" + userConnection.UserSub + " cid:" + connId)
		defer func() {
			util.SockLog.Println("PART /sock sub:" + userConnection.UserSub + " cid:" + connId + " reason:" + exitReason)
		}()

		socketId := util.GetColonJoined(userConnection.UserSub, connId)

		defer a.TearDownSocketConnection(socketId, connId, userConnection.UserSub)

		err = a.Handlers.Database.InitDbSocketConnection(connId, userConnection.UserSub, userConnection.GroupId, userConnection.Roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		err = a.Handlers.Redis.InitRedisSocketConnection(socketId)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

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
			case <-pingTimer.C:
				// let errors log again
				if errorFlag {
					errorFlag = false
				}

				// Close connections not responding to pings
				pingVal, _ := pings.Load(connId)
				if lastSeen, ok := pingVal.(time.Time); ok && time.Since(lastSeen) > 90*time.Second {
					exitReason = "returned due to ping timeout"
					return
				}

				err := a.Handlers.Socket.SendMessage(userConnection.UserSub, connId, &types.SocketMessage{
					Action:  types.SocketActions_PING_PONG,
					Payload: PING,
				})
				if err != nil {
					exitReason = "returned due to ping error " + err.Error()
					return
				}
			case data := <-messages:
				if len(data) == 2 {
					// EOF
					exitReason = "returned due to messages EOF"
					return
				} else {
					if sockRl.Limit(userConnection.UserSub) {
						continue
					}

					go a.SocketRequest(data, connId, socketId, userConnection.UserSub, userConnection.GroupId, userConnection.Roles)
				}
			case err := <-errs:
				// normally shouldn't error, but if it does it could potentially be heavy load
				// so only err log once every ping timer
				if errorFlag {
					continue
				}
				errStr := err.Error()
				sockErr := errors.New("sock read error " + connId + errStr)
				util.ErrorLog.Println(util.ErrCheck(sockErr))

				if strings.Index(errStr, "connection reset by peer") != -1 {
					exitReason = errStr
					return
				}

				errorFlag = true
			}
		}
	}

	mux.HandleFunc("GET /sock",
		a.LimitMiddleware(1, 2)(sockHandler),
	)
}

func (a *API) TearDownSocketConnection(socketId, connId, userSub string) {
	var tearDownFailures string
	// Log errors but attempt to tear down everything
	topics, err := a.Handlers.Redis.HandleUnsub(socketId)
	if err != nil {
		tearDownFailures += "handle_unsub "
	}

	for topic, targets := range topics {
		a.Handlers.Socket.SendMessage(userSub, targets, &types.SocketMessage{
			Action:  types.SocketActions_UNSUBSCRIBE_TOPIC,
			Topic:   topic,
			Payload: socketId,
		})

		err = a.Handlers.Redis.RemoveTopicFromConnection(socketId, topic)
		if err != nil {
			tearDownFailures += "rem_conn_topic "
		}
	}

	err = a.Handlers.Database.RemoveDbSocketConnection(connId)
	if err != nil {
		tearDownFailures += "rem_sock_conn "
	}

	_, err = a.Handlers.Socket.SendCommand(clients.DeleteSocketConnectionSocketCommand, &types.SocketRequestParams{
		UserSub: userSub,
		ConnId:  connId,
	})
	if err != nil {
		tearDownFailures += "del_cmd "
	}

	if tearDownFailures != "" {
		util.ErrorLog.Println("TEARDOWN /sock sub:" + userSub + " socketid:" + socketId + " connid:" + connId + " " + tearDownFailures)
	}

	clients.GetGlobalWorkerPool().CleanUpClientMapping(userSub)
}

func (a *API) SocketRequest(data []byte, connId, socketId, userSub, groupId, roles string) bool {
	socketMessage := a.SocketMessageReceiver(userSub, data)
	if socketMessage == nil {
		return false
	}

	if socketMessage.Action == types.SocketActions_PING_PONG {
		pings.Store(connId, time.Now())
		return true
	}

	a.SocketMessageRouter(socketMessage, connId, socketId, userSub, groupId, roles)
	return true
}
