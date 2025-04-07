package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/time/rate"
)

type SocketConnection struct {
	Conn    net.Conn
	Message []byte
}

type SocketServer struct {
	Connections sync.Map
}

var (
	socketPingTicker         *time.Ticker
	pings                    map[string]time.Time
	PING                     = "PING"
	PONG                     = "PONG"
	sockLimitMu, sockLimited = NewRateLimit("sock")
	sockHandlerLimit         = rate.Limit(30)
	sockHandlerBurst         = 10
)

func (a *API) InitSockServer(mux *http.ServeMux) {

	socketPingTicker = time.NewTicker(30 * time.Second)

	pings = make(map[string]time.Time)

	sockHandler := func(w http.ResponseWriter, req *http.Request) {
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

			ticket := req.URL.Query().Get("ticket")
			if ticket == "" {
				return
			}

			subscriberRequest, err := a.Handlers.Socket.SendCommand(clients.GetAuthSubscriberSocketCommand, &types.SocketRequestParams{
				UserSub: "worker",
				Ticket:  ticket,
			})

			if err = clients.ChannelError(err, subscriberRequest.Error); err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				return
			}

			var wg sync.WaitGroup
			wg.Add(1)
			a.HandleSockConnection(&wg, subscriberRequest.Subscriber, conn, ticket)
			wg.Wait()
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request: Expected WebSocket"))
		}
	}

	mux.HandleFunc("GET /sock",
		a.LimitMiddleware(1, 2)(sockHandler),
	)
}

func (a *API) HandleSockConnection(wg *sync.WaitGroup, subscriber *types.Subscriber, conn net.Conn, ticket string) {

	defer wg.Done()
	defer conn.Close()

	_, connId, err := util.SplitSocketId(ticket)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}

	socketId := util.GetSocketId(subscriber.UserSub, connId)

	response, err := a.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
		UserSub: subscriber.UserSub,
		Ticket:  ticket,
	}, conn)

	if err = clients.ChannelError(err, response.Error); err != nil {
		util.ErrorLog.Println(err)
		return
	}

	defer func(us, t string) {
		a.Handlers.Socket.SendCommand(clients.DeleteSocketConnectionSocketCommand, &types.SocketRequestParams{
			UserSub: us,
			Ticket:  t,
		})

		clients.GetGlobalWorkerPool().CleanUpClientMapping(us)
	}(subscriber.UserSub, ticket)

	var deferSockDbClose func()
	err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
		var txErr error
		deferSockDbClose, txErr = a.Handlers.Database.InitDBSocketConnection(tx, subscriber.UserSub, connId)
		if txErr != nil {
			return util.ErrCheck(txErr)
		}
		return nil
	}, subscriber.UserSub, subscriber.GroupId, subscriber.Roles)
	if err != nil {
		util.ErrorLog.Println(err)
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
			a.Handlers.Socket.SendMessage(subscriber.UserSub, targets, &types.SocketMessage{
				Action:  types.SocketActions_UNSUBSCRIBE_TOPIC,
				Topic:   topic,
				Payload: socketId,
			})
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

	pings[connId] = time.Now()

	err = a.PingPong(subscriber.UserSub, connId)

	for {
		select {
		case <-socketPingTicker.C:
			err := a.PingPong(subscriber.UserSub, connId)
			if err != nil {
				return
			}
		case data := <-messages:
			if len(data) == 2 {
				// EOF
				return
			} else {

				limited := Limiter(sockLimitMu, sockLimited, sockHandlerLimit, sockHandlerBurst, subscriber.UserSub)
				if limited {
					continue
				}
				go a.SocketRequest(subscriber, data, connId, socketId)
			}
		case err := <-errs:
			util.ErrorLog.Println(util.ErrCheck(err))
		}

		if time.Since(pings[connId]) > 1*time.Minute {
			return
		}
	}
}

func (a *API) SocketRequest(subscriber *types.Subscriber, data []byte, connId, socketId string) bool {
	socketMessage := a.SocketMessageReceiver(subscriber.UserSub, data)
	if socketMessage == nil {
		return false
	}

	if socketMessage.Payload == PONG {
		pings[connId] = time.Now()
		return true
	}

	a.SocketMessageRouter(socketMessage, subscriber, connId, socketId)
	return true
}

func (a *API) SocketMessageReceiver(userSub string, data []byte) *types.SocketMessage {
	var socketMessage types.SocketMessage

	messageParams := make([]string, 7)

	if len(data) > util.MAX_SOCKET_MESSAGE_LENGTH {
		return nil
	}

	cursor := 0
	var curr string
	var err error
	for i := 0; i < 7; i++ {
		cursor, curr, err = util.ParseMessage(util.DefaultPadding, cursor, data)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			messageParams[i] = ""
			continue
		}
		messageParams[i] = curr
	}

	actionId, err := strconv.Atoi(messageParams[0])
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return nil
	}

	socketMessage.Action = types.SocketActions(actionId)
	socketMessage.Store = messageParams[1] == "t"
	socketMessage.Historical = messageParams[2] == "t"
	socketMessage.Timestamp = messageParams[3]
	socketMessage.Topic = messageParams[4]
	socketMessage.Sender = messageParams[5]
	socketMessage.Payload = messageParams[6]

	return &socketMessage
}

func (a *API) SocketMessageRouter(sm *types.SocketMessage, subscriber *types.Subscriber, connId, socketId string) {
	var err error

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer func(us string) {
		clients.GetGlobalWorkerPool().CleanUpClientMapping(us)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
		}
		cancel()
	}(subscriber.UserSub)

	if sm.Topic == "" {
		return
	}

	if sm.Action != types.SocketActions_SUBSCRIBE {
		hasSubRequest, err := a.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: subscriber.UserSub,
			Topic:   sm.Topic,
		})

		if err = clients.ChannelError(err, hasSubRequest.Error); err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		if !hasSubRequest.HasSub {
			return
		}
	}

	switch sm.Action {
	case types.SocketActions_SUBSCRIBE:

		hasTracking, err := a.Handlers.Redis.HasTracking(ctx, sm.Topic, socketId)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		} else if hasTracking {
			return
		}

		// Split socket id is only used here as a convenience
		// topics are in the format of context/action:ref-id
		// for example exchange/2:0195ec07-e989-71ac-a0c4-f6a08d1f93f6
		description, handle, err := util.SplitSocketId(sm.Topic)
		if err != nil {
			return
		}

		err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
			var txErr error
			subscribed, txErr := a.Handlers.Database.GetSocketAllowances(tx, subscriber.UserSub, description, handle)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			if !subscribed {
				return errors.New("not subscribed")
			}
			return nil
		}, subscriber.UserSub, subscriber.GroupId, subscriber.Roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		// Update user's topic cids
		a.Handlers.Redis.TrackTopicParticipant(ctx, sm.Topic, socketId)

		// Get Member Info for anyone connected
		_, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, true)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		response, err := a.Handlers.Socket.SendCommand(clients.AddSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: subscriber.UserSub,
			Topic:   sm.Topic,
			Targets: cachedParticipantTargets,
		})

		if err = clients.ChannelError(err, response.Error); err != nil {
			util.ErrorLog.Println(err)
			return
		}

		a.Handlers.Socket.SendMessage(subscriber.UserSub, connId, &types.SocketMessage{
			Action: types.SocketActions_SUBSCRIBE,
			Topic:  sm.Topic,
		})

	case types.SocketActions_UNSUBSCRIBE:

		hasTracking, err := a.Handlers.Redis.HasTracking(ctx, sm.Topic, socketId)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		} else if !hasTracking {
			return
		}

		_, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, true)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		a.Handlers.Socket.SendMessage(subscriber.UserSub, cachedParticipantTargets, &types.SocketMessage{
			Action:  types.SocketActions_UNSUBSCRIBE_TOPIC,
			Topic:   sm.Topic,
			Payload: socketId,
		})

		a.Handlers.Redis.RemoveTopicFromConnection(connId, sm.Topic)

		response, err := a.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: subscriber.UserSub,
			Topic:   sm.Topic,
		})

		if err = clients.ChannelError(err, response.Error); err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

	case types.SocketActions_LOAD_SUBSCRIBERS:

		cachedParticipants, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, false)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		topicMessageParticipants := make(map[string]*types.SocketParticipant)

		err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
			var txErr error

			topicMessageParticipants, txErr = a.Handlers.Database.GetTopicMessageParticipants(tx, sm.Topic)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}

			for scid, topicMessageParticipant := range topicMessageParticipants {
				if cachedParticipant, ok := cachedParticipants[scid]; ok {
					for _, cid := range topicMessageParticipant.Cids {
						if strings.Index(cachedParticipantTargets, cid) == -1 {
							cachedParticipant.Cids = append(cachedParticipant.Cids, cid)
						}
					}
				} else {
					cachedParticipants[scid] = topicMessageParticipant
				}
			}

			cachedParticipants, txErr = a.Handlers.Database.GetSocketParticipantDetails(tx, cachedParticipants)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			return nil
		}, subscriber.UserSub, subscriber.GroupId, subscriber.Roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		cachedParticipantsBytes, err := json.Marshal(cachedParticipants)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		a.Handlers.Socket.SendMessage(subscriber.UserSub, cachedParticipantTargets, &types.SocketMessage{
			Action:  types.SocketActions_LOAD_SUBSCRIBERS,
			Sender:  connId,
			Topic:   sm.Topic,
			Payload: string(cachedParticipantsBytes),
		})

	case types.SocketActions_LOAD_MESSAGES:
		var pageInfo map[string]int
		if err := json.Unmarshal([]byte(sm.Payload), &pageInfo); err != nil {
			util.ErrorLog.Println(err)
			return
		}
		messages := [][]byte{}

		err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
			var txErr error
			messages, txErr = a.Handlers.Database.GetTopicMessages(tx, sm.Topic, pageInfo["page"], 100) // int(pageInfo["pageSize"])
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			return nil
		}, subscriber.UserSub, subscriber.GroupId, subscriber.Roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		for _, messageBytes := range messages {
			if messageBytes != nil {
				response, err := a.Handlers.Socket.SendCommand(clients.SendSocketMessageSocketCommand, &types.SocketRequestParams{
					UserSub:      subscriber.UserSub,
					Targets:      connId,
					MessageBytes: messageBytes,
				})

				if err = clients.ChannelError(err, response.Error); err != nil {
					util.ErrorLog.Println(err)
					return
				}
			}
		}

	default:
		_, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, true)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		err = a.Handlers.Socket.SendMessage(subscriber.UserSub, cachedParticipantTargets, sm)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}
	}

	if sm.Store {
		err = a.Handlers.Database.TxExec(func(tx interfaces.IDatabaseTx) error {
			var txErr error

			txErr = a.Handlers.Database.StoreTopicMessage(tx, connId, sm)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			return nil
		}, subscriber.UserSub, subscriber.GroupId, subscriber.Roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}
	}
}

var pingBytes = util.GenerateMessage(util.DefaultPadding, &types.SocketMessage{Payload: PING})

func (a *API) PingPong(userSub, connId string) error {
	if err := a.Handlers.Socket.SendMessageBytes(userSub, connId, pingBytes); err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return err
	}
	clients.GetGlobalWorkerPool().CleanUpClientMapping(userSub)
	return nil
}
