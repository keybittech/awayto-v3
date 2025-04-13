package main

import (
	"io"
	"net"
	"os"
	"os/exec"
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
	// Create command to run the main server
	cmd := exec.Command("go", "run", "../main.go")

	// Set up pipes to capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// Copy command output to the test process output in separate goroutines
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	setupSockServer()

	time.Sleep(5 * time.Second)

	code := m.Run()

	cmd.Process.Kill()

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
