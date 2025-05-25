package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationJoinGroup(t *testing.T) {
	existingUsers := 1
	t.Run("users can join a group with a code after log in", func(tt *testing.T) {
		for c := existingUsers; c < existingUsers+6; c++ {
			joinViaRegister := c%2 == 0
			userId := fmt.Sprint(time.Now().UnixNano())
			userEmail := "1@" + userId
			userPass := "1"

			if joinViaRegister {
				t.Logf("Code Registering #%s Code: %s", userId, integrationTest.Group.Code)
				// This takes care of attach user to group and activate profile on the backend
				_, err := registerKeycloakUserViaForm(userEmail, userPass, integrationTest.Group.Code)
				if err != nil {
					t.Fatalf("failed to register with group code %v", err)
				}
			} else {
				t.Logf("Normal Registering #%s", userId)
				_, err := registerKeycloakUserViaForm(userEmail, userPass)
				if err != nil {
					t.Fatalf("failed to register without group code %v", err)
				}
			}

			session, connection, token, ticket, connId := getUser(t, userId)

			if len(ticket) != 73 {
				t.Fatalf("bad ticket: got ticket (auth:connid) %s %d", ticket, len(ticket))
			}

			if !util.IsUUID(session.UserSub) {
				t.Fatalf("user sub is not a uuid: %v", session)
			}

			if integrationTest.Group.Code == "" {
				t.Fatalf("no group id to join with %v", integrationTest.Group)
			}

			if !joinViaRegister {
				// Join Group -- puts the user in the app db, adds them to the group, sets profile active = true
				joinGroupRequestBytes, err := protojson.Marshal(&types.JoinGroupRequest{
					Code: integrationTest.Group.Code,
				})
				if err != nil {
					t.Fatalf("error marshalling join group request, user: %d error: %v", c, err)
				}

				joinGroupResponse := &types.JoinGroupResponse{}
				err = apiRequest(token, http.MethodPost, "/api/v1/group/join", joinGroupRequestBytes, nil, joinGroupResponse)
				if err != nil {
					t.Fatalf("error posting join group request, user: %d error: %v", c, err)
				}
				if !joinGroupResponse.Success {
					t.Fatalf("join group internal was unsuccessful %v", joinGroupResponse)
				}

				// Get new token after group setup to check group membersip
				token, session, err = getKeycloakToken(userId)
				if err != nil {
					t.Fatalf("failed to get new token after joining group %v", err)
				}
			}

			if len(session.SubGroupPaths) == 0 {
				t.Fatalf("no subgroup paths after getting new token %v", session)
			}

			testUser := &types.TestUser{
				TestEmail:   "1@" + userId,
				TestPass:    "1",
				TestUserId:  userId,
				TestToken:   token,
				TestTicket:  ticket,
				TestConnId:  connId,
				UserSession: session,
			}

			t.Logf("created user #%d with sub %s connId %s", c, session.UserSub, connId)

			integrationTest.TestUsers[int32(c)] = testUser
			connections[session.UserSub] = connection

			t.Logf("user %d has email %s sub %s", c, testUser.UserSession.UserEmail, testUser.UserSession.UserSub)
		}
	})
}
