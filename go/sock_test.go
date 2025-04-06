package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func getSocketTicket() (string, error) {
	// Get authentication token first
	token, err := getKeycloakToken(0)
	if err != nil {
		return "", fmt.Errorf("error getting auth token: %v", err)
	}

	req, err := http.NewRequest("GET", "https://localhost:7443/api/v1/sock/ticket", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Debug the response if it's not JSON
	if len(body) > 0 && body[0] != '{' {
		return "", fmt.Errorf("unexpected ticket response: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("ticket JSON parse error: %v, body: %s", err, string(body))
	}

	ticket, ok := result["ticket"].(string)
	if !ok {
		return "", fmt.Errorf("ticket not found in response")
	}

	return ticket, nil
}

var exchangeId = "0195ec07-e989-71ac-a0c4-f6a08d1f93f6"

type SocketEvents struct {
	loadSubscribersEvent []byte
	loadMessagesEvent    []byte
	moveBoxEvent         []byte
	changeSettingEvent   []byte
}

var subscriptions [][]byte
var unsubscribe []byte
var socketEvents *SocketEvents

var ticket, topic, connId, socketId string
var targets string
var subscriber *types.Subscriber
var session *types.UserSession

func setupSockServer() {
	var err error
	// Get ticket
	ticket, err = getSocketTicket()
	if err != nil {
		println(err.Error())
		return
	}

	subscriberRequest, err := api.Handlers.Socket.SendCommand(clients.GetSubscriberSocketCommand, interfaces.SocketRequest{
		SocketRequestParams: &types.SocketRequestParams{
			UserSub: "worker",
			Ticket:  ticket,
		},
	})

	if err = clients.ChannelError(err, subscriberRequest.Error); err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		return
	}

	subscriber = subscriberRequest.Subscriber

	session = &types.UserSession{
		UserSub:                 subscriber.UserSub,
		GroupId:                 "group-id",
		AvailableUserGroupRoles: []string{"role1"},
	}

	go api.HandleSockConnection(subscriber, interfaces.NewNullConn(), ticket)
	time.Sleep(2 * time.Second)

	ticketParts := strings.Split(ticket, ":")
	_, connId := ticketParts[0], ticketParts[1]

	topic = "exchange/0:" + exchangeId

	targets = connId
	socketId = util.GetSocketId(subscriber.UserSub, connId)

	subscriptions = [][]byte{
		[]byte("00001800001f00001f0000000047exchange/0:" + exchangeId + "00036" + connId),
		[]byte("00001800001f00001f0000000047exchange/1:" + exchangeId + "00036" + connId),
		[]byte("00001800001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId),
	}

	unsubscribe = []byte("00001900001f00001f0000000047exchange/0:" + exchangeId + "00036" + connId)

	loadSubscribersEvent := []byte("000021000001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId)
	loadMessagesEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId +
		`00024{"page":1,"pageSize":10}`)
	moveBoxEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId +
		`00150{"boxes":[{"id":1743421799040,"color":"#9ec4b8","x":248,"y":301,"text":"E=mc^2"}]}`)
	changeSettingEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId +
		`00032{"settings":{"highlight":false}}`)

	socketEvents = &SocketEvents{
		loadSubscribersEvent,
		loadMessagesEvent,
		moveBoxEvent,
		changeSettingEvent,
	}

	// Subscribe to a topic
	for i := 0; i < len(subscriptions); i++ {
		api.SocketRequest(subscriber, subscriptions[i], connId, socketId)
	}
	time.Sleep(time.Second)
	println("did setup sock")
}

// Socket Events

func BenchmarkSocketSubscribe(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, subscriptions[0], connId, socketId)
		b.StopTimer()
		api.SocketRequest(subscriber, unsubscribe, connId, socketId)
		b.StartTimer()
	}
}

func BenchmarkSocketUnsubscribe(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, unsubscribe, connId, socketId)
		b.StopTimer()
		api.SocketRequest(subscriber, subscriptions[0], connId, socketId)
		b.StartTimer()
	}
}

func BenchmarkSocketLoadSubscribers(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, socketEvents.loadSubscribersEvent, connId, socketId)
	}
}

func BenchmarkSocketLoadMessages(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, socketEvents.loadMessagesEvent, connId, socketId)
	}
}

func BenchmarkSocketDefault(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, socketEvents.changeSettingEvent, connId, socketId)
	}
}

func BenchmarkSocketDefaultStore(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, socketEvents.moveBoxEvent, connId, socketId)
	}
}

func BenchmarkSocketMessageReceiverBasicMessage(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketMessageReceiver(subscriber.UserSub, socketEvents.loadSubscribersEvent)
	}
}

func BenchmarkSocketMessageReceiverComplexMessage(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketMessageReceiver(subscriber.UserSub, socketEvents.moveBoxEvent)
	}
}

// Socket Handling

func BenchmarkSocketRoleCall(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.RoleCall(subscriber.UserSub)
	}
}

func BenchmarkSocketSendMessage(b *testing.B) {
	mergedParticipants := make(map[string]*types.SocketParticipant, 1)
	participant := &types.SocketParticipant{
		Name:   "name",
		Scid:   "scid",
		Cids:   []string{targets},
		Role:   "role",
		Color:  "color",
		Exists: true,
		Online: false,
	}

	mergedParticipants[connId] = participant

	mergedParticipantsBytes, err := json.Marshal(mergedParticipants)
	if err != nil {
		b.Fatal(err)
	}

	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendMessage(subscriber.UserSub, targets, &types.SocketMessage{
			Action:  6,
			Sender:  connId,
			Topic:   topic,
			Payload: string(mergedParticipantsBytes),
		})
	}
}

func BenchmarkSocketGetSocketTicket(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.GetSocketTicket(session)
	}
}

func BenchmarkSocketInitConnection(b *testing.B) {
	mockConn := &interfaces.NullConn{}
	reset(b)
	for c := 0; c < b.N; c++ {
		b.StopTimer()
		ticket, _ = api.Handlers.Socket.GetSocketTicket(session)
		b.StartTimer()
		api.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: subscriber.UserSub,
				Ticket:  ticket,
			},
			Conn: mockConn,
		})
	}
}

func BenchmarkSocketGetSubscriberByTicket(b *testing.B) {
	ticket, _ := api.Handlers.Socket.GetSocketTicket(session)
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.GetSubscriberSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: "worker",
				Ticket:  ticket,
			},
		})
	}
}

func BenchmarkSocketSendMessageBytes(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendMessageBytes(subscriber.UserSub, targets, socketEvents.changeSettingEvent)
	}
}

func BenchmarkSocketAddSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.AddSubscribedTopicSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: subscriber.UserSub,
				Topic:   topic,
				Targets: targets,
			},
		})
	}
}

func BenchmarkSocketHasTopicSubscription(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: subscriber.UserSub,
				Topic:   topic,
			},
		})
	}
}

func BenchmarkSocketDeleteSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: subscriber.UserSub,
				Topic:   topic,
			},
		})
	}
}

func BenchmarkSocketAddAndDeleteSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: subscriber.UserSub,
				Topic:   topic,
			},
		})
		api.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, interfaces.SocketRequest{
			SocketRequestParams: &types.SocketRequestParams{
				UserSub: subscriber.UserSub,
				Topic:   topic,
			},
		})
	}
}

// Socket Redis Functions
func BenchmarkSocketInitRedisSocketConnection(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Redis.InitRedisSocketConnection(socketId)
	}
}

func BenchmarkSocketTrackTopicParticipant(b *testing.B) {
	ctx := context.Background()
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Redis.TrackTopicParticipant(ctx, topic, socketId)
	}
}

func BenchmarkSocketGetCachedParticipants(b *testing.B) {
	ctx := context.Background()
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Redis.GetCachedParticipants(ctx, topic, false)
	}
}

func BenchmarkSocketGetCachedParticipantsTargetsOnly(b *testing.B) {
	ctx := context.Background()
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Redis.GetCachedParticipants(ctx, topic, true)
	}
}

func BenchmarkSocketRemoveTopicFromConnection(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Redis.RemoveTopicFromConnection(socketId, topic)
	}
}

func BenchmarkSocketHandleUnsub(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Redis.HandleUnsub(socketId)
	}
}

// // Utilities
//
// func BenchmarkSocketGenerateMessage(b *testing.B) {
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		util.GenerateMessage(5, &types.SocketMessage{})
// 	}
// }
//
// func BenchmarkSocketGetSocketId(b *testing.B) {
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		util.GetSocketId(session.UserSub, connId)
// 	}
// }
//
// func BenchmarkSocketSplitSocketId(b *testing.B) {
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		util.SplitSocketId(topic)
// 	}
// }
//
// func BenchmarkSocketWriteSocketMessage(b *testing.B) {
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		util.WriteSocketConnectionMessage(socketEvents.loadSubscribersEvent, &interfaces.NullConn{})
// 	}
// }
