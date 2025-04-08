package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type SocketEvents struct {
	loadSubscribersEvent []byte
	loadMessagesEvent    []byte
	moveBoxEvent         []byte
	changeSettingEvent   []byte
	pingMessage          []byte
	pongMessage          []byte
}

var subscriptions [][]byte
var unsubscribe []byte
var socketEvents *SocketEvents

var topic, targets, socketId, exchangeId, userSub, groupId, roles string
var session *types.UserSession

func setupSockServer() {
	ticket, connId, err := getSocketTicket()
	if err != nil {
		log.Fatal("failed to get socket ticket in setup", err.Error())
		return
	}

	subscriberRequest, err := api.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
		UserSub: "worker",
		Ticket:  ticket,
	})

	if err = clients.ChannelError(err, subscriberRequest.Error); err != nil {
		log.Fatal("failed to get subscriber socket command in setup", err.Error())
		return
	}

	userSub = subscriberRequest.UserSub
	groupId = subscriberRequest.GroupId
	roles = subscriberRequest.Roles

	session = &types.UserSession{
		UserSub:                 userSub,
		GroupId:                 "group-id",
		AvailableUserGroupRoles: []string{"role1"},
	}

	exchangeId = "0195ec07-e989-71ac-a0c4-f6a08d1f93f6"
	topic = "exchange/0:" + exchangeId

	targets = connId
	socketId = util.GetColonJoined(userSub, connId)

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
	pingMessage := []byte("000022400001f00001f00000000000000000004PING")
	pongMessage := []byte("000022400001f00001f00000000000000000004PONG")

	socketEvents = &SocketEvents{
		loadSubscribersEvent,
		loadMessagesEvent,
		moveBoxEvent,
		changeSettingEvent,
		pingMessage,
		pongMessage,
	}

	time.Sleep(time.Second)
	println("did setup sock")
}

func getSocketTicket() (string, string, error) {
	// Get authentication token first
	token, err := getKeycloakToken(0)
	if err != nil {
		return "", "", fmt.Errorf("error getting auth token: %v", err)
	}

	req, err := http.NewRequest("GET", "https://localhost:7443/api/v1/sock/ticket", nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Debug the response if it's not JSON
	if len(body) > 0 && body[0] != '{' {
		return "", "", fmt.Errorf("unexpected ticket response: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("ticket JSON parse error: %v, body: %s", err, string(body))
	}

	ticket, ok := result["ticket"].(string)
	if !ok {
		return "", "", fmt.Errorf("ticket not found in response")
	}

	ticketParts := strings.Split(ticket, ":")
	_, connId := ticketParts[0], ticketParts[1]

	return ticket, connId, nil
}

// Socket Events

func BenchmarkSocketSubscribe(b *testing.B) {
	_, connId, err := getSocketTicket()
	if err != nil {
		b.Fatal(err)
	}

	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(subscriptions[0], connId, socketId, userSub, groupId, roles)
		b.StopTimer()
		api.SocketRequest(unsubscribe, connId, socketId, userSub, groupId, roles)
		b.StartTimer()
	}
}

func BenchmarkSocketUnsubscribe(b *testing.B) {
	_, connId, err := getSocketTicket()
	if err != nil {
		b.Fatal(err)
	}

	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(unsubscribe, connId, socketId, userSub, groupId, roles)
		b.StopTimer()
		api.SocketRequest(subscriptions[0], connId, socketId, userSub, groupId, roles)
		b.StartTimer()
	}
}

func BenchmarkSocketLoadSubscribers(b *testing.B) {
	_, connId, err := getSocketTicket()
	if err != nil {
		b.Fatal(err)
	}
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(socketEvents.loadSubscribersEvent, connId, socketId, userSub, groupId, roles)
	}
}

func BenchmarkSocketLoadMessages(b *testing.B) {
	_, connId, err := getSocketTicket()
	if err != nil {
		b.Fatal(err)
	}
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(socketEvents.loadMessagesEvent, connId, socketId, userSub, groupId, roles)
	}
}

func BenchmarkSocketDefault(b *testing.B) {
	_, connId, err := getSocketTicket()
	if err != nil {
		b.Fatal(err)
	}
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(socketEvents.changeSettingEvent, connId, socketId, userSub, groupId, roles)
	}
}

func BenchmarkSocketDefaultStore(b *testing.B) {
	_, connId, err := getSocketTicket()
	if err != nil {
		b.Fatal(err)
	}
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketRequest(socketEvents.moveBoxEvent, connId, socketId, userSub, groupId, roles)
	}
}

func BenchmarkSocketMessageReceiverBasicMessage(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketMessageReceiver(userSub, socketEvents.loadSubscribersEvent)
	}
}

func BenchmarkSocketMessageReceiverComplexMessage(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.SocketMessageReceiver(userSub, socketEvents.moveBoxEvent)
	}
}

// Socket Handling

func BenchmarkSocketRoleCall(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.RoleCall(userSub)
	}
}

func BenchmarkSocketSendMessage(b *testing.B) {
	_, connId, err := getSocketTicket()
	if err != nil {
		b.Fatal(err)
	}
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
		api.Handlers.Socket.SendMessage(userSub, targets, &types.SocketMessage{
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
		b.StopTimer()
		session.UserSub = session.UserSub + strconv.Itoa(c)
		b.StartTimer()

		api.Handlers.Socket.GetSocketTicket(session)
	}
}

func BenchmarkSocketInitConnection(b *testing.B) {
	mockConn := &interfaces.NullConn{}
	reset(b)
	for c := 0; c < b.N; c++ {
		b.StopTimer()
		ticket, _ := api.Handlers.Socket.GetSocketTicket(session)
		b.StartTimer()
		api.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Ticket:  ticket,
		}, mockConn)
	}
}

func BenchmarkSocketSendMessageBytes(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendMessageBytes(userSub, targets, socketEvents.changeSettingEvent)
	}
}

func BenchmarkSocketAddSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.AddSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   topic,
			Targets: targets,
		})
	}
}

func BenchmarkSocketHasTopicSubscription(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   topic,
		})
	}
}

func BenchmarkSocketDeleteSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   topic,
		})
	}
}

func BenchmarkSocketAddAndDeleteSubscribedTopic(b *testing.B) {
	reset(b)
	for c := 0; c < b.N; c++ {
		api.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   topic,
		})
		api.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, &types.SocketRequestParams{
			UserSub: userSub,
			Topic:   topic,
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

func getMockClientSocket(t *testing.T) net.Conn {
	ticket, _, err := getSocketTicket()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("wss://localhost:7443/sock?ticket=" + ticket)
	if err != nil {
		t.Fatal(err)
	}

	sockConn, err := tls.Dial("tcp", u.Host, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Fatal(err)
	}

	// Generate WebSocket key
	keyBytes := make([]byte, 16)
	rand.Read(keyBytes)
	secWebSocketKey := base64.StdEncoding.EncodeToString(keyBytes)

	// Send HTTP Upgrade request
	fmt.Fprintf(sockConn, "GET %s HTTP/1.1\r\n", u.RequestURI())
	fmt.Fprintf(sockConn, "Host: %s\r\n", u.Host)
	fmt.Fprintf(sockConn, "Upgrade: websocket\r\n")
	fmt.Fprintf(sockConn, "Connection: Upgrade\r\n")
	fmt.Fprintf(sockConn, "Sec-WebSocket-Key: %s\r\n", secWebSocketKey)
	fmt.Fprintf(sockConn, "Sec-WebSocket-Version: 13\r\n")
	fmt.Fprintf(sockConn, "\r\n")

	reader := bufio.NewReader(sockConn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
	}

	return sockConn
}

func TestHandleSocketConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping socket ping/pong test")
	}
	time.Sleep(2 * time.Second)
	mockConn := getMockClientSocket(t)
	defer mockConn.Close()

	t.Log("starting 1 minute ping/pong test")

	t.Run("periodically pings clients", func(tt *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(time.Minute)

		mockConn.SetReadDeadline(time.Now().Add(70 * time.Second))

		errChan := make(chan error, 1)

		for {
			if time.Now().After(endTime) {
				break
			}

			message, err := util.ReadSocketConnectionMessage(mockConn)
			if err != nil {
				errChan <- err
				continue
			}

			mockConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			tt.Log("Recieved Ping <-")

			if bytes.Equal(message, socketEvents.pingMessage) && time.Now().Before(endTime) {
				tt.Log("Sent Pong ->")
				err = util.WriteSocketConnectionMessage(socketEvents.pongMessage, mockConn)
				if err != nil {
					t.Fatal(err)
					continue
				}
			}
		}
	})

	t.Run("closes the socket after a period of inactivity", func(tt *testing.T) {
		mockConn.SetReadDeadline(time.Now().Add(1 * time.Minute))

		timeoutStart := time.Now()
		var timeoutErr error

		buffer := make([]byte, 1024)
		for {
			_, err := mockConn.Read(buffer)
			if err != nil {
				timeoutErr = err
				break
			}
		}

		timeoutDuration := time.Since(timeoutStart)

		if timeoutErr == io.EOF {
			if timeoutDuration >= 30*time.Second && timeoutDuration <= 60*time.Second {
				tt.Log("socket closed after a period of inactivity")
			} else {
				tt.Fatal("socket closed not within 30-60 seconds of EOF")
			}
		} else {
			tt.Fatalf("Connection ended with unexpected error: %v\n", timeoutErr)
		}
	})
}
