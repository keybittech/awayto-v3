package main

// type SocketEvents struct {
// 	loadSubscribersEvent []byte
// 	loadMessagesEvent    []byte
// 	moveBoxEvent         []byte
// 	changeSettingEvent   []byte
// 	pingMessage          []byte
// 	pongMessage          []byte
// }
//
// var subscriptions [][]byte
// var unsubscribe []byte
// var socketEvents *SocketEvents

// func setupSockServer() {
//
// 	// userSub = subscriberRequest.UserSub
// 	// groupId = subscriberRequest.GroupId
// 	// roles = subscriberRequest.Roles
// 	//
// 	exchangeId := "0195ec07-e989-71ac-a0c4-f6a08d1f93f6"
// 	// topic = "exchange/0:" + exchangeId
// 	//
// 	// targets = connId
// 	// socketId = util.GetColonJoined(userSub, connId)
//
// 	connId := "test-id"
// 	subscriptions = [][]byte{
// 		[]byte("00001800001f00001f0000000047exchange/0:" + exchangeId + "00036" + connId),
// 		[]byte("00001800001f00001f0000000047exchange/1:" + exchangeId + "00036" + connId),
// 		[]byte("00001800001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId),
// 	}
//
// 	unsubscribe = []byte("00001900001f00001f0000000047exchange/0:" + exchangeId + "00036" + connId)
//
// 	loadSubscribersEvent := []byte("000021000001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId)
// 	loadMessagesEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId +
// 		`00024{"page":1,"pageSize":10}`)
// 	moveBoxEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId +
// 		`00150{"boxes":[{"id":1743421799040,"color":"#9ec4b8","x":248,"y":301,"text":"E=mc^2"}]}`)
// 	changeSettingEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId +
// 		`00032{"settings":{"highlight":false}}`)
// 	pingMessage := []byte("000022400001f00001f00000000000000000004PING")
// 	pongMessage := []byte("000022400001f00001f00000000000000000004PONG")
//
// 	socketEvents = &SocketEvents{
// 		loadSubscribersEvent,
// 		loadMessagesEvent,
// 		moveBoxEvent,
// 		changeSettingEvent,
// 		pingMessage,
// 		pongMessage,
// 	}
//
// 	println("did setup sock")
// }

// Socket Events

// func BenchmarkSocketSubscribe(b *testing.B) {
// 	params, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	exchangeId := "socket-subscribe-topic"
// 	subscribe := []byte("00001800001f00001f0000000047exchange/0:" + exchangeId + "00036" + connId)
// 	unsubscribe := []byte("00001900001f00001f0000000047exchange/0:" + exchangeId + "00036" + connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketRequest(subscribe, connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 		b.StopTimer()
// 		mainApi.SocketRequest(unsubscribe, connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 		b.StartTimer()
// 	}
// }
//
// func BenchmarkSocketUnsubscribe(b *testing.B) {
// 	params, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketRequest(unsubscribe, connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 		b.StopTimer()
// 		mainApi.SocketRequest(subscriptions[0], connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 		b.StartTimer()
// 	}
// }
//
// func BenchmarkSocketLoadSubscribers(b *testing.B) {
// 	params, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketRequest(socketEvents.loadSubscribersEvent, connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 	}
// }
//
// func BenchmarkSocketLoadMessages(b *testing.B) {
// 	params, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketRequest(socketEvents.loadMessagesEvent, connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 	}
// }
//
// func BenchmarkSocketDefault(b *testing.B) {
// 	params, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketRequest(socketEvents.changeSettingEvent, connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 	}
// }
//
// func BenchmarkSocketDefaultStore(b *testing.B) {
// 	params, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketRequest(socketEvents.moveBoxEvent, connId, socketId, params.UserSub, params.GroupId, params.Roles)
// 	}
// }
//
// func BenchmarkSocketMessageReceiverBasicMessage(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketMessageReceiver(session.UserSub, socketEvents.loadSubscribersEvent)
// 	}
// }
//
// func BenchmarkSocketMessageReceiverComplexMessage(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.SocketMessageReceiver(session.UserSub, socketEvents.moveBoxEvent)
// 	}
// }

// Socket Handling

// func BenchmarkSocketRoleCall(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Socket.RoleCall(session.UserSub)
// 	}
// }

// func BenchmarkSocketSendMessage(b *testing.B) {
// 	_, session, _, connId := getUserSocketSession(0)
// 	mergedParticipants := make(map[string]*types.SocketParticipant, 1)
// 	participant := &types.SocketParticipant{
// 		Name:   "name",
// 		Scid:   "scid",
// 		Cids:   []string{connId},
// 		Role:   "role",
// 		Color:  "color",
// 		Exists: true,
// 		Online: false,
// 	}
//
// 	mergedParticipants[connId] = participant
//
// 	mergedParticipantsBytes, err := json.Marshal(mergedParticipants)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
//
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Socket.SendMessage(session.UserSub, connId, &types.SocketMessage{
// 			Action:  6,
// 			Sender:  connId,
// 			Topic:   "test-topic",
// 			Payload: string(mergedParticipantsBytes),
// 		})
// 	}
// }
// func BenchmarkSocketGetSocketTicket(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		b.StopTimer()
// 		session.UserSub = session.UserSub + strconv.Itoa(c)
// 		b.StartTimer()
//
// 		mainApi.Handlers.Socket.GetSocketTicket(session)
// 	}
// }

// func BenchmarkSocketInitConnection(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	mockConn := &util.NullConn{}
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		b.StopTimer()
// 		ticket, _ := mainApi.Handlers.Socket.GetSocketTicket(session)
// 		b.StartTimer()
// 		mainApi.Handlers.Socket.SendCommand(clients.CreateSocketConnectionSocketCommand, &types.SocketRequestParams{
// 			UserSub: session.UserSub,
// 			Ticket:  ticket,
// 		}, mockConn)
// 	}
// }

// This
// func doSendMessageBytesBench(clientCount, messagesPerClient int, b *testing.B) {
// 	time.Sleep(5 * time.Second)
// 	subscribers, connections := getSubscribers(clientCount)
// 	if subscribers == nil || len(subscribers) == 0 {
// 		b.Fatal("couldn't get subscribers for send message bytes")
// 	}
//
// 	testExchangeId := "test-multi-user-send-message-exchange-id"
// 	var targets string
// 	messages := make(map[int][]byte, clientCount)
// 	for sid, subscriber := range subscribers {
// 		targets += subscriber.ConnectionId
// 		messages[sid] = []byte("00001600001f00001f0000000047exchange/2:" + testExchangeId + "00036" + subscriber.ConnectionId + `00032{"settings":{"highlight":false}}`)
// 		subscriptionMessage := []byte("00001800001f00001f0000000047exchange/0:" + testExchangeId + "00036" + subscriber.ConnectionId)
// 		err := util.WriteSocketConnectionMessage(subscriptionMessage, connections[subscriber.UserSub])
// 		if err != nil {
// 			b.Fatal("could not subscribe ", err)
// 			continue
// 		}
// 	}
//
// 	reset(b)
// 	for i := 0; i < b.N/(clientCount*messagesPerClient); i++ {
// 		var wg sync.WaitGroup
//
// 		for c := 0; c < clientCount; c++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				subscriber := subscribers[c]
// 				connection := connections[subscriber.UserSub]
// 				message := messages[c]
//
// 				for j := 0; j < messagesPerClient; j++ {
// 					util.WriteSocketConnectionMessage(message, connection)
// 				}
// 			}()
// 		}
//
// 		wg.Wait()
// 	}
// }

//
// errChan := make(chan error, 1)
//
// for {
// 	if time.Now().After(endTime) {
// 		break
// 	}
//
// 	message, err := util.ReadSocketConnectionMessage(sockConn)
// 	if err != nil {
// 		errChan <- err
// 		continue
// 	}
//
// 	sockConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
// 	tt.Log("Recieved Ping <-")
//
// 	if bytes.Equal(message, socketEvents.pingMessage) && time.Now().Before(endTime) {
// 		tt.Log("Sent Pong ->")

// 	}
// }

// func BenchmarkSocketSendMessageBytes1Client1Message(b *testing.B) {
// 	doSendMessageBytesBench(1, 1, b)
// }
//
// func BenchmarkSocketSendMessageBytes1Client10Messages(b *testing.B) {
// 	doSendMessageBytesBench(1, 10, b)
// }
//
// func BenchmarkSocketSendMessageBytes10Clients1Message(b *testing.B) {
// 	doSendMessageBytesBench(10, 1, b)
// }
//
// func BenchmarkSocketSendMessageBytes10Clients10Messages(b *testing.B) {
// 	doSendMessageBytesBench(10, 10, b)
// }
//
// func BenchmarkSocketAddSubscribedTopic(b *testing.B) {
// 	_, session, _, connId := getUserSocketSession(0)
// 	topic := "test-topic"
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Socket.SendCommand(clients.AddSubscribedTopicSocketCommand, &types.SocketRequestParams{
// 			UserSub: session.UserSub,
// 			Topic:   topic,
// 			Targets: connId,
// 		})
// 	}
// }
//
// func BenchmarkSocketHasTopicSubscription(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	topic := "test-topic"
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, &types.SocketRequestParams{
// 			UserSub: session.UserSub,
// 			Topic:   topic,
// 		})
// 	}
// }
//
// func BenchmarkSocketDeleteSubscribedTopic(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	topic := "test-topic"
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, &types.SocketRequestParams{
// 			UserSub: session.UserSub,
// 			Topic:   topic,
// 		})
// 	}
// }
//
// func BenchmarkSocketAddAndDeleteSubscribedTopic(b *testing.B) {
// 	_, session, _, _ := getUserSocketSession(0)
// 	topic := "test-topic"
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Socket.SendCommand(clients.HasSubscribedTopicSocketCommand, &types.SocketRequestParams{
// 			UserSub: session.UserSub,
// 			Topic:   topic,
// 		})
// 		mainApi.Handlers.Socket.SendCommand(clients.DeleteSubscribedTopicSocketCommand, &types.SocketRequestParams{
// 			UserSub: session.UserSub,
// 			Topic:   topic,
// 		})
// 	}
// }
//
// // Socket Redis Functions
// func BenchmarkSocketInitRedisSocketConnection(b *testing.B) {
// 	_, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Redis.InitRedisSocketConnection(socketId)
// 	}
// }
//
// func BenchmarkSocketTrackTopicParticipant(b *testing.B) {
// 	_, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	topic := "test-topic"
// 	ctx := context.Background()
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Redis.TrackTopicParticipant(ctx, topic, socketId)
// 	}
// }
//
// func BenchmarkSocketGetCachedParticipants(b *testing.B) {
// 	topic := "test-topic"
// 	ctx := context.Background()
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Redis.GetCachedParticipants(ctx, topic, false)
// 	}
// }
//
// func BenchmarkSocketGetCachedParticipantsTargetsOnly(b *testing.B) {
// 	topic := "test-topic"
// 	ctx := context.Background()
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Redis.GetCachedParticipants(ctx, topic, true)
// 	}
// }
//
// func BenchmarkSocketRemoveTopicFromConnection(b *testing.B) {
// 	_, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	topic := "test-topic"
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Redis.RemoveTopicFromConnection(socketId, topic)
// 	}
// }
//
// func BenchmarkSocketHandleUnsub(b *testing.B) {
// 	_, session, _, connId := getUserSocketSession(0)
// 	socketId := util.GetColonJoined(session.UserSub, connId)
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		mainApi.Handlers.Redis.HandleUnsub(socketId)
// 	}
// }

// func TestHandleSocketConnection(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping socket ping/pong test")
// 	}
// 	time.Sleep(10 * time.Second)
//
// 	_, ticket, _, _ := getSocketTicket(int(time.Now().UnixNano()))
//
// 	sockConn := getClientSocketConnection(ticket)
// 	if sockConn != nil {
// 		t.Fatal("failed to get mock connection in handle sock connection")
// 	}
// 	defer sockConn.Close()
//
// 	t.Log("starting 3 minute ping/pong test")
//
// 	t.Run("periodically pings clients", func(tt *testing.T) {
// 		startTime := time.Now()
// 		endTime := startTime.Add(time.Minute)
//
// 		sockConn.SetReadDeadline(time.Now().Add(70 * time.Second))
//
// 		errChan := make(chan error, 1)
//
// 		for {
// 			if time.Now().After(endTime) {
// 				break
// 			}
//
// 			message, err := util.ReadSocketConnectionMessage(sockConn)
// 			if err != nil {
// 				errChan <- err
// 				continue
// 			}
//
// 			sockConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
// 			tt.Log("Recieved Ping <-")
//
// 			if bytes.Equal(message, socketEvents.pingMessage) && time.Now().Before(endTime) {
// 				tt.Log("Sent Pong ->")
// 				err = util.WriteSocketConnectionMessage(socketEvents.pongMessage, sockConn)
// 				if err != nil {
// 					t.Fatal(err)
// 					continue
// 				}
// 			}
// 		}
// 	})
//
// 	t.Run("closes the socket after a period of inactivity", func(tt *testing.T) {
// 		sockConn.SetReadDeadline(time.Now().Add(2 * time.Minute))
//
// 		timeoutStart := time.Now()
// 		var timeoutErr error
//
// 		buffer := make([]byte, 1024)
// 		for {
// 			_, err := sockConn.Read(buffer)
// 			if err != nil {
// 				timeoutErr = err
// 				break
// 			}
// 		}
//
// 		timeoutDuration := time.Since(timeoutStart)
//
// 		if timeoutErr == io.EOF {
// 			if timeoutDuration >= 80*time.Second && timeoutDuration <= 100*time.Second {
// 				tt.Log("socket closed after a period of inactivity")
// 			} else {
// 				tt.Fatal("socket closed not within 30-60 seconds of EOF")
// 			}
// 		} else {
// 			tt.Fatalf("Connection ended with unexpected error: %v\n", timeoutErr)
// 		}
// 	})
// }
