package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

type TestUser struct {
	TestUserId  int
	TestToken   string
	TestTicket  string
	TestConnId  string
	UserSession *types.UserSession
	Quote       *types.IQuote
}

type IntegrationTest struct {
	TestUsers      map[int]*TestUser
	Connections    map[string]net.Conn
	Roles          map[string]*types.IRole
	MemberRole     *types.IRole
	StaffRole      *types.IRole
	Group          *types.IGroup
	MasterService  *types.IService
	GroupService   *types.IGroupService
	MasterSchedule *types.ISchedule
	GroupSchedule  *types.IGroupSchedule
	UserSchedule   *types.ISchedule
	Quote          *types.IQuote
	Booking        *types.IBooking
	DateSlots      []*types.IGroupScheduleDateSlots
}

var integrationTest = &IntegrationTest{}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	cmd := exec.Command(filepath.Join("..", os.Getenv("BINARY_NAME")))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	time.Sleep(2 * time.Second)

	code := m.Run()

	time.Sleep(2 * time.Second)

	if err := cmd.Process.Kill(); err != nil {
		fmt.Println("Failed to close server:", err)
	}

	os.Exit(code)
}

func TestIntegrations(t *testing.T) {
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
}
