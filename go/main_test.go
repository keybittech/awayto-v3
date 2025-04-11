package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestUser struct {
	TestUserId  int
	TestToken   string
	TestTicket  string
	TestConnId  string
	UserSession *types.UserSession
}

type IntegrationTest struct {
	TestUsers          map[int]*TestUser
	Connections        map[string]net.Conn
	Roles              map[string]*types.IRole
	DefaultRole        *types.IRole
	Group              *types.IGroup
	OnboardingService  *types.IService
	GroupService       *types.IGroupService
	OnboardingSchedule *types.ISchedule
	GroupSchedule      *types.IGroupSchedule
	UserSchedule       *types.ISchedule
	Quote              *types.IQuote
	Booking            *types.IBooking
}

var integrationTest *IntegrationTest

func init() {
	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	integrationTest = &IntegrationTest{}

	go main()
	setupSockServer()

	exitCode := m.Run()
	os.Exit(exitCode)
}

func Test_main(t *testing.T) {
	if !testing.Short() {
		t.Run("server runs for 5 seconds", func(t *testing.T) {
			time.Sleep(5 * time.Second)
		})
	} else {
		time.Sleep(2 * time.Second)
	}
}

func TestIntegrationUser(t *testing.T) {
	userId := int(time.Now().UnixNano())

	registerKeycloakUserViaForm(userId)

	session, connection, token, ticket, connId := getUser(userId)

	testUser := &TestUser{
		TestUserId:  userId,
		TestToken:   token,
		TestTicket:  ticket,
		TestConnId:  connId,
		UserSession: session,
	}

	integrationTest.TestUsers = make(map[int]*TestUser, 10)
	integrationTest.TestUsers[0] = testUser
	integrationTest.Connections = map[string]net.Conn{session.UserSub: connection}

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationGroup(t *testing.T) {
	admin := integrationTest.TestUsers[0]
	groupRequest := &types.PostGroupRequest{
		Name:           "the_test_group_" + strconv.Itoa(admin.TestUserId),
		DisplayName:    "The Test Group #" + strconv.Itoa(admin.TestUserId),
		Ai:             true,
		Purpose:        "integration testing group",
		AllowedDomains: "",
	}
	requestBytes, err := protojson.Marshal(groupRequest)
	if err != nil {
		t.Fatal(err)
	}

	postGroupResponse := &types.PostGroupResponse{}
	err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/group", requestBytes, nil, postGroupResponse)
	if err != nil {
		t.Fatal(err)
	}

	if postGroupResponse.Id == "" {
		t.Fatal("integration failed to make group")
	}

	integrationTest.Group = &types.IGroup{
		Id:             postGroupResponse.Id,
		Name:           groupRequest.Name,
		DisplayName:    groupRequest.DisplayName,
		Ai:             groupRequest.Ai,
		Purpose:        groupRequest.Purpose,
		AllowedDomains: groupRequest.AllowedDomains,
	}

	token, err := getKeycloakToken(admin.TestUserId)
	if err != nil {
		t.Fatal(err)
	}
	admin.TestToken = token

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationRoles(t *testing.T) {
	admin := integrationTest.TestUsers[0]

	staffRoleRequest := &types.PostRoleRequest{Name: "Staff"}
	staffRoleRequestBytes, err := protojson.Marshal(staffRoleRequest)
	if err != nil {
		t.Fatal(err)
	}

	staffRoleResponse := &types.PostRoleResponse{}
	err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/roles", staffRoleRequestBytes, nil, staffRoleResponse)
	if err != nil {
		t.Fatal(err)
	}

	memberRoleRequest := &types.PostRoleRequest{Name: "Member"}
	memberRoleRequestBytes, err := protojson.Marshal(memberRoleRequest)
	if err != nil {
		t.Fatal(err)
	}

	memberRoleResponse := &types.PostRoleResponse{}
	err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/roles", memberRoleRequestBytes, nil, memberRoleResponse)
	if err != nil {
		t.Fatal(err)
	}

	roles := make(map[string]*types.IRole, 2)
	roles[staffRoleResponse.Id] = &types.IRole{
		Id:   staffRoleResponse.Id,
		Name: staffRoleRequest.Name,
	}
	roles[memberRoleResponse.Id] = &types.IRole{
		Id:   memberRoleResponse.Id,
		Name: memberRoleRequest.Name,
	}

	patchGroupRolesRequest := &types.PatchGroupRolesRequest{
		Roles:         roles,
		DefaultRoleId: memberRoleResponse.Id,
	}
	patchGroupRolesRequestBytes, err := protojson.Marshal(patchGroupRolesRequest)
	if err != nil {
		t.Fatal(err)
	}

	patchGroupRolesResponse := &types.PatchGroupRolesResponse{}
	err = apiRequest(admin.TestToken, http.MethodPatch, "/api/v1/group/roles", patchGroupRolesRequestBytes, nil, patchGroupRolesResponse)
	if err != nil {
		t.Fatal(err)
	}

	integrationTest.Roles = roles
	integrationTest.DefaultRole = roles[staffRoleResponse.Id]

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationService(t *testing.T) {
	admin := integrationTest.TestUsers[0]

	postServiceAddon1Request := &types.PostServiceAddonRequest{Name: "test addon 1"}
	postServiceAddon1RequestBytes, err := protojson.Marshal(postServiceAddon1Request)
	if err != nil {
		t.Fatal(err)
	}
	postServiceAddon1Response := &types.PostServiceAddonResponse{}
	err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/service_addons", postServiceAddon1RequestBytes, nil, postServiceAddon1Response)
	if err != nil {
		t.Fatal(err)
	}

	postServiceAddon2Request := &types.PostServiceAddonRequest{Name: "test addon 2"}
	postServiceAddon2RequestBytes, err := protojson.Marshal(postServiceAddon2Request)
	if err != nil {
		t.Fatal(err)
	}
	postServiceAddon2Response := &types.PostServiceAddonResponse{}
	err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/service_addons", postServiceAddon2RequestBytes, nil, postServiceAddon2Response)
	if err != nil {
		t.Fatal(err)
	}

	serviceAddons := make(map[string]*types.IServiceAddon, 2)
	serviceAddons[postServiceAddon1Response.Id] = &types.IServiceAddon{
		Id:   postServiceAddon1Response.Id,
		Name: postServiceAddon1Request.Name,
	}
	serviceAddons[postServiceAddon2Response.Id] = &types.IServiceAddon{
		Id:   postServiceAddon2Response.Id,
		Name: postServiceAddon2Request.Name,
	}

	tiers := make(map[string]*types.IServiceTier, 1)
	tiers[strconv.Itoa(int(time.Now().UnixNano()))] = &types.IServiceTier{
		Name:   "test tier",
		Addons: serviceAddons,
	}

	integrationTest.OnboardingService = &types.IService{
		Name:  "test service",
		Tiers: tiers,
	}

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationSchedule(t *testing.T) {
	admin := integrationTest.TestUsers[0]

	startTime := time.Date(2023, time.March, 03, 0, 0, 0, 0, time.UTC).Unix()
	endTime := time.Date(2033, time.March, 03, 0, 0, 0, 0, time.UTC).Unix()

	integrationTest.OnboardingSchedule = &types.ISchedule{
		Name:         "test schedule",
		SlotDuration: 30,
		StartTime:    &timestamppb.Timestamp{Seconds: startTime},
		EndTime:      &timestamppb.Timestamp{Seconds: endTime},
	}

	lookupsResponse := &types.GetLookupsResponse{}
	err := apiRequest(admin.TestToken, http.MethodGet, "/api/v1/lookup", nil, nil, lookupsResponse)
	if err != nil {
		t.Fatal(err)
	}

	if lookupsResponse.TimeUnits == nil {
		t.Fatal("did not get integration time units")
	}

	for _, tu := range lookupsResponse.TimeUnits {
		if tu.Name == "week" {
			integrationTest.OnboardingSchedule.ScheduleTimeUnitId = tu.Id
		} else if tu.Name == "hour" {
			integrationTest.OnboardingSchedule.BracketTimeUnitId = tu.Id
		} else if tu.Name == "minute" {
			integrationTest.OnboardingSchedule.SlotTimeUnitId = tu.Id
		}
	}

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationOnboarding(t *testing.T) {
	admin := integrationTest.TestUsers[0]

	onboardingRequest := &types.CompleteOnboardingRequest{
		Service:  integrationTest.OnboardingService,
		Schedule: integrationTest.OnboardingSchedule,
	}
	onboardingRequestBytes, err := protojson.Marshal(onboardingRequest)
	if err != nil {
		t.Fatal(err)
	}

	onboardingResponse := &types.CompleteOnboardingResponse{}
	err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/group/onboard", onboardingRequestBytes, nil, onboardingResponse)
	if err != nil {
		t.Fatal(err)
	}

	integrationTest.OnboardingService.Id = onboardingResponse.ServiceId
	integrationTest.OnboardingSchedule.Id = onboardingResponse.ScheduleId

	integrationTest.GroupService = &types.IGroupService{
		Id:      onboardingResponse.GroupServiceId,
		GroupId: integrationTest.Group.Id,
		Service: integrationTest.OnboardingService,
	}

	integrationTest.GroupSchedule = &types.IGroupSchedule{
		Id:       onboardingResponse.GroupScheduleId,
		GroupId:  integrationTest.Group.Id,
		Schedule: integrationTest.OnboardingSchedule,
	}

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationUserSchedule(t *testing.T) {
	admin := integrationTest.TestUsers[0]

	brackets := make(map[string]*types.IScheduleBracket, 1)

	services := make(map[string]*types.IService, 1)
	services[strconv.Itoa(int(time.Now().UnixNano()))] = integrationTest.GroupService.Service

	bracketId := uuid.NewString()

	slots := make(map[string]*types.IScheduleBracketSlot, 2)
	slots[strconv.Itoa(int(time.Now().UnixNano()))] = &types.IScheduleBracketSlot{
		ScheduleBracketId: bracketId,
		StartTime:         "P2DT1H",
	}
	slots[strconv.Itoa(int(time.Now().UnixNano()))] = &types.IScheduleBracketSlot{
		ScheduleBracketId: bracketId,
		StartTime:         "P3DT4H",
	}

	brackets[bracketId] = &types.IScheduleBracket{
		Id:         bracketId,
		Automatic:  false,
		Duration:   13,
		Multiplier: 100,
		Services:   services,
		Slots:      slots,
	}

	userScheduleRequest := &types.PostScheduleRequest{
		GroupScheduleId:    integrationTest.GroupSchedule.Schedule.Id,
		Brackets:           brackets,
		Name:               integrationTest.OnboardingSchedule.Name,
		StartTime:          integrationTest.OnboardingSchedule.StartTime,
		EndTime:            integrationTest.OnboardingSchedule.EndTime,
		ScheduleTimeUnitId: integrationTest.OnboardingSchedule.ScheduleTimeUnitId,
		BracketTimeUnitId:  integrationTest.OnboardingSchedule.BracketTimeUnitId,
		SlotTimeUnitId:     integrationTest.OnboardingSchedule.SlotTimeUnitId,
		SlotDuration:       integrationTest.OnboardingSchedule.SlotDuration,
	}
	userScheduleRequestBytes, err := protojson.Marshal(userScheduleRequest)
	if err != nil {
		t.Fatal(err)
	}

	userScheduleResponse := &types.PostScheduleResponse{}
	err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/schedules", userScheduleRequestBytes, nil, userScheduleResponse)
	if err != nil {
		t.Fatal(err)
	}

	integrationTest.UserSchedule = integrationTest.OnboardingSchedule
	integrationTest.UserSchedule.Id = userScheduleResponse.Id

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationQuotes(t *testing.T) {

}

func TestIntegrationBookings(t *testing.T) {

}

func TestIntegrationExchange(t *testing.T) {

}

// func BenchmarkBoolFormat(b *testing.B) {
// 	b.ReportAllocs()
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
//
// 		// For true
// 		var v bool = true
// 		var expected rune = 't'
// 		var actual rune
//
// 		// Zero-allocation way to get first char of bool
// 		if v {
// 			actual = 't'
// 		} else {
// 			actual = 'f'
// 		}
//
// 		if actual != expected {
// 			b.Fail()
// 		}
// 	}
// }
//
// func BenchmarkBoolAllocate(b *testing.B) {
// 	b.ReportAllocs()
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		if false {
// 			_ = "t"
// 		}
// 	}
// }
