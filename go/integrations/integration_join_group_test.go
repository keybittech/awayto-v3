package main_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationJoinGroup(t *testing.T) {
	existingUsers := 1
	t.Run("users can join a group with a code after log in", func(tt *testing.T) {
		if testutil.IntegrationTest.Group.Code == "" {
			t.Fatalf("no group id to join with %v", testutil.IntegrationTest.Group)
		}

		for c := existingUsers; c < existingUsers+6; c++ {
			joinViaRegister := c%2 == 0

			userId := fmt.Sprint(time.Now().UnixNano())
			userEmail := strconv.Itoa(c+1) + "@" + userId
			testUser := testutil.NewTestUser(userId, userEmail, "1")
			testutil.IntegrationTest.TestUsers[int32(c)] = testUser

			if joinViaRegister {
				t.Logf("Code Registering #%s Code: %s", userId, testutil.IntegrationTest.Group.Code)
				// This takes care of attach user to group and activate profile on the backend
				err := testUser.RegisterKeycloakUserViaForm(testutil.IntegrationTest.Group.Code)
				if err != nil {
					t.Fatalf("failed to register with group code %v", err)
				}
			} else {
				t.Logf("Normal Registering #%s", userId)
				err := testUser.RegisterKeycloakUserViaForm()
				if err != nil {
					t.Fatalf("failed to register without group code %v", err)
				}
			}

			err := testUser.Login()
			if err != nil {
				t.Fatalf("could not login as test user, %v", err)
			}

			profile, err := testUser.GetProfileDetails()
			if err != nil {
				t.Fatalf("could not get test user profile, %v", err)
			}

			if userEmail != profile.GetEmail() {
				t.Fatalf("user emails didn't match, used: %s, got: %s", userEmail, profile.GetEmail())
			}

			if !joinViaRegister {
				// Join Group -- puts the user in the app db, adds them to the group, sets profile active = true
				joinGroupRequestBytes, err := protojson.Marshal(&types.JoinGroupRequest{
					Code: testutil.IntegrationTest.Group.Code,
				})
				if err != nil {
					t.Fatalf("error marshalling join group request, user: %d error: %v", c, err)
				}

				joinGroupResponse := &types.JoinGroupResponse{}
				err = testUser.DoHandler(http.MethodPost, "/api/v1/group/join", joinGroupRequestBytes, nil, joinGroupResponse)
				if err != nil {
					t.Fatalf("error posting join group request, user: %d error: %v", c, err)
				}
				if !joinGroupResponse.Success {
					t.Fatalf("join group internal was unsuccessful %v", joinGroupResponse)
				}
			}

			profile, err = testUser.GetProfileDetails()
			if err != nil {
				t.Fatalf("could not get test user profile after join, %v", err)
			}

			if profile.GetRoleName() == "" {
				t.Fatalf("no role name after joining group %v", profile)
			}

			t.Logf("created user #%d, email: %s", c, testUser.GetTestEmail())
		}
	})
}
