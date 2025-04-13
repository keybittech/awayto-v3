package main

import (
	"net"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testIntegrationUser(t *testing.T) {
	integrationTest.TestUsers = make(map[int]*TestUser, 10)
	integrationTest.Connections = make(map[string]net.Conn, 10)

	t.Run("user can register and connect", func(t *testing.T) {
		userId := int(time.Now().UnixNano())
		registerKeycloakUserViaForm(userId)

		session, connection, token, ticket, connId := getUser(userId)

		if !util.IsUUID(session.UserSub) {
			t.Errorf("user sub is not a uuid: %s", session.UserSub)
		}

		testUser := &TestUser{
			TestUserId:  userId,
			TestToken:   token,
			TestTicket:  ticket,
			TestConnId:  connId,
			UserSession: session,
		}

		integrationTest.TestUsers[0] = testUser
		integrationTest.Connections[session.UserSub] = connection

		t.Logf("created user #%d with email %s sub %s connId %s", 1, session.UserEmail, session.UserSub, connId)
	})

	failCheck(t)
}
