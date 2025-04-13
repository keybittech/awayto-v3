package integrations

import (
	"net"
	"testing"

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

func TestIntegrations(t *testing.T) {
	setupSockServer()

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
