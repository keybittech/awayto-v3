package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

var publicKey *rsa.PublicKey

type TestUser struct {
	TestUserId  int
	TestToken   string
	TestTicket  string
	TestConnId  string
	UserSession *types.UserSession
	Quotes      []*types.IQuote
	Bookings    []*types.IBooking
}

type IntegrationTest struct {
	TestUsers       map[int]*TestUser
	Connections     map[string]net.Conn
	Roles           map[string]*types.IRole
	MemberRole      *types.IRole
	StaffRole       *types.IRole
	Group           *types.IGroup
	MasterService   *types.IService
	GroupService    *types.IGroupService
	MasterSchedule  *types.ISchedule
	MasterSchedules []*types.ISchedule
	GroupSchedule   *types.IGroupSchedule
	GroupSchedules  []*types.IGroupSchedule
	UserSchedule    *types.ISchedule
	UserSchedules   []*types.ISchedule
	DateSlots       []*types.IGroupScheduleDateSlots
	Quotes          []*types.IQuote
	Bookings        []*types.IBooking
}

var integrationTest = &IntegrationTest{}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	kc := &clients.KeycloakClient{
		Server: os.Getenv("KC_INTERNAL"),
		Realm:  os.Getenv("KC_REALM"),
	}

	oidcToken, err := kc.DirectGrantAuthentication()
	if err != nil {
		log.Fatal(err)
	}

	kc.Token = oidcToken

	pk, err := kc.FetchPublicKey()
	if err != nil {
		log.Fatal(err)
	}

	publicKey = pk

	cmd := exec.Command(filepath.Join("../" + os.Getenv("BINARY_NAME")))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	time.Sleep(2 * time.Second)

	code := m.Run()

	jsonBytes, _ := json.MarshalIndent(integrationTest, "", "  ")
	os.WriteFile("integration_results.json", jsonBytes, 0600)

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
