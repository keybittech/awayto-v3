package main_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
)

func testIntegrationUser(t *testing.T) {
	testutil.IntegrationTest.TestUsers = make(map[int32]*testutil.TestUsersStruct, 10)

	t.Run("user can register and connect", func(tt *testing.T) {
		userId := fmt.Sprint(time.Now().UnixNano())
		userEmail := "1@" + userId

		testUser := testutil.NewTestUser(userId, userEmail, "1")
		testutil.IntegrationTest.TestUsers[0] = testUser

		err := testUser.RegisterKeycloakUserViaForm()
		if err != nil {
			t.Fatalf("could not register as admin, %v", err)
		}

		err = testUser.Login()
		if err != nil {
			t.Fatalf("could not login as admin, %v", err)
		}

		profile, err := testUser.GetProfileDetails()
		if err != nil {
			t.Fatalf("could not get admin profile, %v", err)
		}

		if userEmail != profile.GetEmail() {
			t.Fatalf("admin emails didn't match, used: %s, got: %s", userEmail, profile.GetEmail())
		}

		t.Logf("created user %s with pass %s", testUser.GetTestEmail(), testUser.GetTestPass())
	})
}
