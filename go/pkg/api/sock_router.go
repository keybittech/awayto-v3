package api

import (
	"context"
	json "encoding/json"
	"errors"
	"strconv"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (a *API) SocketMessageReceiver(data []byte) *types.SocketMessage {
	finish := util.RunTimer()
	defer finish()
	var socketMessage types.SocketMessage

	messageParams := make([]string, 7)

	cursor := 0
	var curr string
	var err error
	for i := range 7 {
		cursor, curr, err = util.ParseMessage(util.DefaultPadding, cursor, data)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			messageParams[i] = ""
			continue
		}
		messageParams[i] = curr
	}

	actionId, err := strconv.ParseInt(messageParams[0], 10, 32)
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

func (a *API) SocketMessageRouter(ctx context.Context, connId, socketId string, sm *types.SocketMessage, ds clients.DbSession) {
	finish := util.RunTimer()
	defer finish()

	if sm.Topic == "" {
		return
	}

	if sm.Action != types.SocketActions_SUBSCRIBE {
		hasSubRequest, err := a.Handlers.Socket.SendCommand(ctx, clients.HasSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: ds.ConcurrentUserSession.GetUserSub(),
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
		_, handle, err := util.SplitColonJoined(sm.Topic)
		if err != nil {
			return
		}

		subscribed, err := ds.GetSocketAllowances(ctx, handle)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		if !subscribed {
			util.DebugLog.Println(errors.New("not subscribed"))
			return
		}

		// Update user's topic cids
		err = a.Handlers.Redis.TrackTopicParticipant(ctx, sm.Topic, socketId)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		// Get Member Info for anyone connected
		_, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, true)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		_, err = a.Handlers.Socket.SendCommand(ctx, clients.AddSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: ds.ConcurrentUserSession.GetUserSub(),
			Topic:   sm.Topic,
			Targets: cachedParticipantTargets,
		})

		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		err = a.Handlers.Socket.SendMessage(ctx, ds.ConcurrentUserSession.GetUserSub(), connId, &types.SocketMessage{
			Action: types.SocketActions_SUBSCRIBE,
			Topic:  sm.Topic,
		})
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

	case types.SocketActions_UNSUBSCRIBE:

		hasTracking, err := a.Handlers.Redis.HasTracking(ctx, sm.Topic, socketId)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		} else if !hasTracking {
			return
		}

		_, cachedParticipantTargets, err := a.Handlers.Redis.GetCachedParticipants(ctx, sm.Topic, true)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		if cachedParticipantTargets != "" {
			err = a.Handlers.Socket.SendMessage(ctx, ds.ConcurrentUserSession.GetUserSub(), cachedParticipantTargets, &types.SocketMessage{
				Action:  types.SocketActions_UNSUBSCRIBE,
				Topic:   sm.Topic,
				Payload: socketId,
			})
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
			}
		}

		err = a.Handlers.Redis.RemoveTopicFromConnection(ctx, socketId, sm.Topic)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
		}

		_, err = a.Handlers.Socket.SendCommand(ctx, clients.DeleteSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: ds.ConcurrentUserSession.GetUserSub(),
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

		err = ds.GetTopicMessageParticipants(ctx, participants)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		err = ds.GetSocketParticipantDetails(ctx, participants)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		participantsBytes, err := json.Marshal(participants)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		err = a.Handlers.Socket.SendMessage(ctx, ds.ConcurrentUserSession.GetUserSub(), onlineTargets, &types.SocketMessage{
			Action:  types.SocketActions_LOAD_SUBSCRIBERS,
			Sender:  connId,
			Topic:   sm.Topic,
			Payload: string(participantsBytes),
		})
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

	case types.SocketActions_LOAD_MESSAGES:
		var pageInfo map[string]int
		if err := json.Unmarshal([]byte(sm.Payload), &pageInfo); err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		messages, err := ds.GetTopicMessages(ctx, pageInfo["page"], 100) // int(pageInfo["pageSize"])
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}

		for _, messageBytes := range messages {
			if messageBytes != nil {
				_, err := a.Handlers.Socket.SendCommand(ctx, clients.SendSocketMessageSocketCommand, &types.SocketRequestParams{
					UserSub:      ds.ConcurrentUserSession.GetUserSub(),
					Targets:      connId,
					MessageBytes: messageBytes,
				})

				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
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

		err = a.Handlers.Socket.SendMessage(ctx, ds.ConcurrentUserSession.GetUserSub(), cachedParticipantTargets, sm)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}
	}

	if sm.Store {
		err := ds.StoreTopicMessage(ctx, connId, sm)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			return
		}
	}
}
