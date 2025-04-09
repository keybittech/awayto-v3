package api

import (
	"io"
	"net"
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
	PING             = "PING"
	PONG             = "PONG"
	sockHandlerLimit = rate.Limit(30)
	sockHandlerBurst = 10
)

var (
	socketPingTicker         *time.Ticker
	pings                    sync.Map // map[string]time.Time
	sockLimitMu, sockLimited = NewRateLimit("sock")
	pingBytes                = util.GenerateMessage(util.DefaultPadding, &types.SocketMessage{Payload: PING})
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

		go func(c net.Conn, e chan error, m chan []byte) {
			for {
				data, err := util.ReadSocketConnectionMessage(c)
				pings.Store(connId, time.Now())
				if err != nil {
					if io.EOF == err {
						return
					}
					e <- err
				} else {
					m <- data
				}
			}
		}(conn, errs, messages)

		pings.Store(connId, time.Now())

		pingTimer := time.NewTicker(30 * time.Second)

		for {
			select {
			case <-pingTimer.C:
				// Close connections not responding to pings
				pingVal, _ := pings.Load(connId)
				if lastSeen, ok := pingVal.(time.Time); ok && time.Since(lastSeen) > 90*time.Second {
					println("returned due to ping timeout")
					return
				}

				err := a.Handlers.Socket.SendMessage(userConnection.UserSub, connId, &types.SocketMessage{
					Action:  types.SocketActions_PING_PONG,
					Payload: PING,
				})
				if err != nil {
					println("returned due to ping error ", err.Error())
					return
				}
			case data := <-messages:
				if len(data) == 2 {
					// EOF
					println("returned due to messages EOF")
					return
				} else {

					limited := Limiter(sockLimitMu, sockLimited, sockHandlerLimit, sockHandlerBurst, userConnection.UserSub)
					if limited {
						continue
					}
					go a.SocketRequest(data, connId, socketId, userConnection.UserSub, userConnection.GroupId, userConnection.Roles)
				}
			case err := <-errs:
				println("read error " + err.Error())
				util.ErrorLog.Println(util.ErrCheck(err))
			}
		}
	}

	mux.HandleFunc("GET /sock",
		a.LimitMiddleware(1, 2)(sockHandler),
	)
}

func (a *API) TearDownSocketConnection(socketId, connId, userSub string) bool {
	topics, err := a.Handlers.Redis.HandleUnsub(socketId)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return false
	}

	for topic, targets := range topics {
		err = a.Handlers.Socket.SendMessage(userSub, targets, &types.SocketMessage{
			Action:  types.SocketActions_UNSUBSCRIBE_TOPIC,
			Topic:   topic,
			Payload: socketId,
		})
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return false
		}

		err = a.Handlers.Redis.RemoveTopicFromConnection(socketId, topic)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return false
		}
	}

	err = a.Handlers.Database.RemoveDbSocketConnection(connId)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return false
	}

	_, err = a.Handlers.Socket.SendCommand(clients.DeleteSocketConnectionSocketCommand, &types.SocketRequestParams{
		UserSub: userSub,
		ConnId:  connId,
	})
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return false
	}

	clients.GetGlobalWorkerPool().CleanUpClientMapping(userSub)
	return true
}

func (a *API) SocketRequest(data []byte, connId, socketId, userSub, groupId, roles string) bool {
	socketMessage := a.SocketMessageReceiver(userSub, data)
	if socketMessage == nil {
		return false
	}

	a.SocketMessageRouter(socketMessage, connId, socketId, userSub, groupId, roles)
	return true
}
