package main

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testIntegrationUser(t *testing.T) {
	integrationTest.TestUsers = make(map[int32]*types.TestUser, 10)
	connections = make(map[string]net.Conn, 10)

	t.Run("user can register and connect", func(t *testing.T) {
		userId := fmt.Sprint(time.Now().UnixNano())
		userEmail := "1@" + userId
		userPass := "1"
		registerKeycloakUserViaForm(userEmail, userPass)

		session, connection, token, ticket, connId := getUser(userId)

		if !util.IsUUID(session.UserSub) {
			t.Fatalf("user sub is not a uuid: %s", session.UserSub)
		}

		testUser := &types.TestUser{
			TestEmail:   userEmail,
			TestPass:    userPass,
			TestUserId:  userId,
			TestToken:   token,
			TestTicket:  ticket,
			TestConnId:  connId,
			UserSession: session,
		}

		integrationTest.TestUsers[0] = testUser
		connections[session.UserSub] = connection

		t.Logf("created user #%d with email %s sub %s connId %s", 1, session.UserEmail, session.UserSub, connId)
	})
}
