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
	t.Run("users can join a group with a code after log in", func(t *testing.T) {
		for c := existingUsers; c < existingUsers+6; c++ {
			time.Sleep(500 * time.Millisecond)
			joinViaRegister := c%2 == 0
			userId := fmt.Sprint(time.Now().UnixNano())

			if joinViaRegister {
				t.Logf("Code Registering #%s Code: %s", userId, integrationTest.Group.Code)
				// This takes care of attach user to group and activate profile on the backend
				_, err := registerKeycloakUserViaForm(userId, integrationTest.Group.Code)
				if err != nil {
					t.Fatalf("failed to register with group code %v", err)
				}
			} else {
				t.Logf("Normal Registering #%s", userId)
				_, err := registerKeycloakUserViaForm(userId)
				if err != nil {
					t.Fatalf("failed to register without group code %v", err)
				}
			}

			session, connection, token, ticket, connId := getUser(userId)

			if len(ticket) != 73 {
				t.Errorf("bad ticket: got ticket (auth:connid) %s %d", ticket, len(ticket))
			}

			if !util.IsUUID(session.UserSub) {
				t.Errorf("user sub is not a uuid: %v", session)
			}

			if integrationTest.Group.Code == "" {
				t.Errorf("no group id to join with %v", integrationTest.Group)
			}

			if !joinViaRegister {
				// Join Group -- puts the user in the app db
				joinGroupRequestBytes, err := protojson.Marshal(&types.JoinGroupRequest{
					Code: integrationTest.Group.Code,
				})
				if err != nil {
					t.Errorf("error marshalling join group request, user: %d error: %v", c, err)
				}

				joinGroupResponse := &types.JoinGroupResponse{}
				err = apiRequest(token, http.MethodPost, "/api/v1/group/join", joinGroupRequestBytes, nil, joinGroupResponse)
				if err != nil {
					t.Errorf("error posting join group request, user: %d error: %v", c, err)
				}
				if !joinGroupResponse.Success {
					t.Errorf("join group internal was unsuccessful %v", joinGroupResponse)
				}

				// Attach User to Group -- adds the user to keycloak records
				attachUserRequestBytes, err := protojson.Marshal(&types.AttachUserRequest{
					Code: integrationTest.Group.Code,
				})
				if err != nil {
					t.Errorf("error marshalling attach user request, user: %d error: %v", c, err)
				}

				attachUserResponse := &types.AttachUserResponse{}
				err = apiRequest(token, http.MethodPost, "/api/v1/group/attach/user", attachUserRequestBytes, nil, attachUserResponse)
				if err != nil {
					t.Errorf("error posting attach user request, user: %d error: %v", c, err)
				}
				if !attachUserResponse.Success {
					t.Errorf("attach user internal was unsuccessful %v", attachUserResponse)
				}

				// Activate Profile -- lets the user view the internal login pages
				activateProfileResponse := &types.ActivateProfileResponse{}
				err = apiRequest(token, http.MethodPatch, "/api/v1/profile/activate", nil, nil, activateProfileResponse)
				if err != nil {
					t.Errorf("error patch activate profile group request, user: %d error: %v", c, err)
				}
				if !activateProfileResponse.Success {
					t.Errorf("activate profile internal was unsuccessful %v", activateProfileResponse)
				}

				// Get new token after group setup to check group membersip
				token, session, err = getKeycloakToken(userId)
				if err != nil {
					t.Errorf("failed to get new token after joining group %v", err)
				}
			}

			if len(session.SubGroups) == 0 {
				t.Errorf("no group id after getting new token %v", session)
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
