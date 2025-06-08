package main_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func TestMain(m *testing.M) {
	util.ParseEnv()

	cmd, err := testutil.StartTestServer()
	if err != nil {
		panic(err)
	}
	if cmd != nil {
		defer func() {
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("Failed to close server: %v", util.ErrCheck(err))
			}
		}()
	}

	code := m.Run()

	testutil.SaveIntegrations()

	os.Exit(code)
}

func TestIntegrations(t *testing.T) {
	defer testutil.TestPanic(t)

	testIntegrationUser(t)
	testIntegrationGroup(t)
	testIntegrationRoles(t)
	testIntegrationService(t)
	testIntegrationSchedule(t)
	testIntegrationOnboarding(t)
	testIntegrationJoinGroup(t)
	testIntegrationPromoteUser(t)
	testIntegrationUserSchedule(t)
	testIntegrationQuotes(t)
	testIntegrationBookings(t)
	testIntegrationLogout(t)
}
