package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"
)

type SocketServer struct {
	Connections map[string]net.Conn
}

func (a *API) InitSockServer(mux *http.ServeMux) {

	sockHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.Header.Get("Connection"), "Upgrade") && req.Header.Get("Upgrade") == "websocket" {
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
			w.Header().Set("Sec-WebSocket-Accept", util.ComputeWebSocketAcceptKey(req.Header.Get("Sec-WebSocket-Key")))
			w.WriteHeader(http.StatusSwitchingProtocols)

			hijacker, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "Websocket upgrade failed", http.StatusInternalServerError)
				return
			}

			conn, _, err := hijacker.Hijack()
			if err != nil {
				http.Error(w, "Websocket upgrade failed", http.StatusInternalServerError)
				return
			}

			go a.HandleSockConnection(conn, req)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request: Expected WebSocket"))
		}
	})

	middlewareHandler := ApplyMiddleware(sockHandler, []Middleware{
		a.CorsMiddleware,
		a.SocketAuthMiddleware,
	})

	mux.HandleFunc("GET /sock", middlewareHandler)
}

func (a *API) HandleSockConnection(conn net.Conn, req *http.Request) {

	defer conn.Close()

	ticket := req.URL.Query().Get("ticket")
	if ticket == "" {
		util.ErrorLog.Println(util.ErrCheck(fmt.Errorf("no ticket")))
		return
	}

	subscriber, err := a.Handlers.Socket.GetSubscriberByTicket(ticket)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}

	_, connId, err := util.SplitSocketId(ticket)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}

	socketId := util.GetSocketId(subscriber.UserSub, connId)

	deferSockClose, err := a.Handlers.Socket.InitConnection(conn, subscriber.UserSub, ticket)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}
	defer deferSockClose()

	deferSockDbClose, err := a.Handlers.Database.InitDBSocketConnection(subscriber.UserSub, connId)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}
	defer deferSockDbClose()

	err = a.Handlers.Redis.InitRedisSocketConnection(socketId)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}
	defer func() {
		topics, err := a.Handlers.Redis.HandleUnsub(socketId)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		for topic, targets := range topics {
			a.Handlers.Socket.NotifyTopicUnsub(topic, socketId, targets)
			a.Handlers.Redis.RemoveTopicFromConnection(socketId, topic)
		}
	}()

	messages := make(chan []byte)
	errs := make(chan error)
	go func() {
		for {
			data, err := util.ReadSocketConnectionMessage(conn)
			if err != nil {
				errs <- err
				if io.EOF == err {
					return
				}
			} else {
				messages <- data
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	lastPong := time.Now()

	for {
		select {
		case <-ticker.C:
			pingBytes, _ := json.Marshal(&clients.SocketMessage{Payload: "PING"})
			if err := util.WriteSocketConnectionMessage(pingBytes, conn); err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}
		case data := <-messages:

			if len(data) == 2 {
				// EOF
				return
			} else {

				var socketMessage clients.SocketMessage

				if err := json.Unmarshal(data, &socketMessage); err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
					continue
				}

				if socketMessage.Payload == "PONG" {
					lastPong = time.Now()
					continue
				}

				go a.SocketMessageRouter(socketMessage, subscriber.UserSub, connId)
			}
		case err := <-errs:
			util.ErrorLog.Println(util.ErrCheck(err))
		}

		if time.Since(lastPong) > 1*time.Minute {
			return
		}
	}
}

func (a *API) SocketMessageRouter(sm clients.SocketMessage, userSub, connId string) {

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	if sm.Topic == "" {
		return
	}

	description, handle, err := util.SplitSocketId(sm.Topic)
	if err != nil {
		return
	}

	socketId := util.GetSocketId(userSub, connId)

	switch sm.Action {
	case types.SocketActions_SUBSCRIBE:

		allowances, err := a.Handlers.Database.GetSocketAllowances(userSub)
		if err != nil {
			return
		}
		subscribed := false

		for _, allowance := range allowances {
			switch description {
			case fmt.Sprintf("exchange/%d", types.ExchangeActions_EXCHANGE_TEXT.Number()),
				fmt.Sprintf("exchange/%d", types.ExchangeActions_EXCHANGE_CALL.Number()),
				fmt.Sprintf("exchange/%d", types.ExchangeActions_EXCHANGE_WHITEBOARD.Number()):

				if allowance.Id == handle {
					subscribed = true
				}
			default:
				return
			}
		}

		if !subscribed {
			return
		}

		// Update user's topic cids
		a.Handlers.Redis.TrackTopicParticipant(ctx, sm.Topic, socketId)

		// Get Member Info for anyone connected
		cachedParticipants := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic)
		targets := a.Handlers.Redis.GetParticipantTargets(cachedParticipants)

		// Setup server client with new subscription to topic including any existing connections
		a.Handlers.Socket.AddSubscribedTopic(userSub, sm.Topic, targets)

		a.Handlers.Socket.SendMessage([]string{connId}, clients.SocketMessage{
			Action: types.SocketActions_SUBSCRIBE,
			Topic:  sm.Topic,
		})

	case types.SocketActions_UNSUBSCRIBE:

		cachedParticipants := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic)
		targets := a.Handlers.Redis.GetParticipantTargets(cachedParticipants)

		a.Handlers.Socket.NotifyTopicUnsub(sm.Topic, socketId, targets)
		a.Handlers.Redis.RemoveTopicFromConnection(connId, sm.Topic)
		a.Handlers.Socket.DeleteSubscribedTopic(userSub, sm.Topic)

	case types.SocketActions_LOAD_SUBSCRIBERS:

		cachedParticipants := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic)
		topicMessageParticipants := a.Handlers.Database.GetTopicMessageParticipants(sm.Topic)

		mergedParticipants := make(clients.SocketParticipants)

		for _, participants := range []clients.SocketParticipants{cachedParticipants, topicMessageParticipants} {
			for userSub, participant := range participants {
				if mergedP, ok := mergedParticipants[userSub]; ok {
					for _, cid := range participant.Cids {
						if !slices.Contains(mergedP.Cids, cid) {
							mergedP.Cids = append(mergedP.Cids, cid)
						}
					}
					if !mergedP.Online && participant.Online {
						mergedP.Online = true
					}
				} else {
					mergedParticipants[userSub] = participant
				}
			}
		}

		mergedParticipants = a.Handlers.Database.GetSocketParticipantDetails(mergedParticipants)

		targets := a.Handlers.Redis.GetParticipantTargets(cachedParticipants)
		a.Handlers.Socket.SendMessage(targets, clients.SocketMessage{
			Action:  types.SocketActions_LOAD_SUBSCRIBERS,
			Sender:  connId,
			Topic:   sm.Topic,
			Payload: mergedParticipants,
		})

	case types.SocketActions_LOAD_MESSAGES:

		if a.Handlers.Socket.HasTopicSubscription(userSub, sm.Topic) {
			targets := []string{connId}
			pageInfo := sm.Payload.(map[string]interface{})
			messages := a.Handlers.Database.GetTopicMessages(sm.Topic, int(pageInfo["page"].(float64)), int(pageInfo["pageSize"].(float64)))
			for _, messageBytes := range messages {
				if messageBytes != nil {
					a.Handlers.Socket.SendMessageBytes(targets, messageBytes)
				}
			}
		}

	default:
		if a.Handlers.Socket.HasTopicSubscription(userSub, sm.Topic) {
			cachedParticipants := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic)
			targets := a.Handlers.Redis.GetParticipantTargets(cachedParticipants)
			a.Handlers.Socket.SendMessage(targets, sm)
		}
	}

	if sm.Store && a.Handlers.Socket.HasTopicSubscription(userSub, sm.Topic) {
		a.Handlers.Database.StoreTopicMessage(connId, sm.Topic, sm)
	}
}
