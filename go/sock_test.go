package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type NullConn struct{}

func (n NullConn) Read(b []byte) (int, error)         { return 0, io.EOF }
func (n NullConn) Write(b []byte) (int, error)        { return len(b), nil }
func (n NullConn) Close() error                       { return nil }
func (n NullConn) LocalAddr() net.Addr                { return nil }
func (n NullConn) RemoteAddr() net.Addr               { return nil }
func (n NullConn) SetDeadline(t time.Time) error      { return nil }
func (n NullConn) SetReadDeadline(t time.Time) error  { return nil }
func (n NullConn) SetWriteDeadline(t time.Time) error { return nil }

// NewNullConn returns a new no-op connection
func NewNullConn() net.Conn {
	return NullConn{}
}

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
var socketEvents *SocketEvents

var ticket, topic, connId, socketId string
var targets []string
var subscriber *clients.Subscriber
var session *clients.UserSession

func setupSockServer() {
	var err error
	// Get ticket
	ticket, err = getSocketTicket()
	if err != nil {
		println(err.Error())
		return
	}

	subscriber, err = api.Handlers.Socket.GetSubscriberByTicket(ticket)
	if err != nil {
		println(err.Error())
		return
	}

	session = &clients.UserSession{
		UserSub:                 subscriber.UserSub,
		GroupId:                 "group-id",
		AvailableUserGroupRoles: []string{"role1"},
	}

	go api.HandleSockConnection(subscriber, &NullConn{}, ticket)
	time.Sleep(2 * time.Second)

	ticketParts := strings.Split(ticket, ":")
	_, connId := ticketParts[0], ticketParts[1]

	topic = "exchange/0:" + exchangeId

	targets = []string{connId}
	socketId = util.GetSocketId(subscriber.UserSub, connId)

	subscriptions = [][]byte{
		[]byte("00001800001f00001f0000000047exchange/0:" + exchangeId + "00036" + connId),
		[]byte("00001800001f00001f0000000047exchange/1:" + exchangeId + "00036" + connId),
		[]byte("00001800001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId),
	}

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
		api.SocketRequest(subscriber, connId, subscriptions[i])
	}
	time.Sleep(time.Second)
	println("did setup sock")
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

// Socket Events

func BenchmarkSocketSubscribe(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, connId, subscriptions[0])
	}
}

func BenchmarkSocketUnsubscribe(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, connId, subscriptions[0])
	}
}

func BenchmarkSocketLoadSubscribers(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, connId, socketEvents.loadSubscribersEvent)
	}
}

func BenchmarkSocketLoadMessages(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, connId, socketEvents.loadMessagesEvent)
	}
}

func BenchmarkSocketDefault(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, connId, socketEvents.changeSettingEvent)
	}
}

func BenchmarkSocketDefaultStore(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriber, connId, socketEvents.moveBoxEvent)
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
		Cids:   targets,
		Role:   "role",
		Color:  "color",
		Exists: true,
		Online: false,
	}

	mergedParticipants[connId] = participant

	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendMessage(targets, &clients.SocketMessage{
			Action:  6,
			Sender:  connId,
			Topic:   topic,
			Payload: mergedParticipants,
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
	ticket, _ := api.Handlers.Socket.GetSocketTicket(session)
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.InitConnection(&NullConn{}, subscriber.UserSub, ticket)
	}
}

func BenchmarkSocketGetSubscriberByTicket(b *testing.B) {
	ticket, _ := api.Handlers.Socket.GetSocketTicket(session)
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.GetSubscriberByTicket(ticket)
	}
}

func BenchmarkSocketNotifyTopicUnsub(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.NotifyTopicUnsub(topic, socketId, targets)
	}
}

func BenchmarkSocketSendMessageBytes(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendMessageBytes(targets, socketEvents.changeSettingEvent)
	}
}

func BenchmarkSocketAddSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.AddSubscribedTopic(subscriber.UserSub, topic, targets)
	}
}

func BenchmarkSocketHasTopicSubscription(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.HasTopicSubscription(subscriber.UserSub, topic)
	}
}

func BenchmarkSocketDeleteSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.DeleteSubscribedTopic(subscriber.UserSub, topic)
	}
}

func BenchmarkSocketAddAndDeleteSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.AddSubscribedTopic(subscriber.UserSub, topic, targets)
		api.Handlers.Socket.DeleteSubscribedTopic(subscriber.UserSub, topic)
	}
}

func BenchmarkSocketGetSubscribedTopicTargets(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.GetSubscribedTopicTargets(subscriber.UserSub, topic)
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
		api.Handlers.Redis.GetCachedParticipants(ctx, topic)
	}
}

func BenchmarkSocketGetParticipantTargets(b *testing.B) {
	participants := make(clients.SocketParticipants, 1)
	participant := &types.SocketParticipant{
		Name:   "name",
		Scid:   "scid",
		Cids:   targets,
		Role:   "role",
		Color:  "color",
		Exists: true,
		Online: false,
	}

	participants[connId] = participant
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Redis.GetParticipantTargets(participants)
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

// Socket Database Functions

// Utilities

func BenchmarkSocketSplitSocketId(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		util.SplitSocketId(topic)
	}
}

func BenchmarkSocketWriteSocketMessage(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		util.WriteSocketConnectionMessage(socketEvents.loadSubscribersEvent, &NullConn{})
	}
}
