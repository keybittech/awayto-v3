package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

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

func (a *API) SocketMessageRouter(sm *types.SocketMessage, connId, socketId, userSub, groupId, roles string) {
	var err error

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(250*time.Millisecond))
	defer func(us string) {
		clients.GetGlobalWorkerPool().CleanUpClientMapping(us)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
		}
		cancel()
	}(userSub)

	if sm.Topic == "" {
		return
	}

	if sm.Action != types.SocketActions_SUBSCRIBE {
		hasSubRequest, err := a.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   sm.Topic,
		})

		if err != nil {
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
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		} else if hasTracking {
			return
		}

		// topics are in the format of context/action:ref-id
		// for example exchange/2:0195ec07-e989-71ac-a0c4-f6a08d1f93f6
		description, handle, err := util.SplitColonJoined(sm.Topic)
		if err != nil {
			return
		}

		err = a.Handlers.Database.TxExec(func(tx *sql.Tx) error {
			var txErr error
			subscribed, txErr := a.Handlers.Database.GetSocketAllowances(tx, userSub, description, handle)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			if !subscribed {
				return errors.New("not subscribed")
			}
			return nil
		}, userSub, groupId, roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		// Update user's topic cids
		a.Handlers.Redis.TrackTopicParticipant(ctx, sm.Topic, socketId)

		// Get Member Info for anyone connected
		_, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, true)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		_, err = a.Handlers.Socket.SendCommand(clients.AddSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   sm.Topic,
			Targets: cachedParticipantTargets,
		})

		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		a.Handlers.Socket.SendMessage(userSub, connId, &types.SocketMessage{
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

		a.Handlers.Socket.SendMessage(userSub, cachedParticipantTargets, &types.SocketMessage{
			Action:  types.SocketActions_UNSUBSCRIBE_TOPIC,
			Topic:   sm.Topic,
			Payload: socketId,
		})

		a.Handlers.Redis.RemoveTopicFromConnection(socketId, sm.Topic)

		_, err = a.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   sm.Topic,
		})
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

	case types.SocketActions_LOAD_SUBSCRIBERS:

		participants, onlineTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, false)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		err = a.Handlers.Database.TxExec(func(tx *sql.Tx) error {
			var txErr error

			txErr = a.Handlers.Database.GetTopicMessageParticipants(tx, sm.Topic, participants)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}

			txErr = a.Handlers.Database.GetSocketParticipantDetails(tx, participants)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			return nil
		}, userSub, groupId, roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		cachedParticipantsBytes, err := json.Marshal(participants)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		a.Handlers.Socket.SendMessage(userSub, onlineTargets, &types.SocketMessage{
			Action:  types.SocketActions_LOAD_SUBSCRIBERS,
			Sender:  connId,
			Topic:   sm.Topic,
			Payload: string(cachedParticipantsBytes),
		})

	case types.SocketActions_LOAD_MESSAGES:
		var pageInfo map[string]int
		if err := json.Unmarshal([]byte(sm.Payload), &pageInfo); err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}
		messages := [][]byte{}

		err = a.Handlers.Database.TxExec(func(tx *sql.Tx) error {
			var txErr error
			messages, txErr = a.Handlers.Database.GetTopicMessages(tx, sm.Topic, pageInfo["page"], 100) // int(pageInfo["pageSize"])
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			return nil
		}, userSub, groupId, roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}

		for _, messageBytes := range messages {
			if messageBytes != nil {
				_, err := a.Handlers.Socket.SendCommand(clients.SendSocketMessageSocketCommand, &types.SocketRequestParams{
					UserSub:      userSub,
					Targets:      connId,
					MessageBytes: messageBytes,
				})

				if err != nil {
					util.ErrorLog.Println(err)
					return
				}
			}
		}

	default:
		_, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, true)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		err = a.Handlers.Socket.SendMessage(userSub, cachedParticipantTargets, sm)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}
	}

	if sm.Store {
		err = a.Handlers.Database.TxExec(func(tx *sql.Tx) error {
			var txErr error

			txErr = a.Handlers.Database.StoreTopicMessage(tx, connId, sm)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}
			return nil
		}, userSub, groupId, roles)
		if err != nil {
			util.ErrorLog.Println(err)
			return
		}
	}
}
