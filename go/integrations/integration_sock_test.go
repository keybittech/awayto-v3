package main_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testIntegrationSock(t *testing.T) {
	staff1 := testutil.IntegrationTest.TestUsers[1]
	member1 := testutil.IntegrationTest.TestUsers[4]

	err := staff1.GetVaultKey()
	if err != nil {
		t.Fatalf("error getting staff vault key, %v", err)
	}
	err = member1.GetVaultKey()
	if err != nil {
		t.Fatalf("error getting member vault key, %v", err)
	}

	t.Run("user can get a ticket", func(tt *testing.T) {
		err := staff1.GetSocketTicket()
		if err != nil {
			t.Fatalf("error getting staff socket ticket, %v", err)
		}

		if len(strings.Split(staff1.GetTestTicket(), ":")) != 2 {
			t.Fatalf("invalid user ticket format, %v", staff1.GetTestTicket())
		}

		if !util.IsUUID(staff1.GetTestConnId()) {
			t.Fatalf("invalid user connection id format, %v", staff1.GetTestTicket())
		}

		err = member1.GetSocketTicket()
		if err != nil {
			t.Fatalf("error getting member socket ticket, %v", err)
		}
	})

	t.Run("user can connect to the socket", func(tt *testing.T) {
		err := staff1.GetSocketConnection()
		if err != nil {
			t.Fatalf("error getting staff socket connection, %v", err)
		}

		if staff1.Socket == nil {
			t.Fatalf("user socket is nil")
		}

		err = member1.GetSocketConnection()
		if err != nil {
			t.Fatalf("error getting member socket connection, %v", err)
		}
	})

	bookings := staff1.Bookings
	bookingId := bookings[0].GetId()
	staff1ConnId := staff1.GetTestConnId()
	member1ConnId := member1.GetTestConnId()

	done := make(chan struct{})

	t.Cleanup(func() {
		close(done)
		staff1.Socket.Close()
		member1.Socket.Close()
	})

	messages := make(chan []byte, 10)
	errors := make(chan error, 1)
	go func(tt *testing.T) {
		for {
			select {
			case <-done:
				return
			default:
				_, message, err := staff1.Socket.ReadMessage()
				if err != nil {
					select {
					case errors <- err:
					case <-done:
					}
					return
				}
				messages <- message
			}
		}
	}(t)

	t.Run("cannot act on unsubscribed topics", func(tt *testing.T) {
		loadSubscribersEvent := []byte("000021000001f00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + "00000")
		if err := staff1.WriteSocketMessage(loadSubscribersEvent); err != nil {
			t.Fatalf("failed to write load sub message, %v", err)
		}

		subscribeEvent := []byte("00001800001f00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + "00000")
		if err := staff1.WriteSocketMessage(subscribeEvent); err != nil {
			t.Fatalf("failed to write cannot act subscribe message, %v", err)
		}

		select {
		case msg := <-messages:
			if bytes.Equal(msg[0:6], []byte("000018")) {
				if err := staff1.WriteSocketMessage([]byte("00001900001f00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + "00000")); err != nil {
					t.Fatalf("failed to write cannot act unsubscribe message, %v", err)
				}
				return
			}
		case err := <-errors:
			tt.Fatalf("cannot act socket read error, %v", err)
		case <-time.After(2 * time.Second):
			tt.Fatal("cannot act timeout")
		}
	})

	t.Run("user can subscribe to topics", func(tt *testing.T) {
		for _, sub := range [][]byte{
			[]byte("00001800001f00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + "00000"),
			[]byte("00001800001f00001f0000000047exchange/1:" + bookingId + "00036" + staff1ConnId + "00000"),
			[]byte("00001800001f00001f0000000047exchange/2:" + bookingId + "00036" + staff1ConnId + "00000"),
		} {
			if err := staff1.WriteSocketMessage(sub); err != nil {
				t.Fatalf("failed to write socket message, %v", err)
			}
		}

		expectedSubs := 3

		for {
			select {
			case msg := <-messages:
				if bytes.Equal(msg[0:6], []byte("000018")) {
					expectedSubs--
					if expectedSubs == 0 {
						member1.WriteSocketMessage([]byte("00001800001f00001f0000000047exchange/0:" + bookingId + "00036" + member1ConnId + "00000"))
						return
					}
				}
				continue
			case err := <-errors:
				tt.Fatalf("subscribe socket read error, %v", err)
			case <-time.After(2 * time.Second):
				tt.Fatal("subscribe timeout")
			}
		}
	})

	t.Run("can load user list of topic", func(tt *testing.T) {
		loadSubscribersEvent := []byte("000021000001f00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + "00000")
		if err := staff1.WriteSocketMessage(loadSubscribersEvent); err != nil {
			t.Fatalf("load topic users failed to write load message, %v", err)
		}

		select {
		case msg := <-messages:
			if !bytes.Contains(msg, []byte("cids")) {
				tt.Fatal("load topic no cids")
			}
			if !bytes.Contains(msg, []byte("scid")) {
				tt.Fatal("load topic no user id")
			}
			return
		case err := <-errors:
			tt.Fatalf("load topic users socket read error, %v", err)
		case <-time.After(2 * time.Second):
			tt.Fatal("load topic users timeout")
		}
	})

	t.Run("can send topic messages", func(tt *testing.T) {
		greetingMessage := []byte("000021200001t00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + "00048" + `{"style":"written","message":"Hi, how are you?"}`)
		if err := staff1.WriteSocketMessage(greetingMessage); err != nil {
			t.Fatalf("send failed to write send greeting, %v", err)
		}

		responseMessage := []byte("000021200001t00001f0000000047exchange/0:" + bookingId + "00036" + member1ConnId + "00051" + `{"style":"written","message":"Great, thanks! You?"}`)
		if err := member1.WriteSocketMessage(responseMessage); err != nil {
			t.Fatalf("send failed to write send response, %v", err)
		}

		idleTimeout := 2 * time.Second
		timer := time.NewTimer(idleTimeout)
		defer timer.Stop()
		sent := 0

		for {
			select {
			case <-messages:
				sent++
				timer.Reset(idleTimeout)
			case err := <-errors:
				tt.Fatalf("send socket read error, %v", err)
			case <-timer.C:
				if sent > 0 {
					println("sent", sent)
					return
				}
				tt.Fatal("send timeout")
			}
		}
	})

	t.Run("can receive topic messages", func(tt *testing.T) {
		loadMessagesEvent := []byte("00001600001f00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + `00024{"page":1,"pageSize":10}`)
		if err := staff1.WriteSocketMessage(loadMessagesEvent); err != nil {
			t.Fatalf("receive failed to write load message, %v", err)
		}

		idleTimeout := 2 * time.Second
		timer := time.NewTimer(idleTimeout)
		defer timer.Stop()
		received := 0

		for {
			select {
			case <-messages:
				received++
				timer.Reset(idleTimeout)
			case err := <-errors:
				tt.Fatalf("receive socket read error, %v", err)
			case <-timer.C:
				if received > 0 {
					println("received", received)
					return
				}
				tt.Fatal("receive timeout")
			}
		}
	})

	t.Run("user can unsubscribe from topics", func(tt *testing.T) {
		for _, sub := range [][]byte{
			[]byte("00001900001f00001f0000000047exchange/0:" + bookingId + "00036" + staff1ConnId + "00000"),
			[]byte("00001900001f00001f0000000047exchange/1:" + bookingId + "00036" + staff1ConnId + "00000"),
			[]byte("00001900001f00001f0000000047exchange/2:" + bookingId + "00036" + staff1ConnId + "00000"),
		} {
			if err := staff1.WriteSocketMessage(sub); err != nil {
				t.Fatalf("failed to write socket message, %v", err)
			}
		}

		expectedUnsubs := 3

		for {
			select {
			case msg := <-messages:
				if bytes.Equal(msg[0:6], []byte("000019")) {
					expectedUnsubs--
					if expectedUnsubs == 0 {
						member1.WriteSocketMessage([]byte("00001900001f00001f0000000047exchange/0:" + bookingId + "00036" + member1ConnId + "00000"))
						return
					}
				}
				continue
			case err := <-errors:
				tt.Fatalf("unsubscribe socket read error, %v", err)
			case <-time.After(2 * time.Second):
				tt.Fatal("unsubscribe timeout")
			}
		}
	})
}

// 	moveBoxEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId + `00150{"boxes":[{"id":1743421799040,"color":"#9ec4b8","x":248,"y":301,"text":"E=mc^2"}]}`)
// 	changeSettingEvent := []byte("00001600001f00001f0000000047exchange/2:" + exchangeId + "00036" + connId + `00032{"settings":{"highlight":false}}`)
