package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
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
	MemberRole         *types.IRole
	StaffRole          *types.IRole
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

	integrationTest.TestUsers = make(map[int]*TestUser, 10)
	integrationTest.Connections = make(map[string]net.Conn, 10)

	t.Run("user can register and connect", func(t *testing.T) {
		userId := int(time.Now().UnixNano())
		registerKeycloakUserViaForm(userId)

		session, connection, token, ticket, connId := getUser(userId)

		if !util.IsUUID(session.UserSub) {
			t.Errorf("user sub is not a uuid: %s", session.UserSub)
		}

		testUser := &TestUser{
			TestUserId:  userId,
			TestToken:   token,
			TestTicket:  ticket,
			TestConnId:  connId,
			UserSession: session,
		}

		integrationTest.TestUsers[0] = testUser
		integrationTest.Connections[session.UserSub] = connection

		t.Logf("created user #%d with sub %s connId %s", 1, session.UserSub, connId)

	})

	t.Logf("Integration Test: %+v", integrationTest)
}

func TestIntegrationGroup(t *testing.T) {
	t.Run("admin can create a group", func(t *testing.T) {
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
			t.Errorf("error marshalling group request: %v", err)
		}

		postGroupResponse := &types.PostGroupResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/group", requestBytes, nil, postGroupResponse)
		if err != nil {
			t.Errorf("error posting group: %v", err)
		}

		if !util.IsUUID(postGroupResponse.Id) {
			t.Errorf("group id is not a uuid: %s", postGroupResponse.Id)
		}

		if len(postGroupResponse.Code) != 8 {
			t.Errorf("group id is not 8 length: %s", postGroupResponse.Code)
		}

		integrationTest.Group = &types.IGroup{
			Id:             postGroupResponse.Id,
			Code:           postGroupResponse.Code,
			Name:           groupRequest.Name,
			DisplayName:    groupRequest.DisplayName,
			Ai:             groupRequest.Ai,
			Purpose:        groupRequest.Purpose,
			AllowedDomains: groupRequest.AllowedDomains,
		}

		token, session, err := getKeycloakToken(admin.TestUserId)
		if err != nil {
			t.Error(err)
		}

		if !slices.Contains(session.AvailableUserGroupRoles, types.SiteRoles_APP_GROUP_ADMIN.String()) {
			t.Error("admin doesn't have admin role")
		}

		admin.TestToken = token
		admin.UserSession = session

	})
	t.Logf("Integration Test: %+v", integrationTest)
}

func TestIntegrationRoles(t *testing.T) {
	t.Run("admin can create roles", func(t *testing.T) {
		admin := integrationTest.TestUsers[0]

		staffRoleRequest := &types.PostRoleRequest{Name: "Staff"}
		staffRoleRequestBytes, err := protojson.Marshal(staffRoleRequest)
		if err != nil {
			t.Errorf("error marshalling staff role request: %v", err)
		}

		staffRoleResponse := &types.PostRoleResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/roles", staffRoleRequestBytes, nil, staffRoleResponse)
		if err != nil {
			t.Errorf("error posting staff role request: %v", err)
		}

		if !util.IsUUID(staffRoleResponse.Id) {
			t.Errorf("staff role id is not a uuid: %s", staffRoleResponse.Id)
		}

		memberRoleRequest := &types.PostRoleRequest{Name: "Member"}
		memberRoleRequestBytes, err := protojson.Marshal(memberRoleRequest)
		if err != nil {
			t.Errorf("error marshalling member role request: %v", err)
		}

		memberRoleResponse := &types.PostRoleResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/roles", memberRoleRequestBytes, nil, memberRoleResponse)
		if err != nil {
			t.Errorf("error posting member role request: %v", err)
		}

		if !util.IsUUID(memberRoleResponse.Id) {
			t.Errorf("member role id is not a uuid: %s", memberRoleResponse.Id)
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
			t.Errorf("error marshalling group roles request: %v", err)
		}

		patchGroupRolesResponse := &types.PatchGroupRolesResponse{}
		err = apiRequest(admin.TestToken, http.MethodPatch, "/api/v1/group/roles", patchGroupRolesRequestBytes, nil, patchGroupRolesResponse)
		if err != nil {
			t.Errorf("error patching group roles request: %v", err)
		}

		integrationTest.Roles = roles
		integrationTest.MemberRole = roles[memberRoleResponse.Id]
		integrationTest.StaffRole = roles[staffRoleResponse.Id]
	})

	t.Logf("Integration Test: %+v", integrationTest)
}

func TestIntegrationService(t *testing.T) {
	t.Run("admin can create service addons and generate a schedule", func(t *testing.T) {

		admin := integrationTest.TestUsers[0]

		postServiceAddon1Request := &types.PostServiceAddonRequest{Name: "test addon 1"}
		postServiceAddon1RequestBytes, err := protojson.Marshal(postServiceAddon1Request)
		if err != nil {
			t.Errorf("error marshalling addon 1 request: %v", err)
		}

		postServiceAddon1Response := &types.PostServiceAddonResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/service_addons", postServiceAddon1RequestBytes, nil, postServiceAddon1Response)
		if err != nil {
			t.Errorf("error requesting addon 1 request: %v", err)
		}

		if !util.IsUUID(postServiceAddon1Response.Id) {
			t.Errorf("addon 1 id is not a uuid: %s", postServiceAddon1Response.Id)
		}

		postServiceAddon2Request := &types.PostServiceAddonRequest{Name: "test addon 2"}
		postServiceAddon2RequestBytes, err := protojson.Marshal(postServiceAddon2Request)
		if err != nil {
			t.Errorf("error marshalling addon 2 request: %v", err)
		}

		postServiceAddon2Response := &types.PostServiceAddonResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/service_addons", postServiceAddon2RequestBytes, nil, postServiceAddon2Response)
		if err != nil {
			t.Errorf("error posting addon 2 request: %v", err)
		}

		if !util.IsUUID(postServiceAddon2Response.Id) {
			t.Errorf("addon 2 id is not a uuid: %s", postServiceAddon2Response.Id)
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
	})
	t.Logf("Integration Test: %+v", integrationTest)
}

func TestIntegrationSchedule(t *testing.T) {
	t.Run("admin can get lookups and generate a schedule", func(t *testing.T) {
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
			t.Errorf("error getting lookups request: %v", err)
		}

		if lookupsResponse.TimeUnits == nil {
			t.Error("did not get integration time units")
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
	})
	t.Logf("Integration Test: %+v", integrationTest)
}

func TestIntegrationOnboarding(t *testing.T) {
	t.Run("admin can complete onboarding with group, roles, service, schedule", func(t *testing.T) {
		admin := integrationTest.TestUsers[0]

		onboardingRequest := &types.CompleteOnboardingRequest{
			Service:  integrationTest.OnboardingService,
			Schedule: integrationTest.OnboardingSchedule,
		}
		onboardingRequestBytes, err := protojson.Marshal(onboardingRequest)
		if err != nil {
			t.Errorf("error marshalling onboarding request: %v", err)
		}

		onboardingResponse := &types.CompleteOnboardingResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/group/onboard", onboardingRequestBytes, nil, onboardingResponse)
		if err != nil {
			t.Errorf("error posting onboarding request: %v", err)
		}

		if !util.IsUUID(onboardingResponse.ServiceId) {
			t.Errorf("service id is not a uuid: %s", onboardingResponse.ServiceId)
		}

		if !util.IsUUID(onboardingResponse.GroupServiceId) {
			t.Errorf("group service 2 id is not a uuid: %s", onboardingResponse.GroupServiceId)
		}

		if !util.IsUUID(onboardingResponse.ScheduleId) {
			t.Errorf("schedule 2 id is not a uuid: %s", onboardingResponse.ScheduleId)
		}

		if !util.IsUUID(onboardingResponse.GroupScheduleId) {
			t.Errorf("group schedule 2 id is not a uuid: %s", onboardingResponse.GroupScheduleId)
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
	})
	t.Logf("Integration Test: %+v", integrationTest)
}

func TestIntegrationUserSchedule(t *testing.T) {
	t.Run("user can create a personal schedule using a group schedule id", func(t *testing.T) {
		admin := integrationTest.TestUsers[0]

		brackets := make(map[string]*types.IScheduleBracket, 1)
		services := make(map[string]*types.IService, 1)
		slots := make(map[string]*types.IScheduleBracketSlot, 2)

		bracketId := uuid.NewString()

		services[strconv.Itoa(int(time.Now().UnixNano()))] = integrationTest.GroupService.Service

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

		userScheduleRequestBytes, err := protojson.Marshal(&types.PostScheduleRequest{
			GroupScheduleId:    integrationTest.GroupSchedule.Schedule.Id,
			Brackets:           brackets,
			Name:               integrationTest.OnboardingSchedule.Name,
			StartTime:          integrationTest.OnboardingSchedule.StartTime,
			EndTime:            integrationTest.OnboardingSchedule.EndTime,
			ScheduleTimeUnitId: integrationTest.OnboardingSchedule.ScheduleTimeUnitId,
			BracketTimeUnitId:  integrationTest.OnboardingSchedule.BracketTimeUnitId,
			SlotTimeUnitId:     integrationTest.OnboardingSchedule.SlotTimeUnitId,
			SlotDuration:       integrationTest.OnboardingSchedule.SlotDuration,
		})
		if err != nil {
			t.Errorf("error marshalling user schedule request: %v", err)
		}

		userScheduleResponse := &types.PostScheduleResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/schedules", userScheduleRequestBytes, nil, userScheduleResponse)
		if err != nil {
			t.Errorf("error posting user schedule request: %v", err)
		}

		integrationTest.UserSchedule = integrationTest.OnboardingSchedule
		integrationTest.UserSchedule.Id = userScheduleResponse.Id
	})
	t.Logf("Integration Test: %+v", integrationTest)
}

func TestIntegrationJoinGroup(t *testing.T) {
	existingUsers := 1
	t.Run("users can join a group with a code after log in", func(t *testing.T) {
		for c := existingUsers; c < existingUsers+6; c++ {
			joinViaRegister := c%2 == 0
			userId := int(time.Now().UnixNano())

			t.Logf("Registering #%d", userId)
			if joinViaRegister {
				// This takes care of attach user to group and activate profile on the backend
				registerKeycloakUserViaForm(userId, integrationTest.Group.Code)
			} else {
				registerKeycloakUserViaForm(userId)
			}
			time.Sleep(time.Second)

			session, connection, token, ticket, connId := getUser(userId)
			t.Logf("got connections %v %s %s", session, ticket, connId)

			if !util.IsUUID(session.UserSub) {
				t.Errorf("user sub is not a uuid: %s", session.UserSub)
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
				time.Sleep(time.Second)

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
				time.Sleep(time.Second)

				// Activate Profile -- lets the user view the internal login pages
				activateProfileResponse := &types.ActivateProfileResponse{}
				err = apiRequest(token, http.MethodPatch, "/api/v1/profile/activate", nil, nil, activateProfileResponse)
				if err != nil {
					t.Errorf("error patch activate profile group request, user: %d error: %v", c, err)
				}
				if !activateProfileResponse.Success {
					t.Errorf("activate profile internal was unsuccessful %v", activateProfileResponse)
				}
				time.Sleep(time.Second)

				// Get new token after group setup to check group membersip
				token, session, err = getKeycloakToken(userId)
				if err != nil {
					t.Errorf("failed to get new token after joining group %v", err)
				}
			}

			if len(session.SubGroups) == 0 {
				t.Errorf("no group id after getting new token %v", session)
			}

			testUser := &TestUser{
				TestUserId:  userId,
				TestToken:   token,
				TestTicket:  ticket,
				TestConnId:  connId,
				UserSession: session,
			}

			t.Logf("created user #%d with sub %s connId %s", c, session.UserSub, connId)

			integrationTest.TestUsers[c] = testUser
			integrationTest.Connections[session.UserSub] = connection

			t.Logf("user %d has sub %s", c, integrationTest.TestUsers[c].UserSession.UserSub)
		}
	})
}

func TestIntegrationPromoteUser(t *testing.T) {
	admin := integrationTest.TestUsers[0]
	staff1 := integrationTest.TestUsers[1]
	staff2 := integrationTest.TestUsers[2]
	member1 := integrationTest.TestUsers[4]

	staffRoleFullName := "/" + integrationTest.Group.Name + "/" + integrationTest.StaffRole.Name
	memberRoleFullName := "/" + integrationTest.Group.Name + "/" + integrationTest.MemberRole.Name

	t.Log("initialize group assignments cache")
	err := apiRequest(admin.TestToken, http.MethodGet, "/api/v1/group/assignments", nil, nil, nil)
	if err != nil {
		t.Errorf("error get group assignments request: %v", err)
	}

	t.Run("user cannot update role permissions", func(t *testing.T) {
		t.Logf("user add admin ability to member role %s", memberRoleFullName)
		err := patchGroupAssignments(member1.TestToken, memberRoleFullName, types.SiteRoles_APP_GROUP_ADMIN.String())
		if err == nil {
			t.Errorf("user was able to add admin to their own role %v", err)
		} else {
			t.Logf("failed to patch admin to member as member %v", err)
		}
	})

	t.Run("admin can update role permissions", func(t *testing.T) {
		t.Logf("admin add scheduling ability to staff role %s", staffRoleFullName)
		err := patchGroupAssignments(admin.TestToken, staffRoleFullName, types.SiteRoles_APP_GROUP_SCHEDULES.String())
		if err != nil {
			t.Errorf("admin add scheduling ability to staff err %v", err)
		}

		t.Logf("admin add booking ability to member role %s", memberRoleFullName)
		err = patchGroupAssignments(admin.TestToken, memberRoleFullName, types.SiteRoles_APP_GROUP_BOOKINGS.String())
		if err != nil {
			t.Errorf("admin add booking ability to member err %v", err)
		}
	})

	t.Run("user cannot change their own role", func(t *testing.T) {
		t.Log("user attempt to modify their own role to staff")
		err := patchGroupUser(member1.TestToken, member1.UserSession.UserSub, integrationTest.StaffRole.Id)
		if err != nil && !strings.Contains(err.Error(), "Forbidden") {
			t.Errorf("user patches self err %v", err)
		}

		_, session, err := getKeycloakToken(member1.TestUserId)
		if err != nil {
			t.Errorf("failed to get new token after user patches self %v", err)
		}

		if slices.Contains(session.AvailableUserGroupRoles, types.SiteRoles_APP_GROUP_ADMIN.String()) {
			t.Error("user has admin role after trying to add admin role to themselves")
		}
		time.Sleep(time.Second)
	})

	t.Run("admin can promote users to staff", func(t *testing.T) {
		t.Log("admin attempt to modify user to staff")
		err := patchGroupUser(admin.TestToken, staff1.UserSession.UserSub, integrationTest.StaffRole.Id)
		if err != nil {
			t.Errorf("admin promotes staff err %v", err)
		}

		_, session, err := getKeycloakToken(staff1.TestUserId)
		if err != nil {
			t.Errorf("failed to get new token after admin promotes staff %v", err)
		}

		if !slices.Contains(session.AvailableUserGroupRoles, types.SiteRoles_APP_GROUP_SCHEDULES.String()) {
			t.Error("staff does not have APP_GROUP_SCHEDULES after admin promotion")
		}
		time.Sleep(time.Second)
	})

	t.Run("APP_GROUP_USERS permission allows user role changes", func(t *testing.T) {
		t.Log("staff attempt promote staff without permissions")
		err := patchGroupUser(staff1.TestToken, staff2.UserSession.UserSub, integrationTest.StaffRole.Id)
		if err != nil && !strings.Contains(err.Error(), "Forbidden") {
			t.Errorf("staff promotes staff without permissions err %v", err)
		}

		_, session, err := getKeycloakToken(staff2.TestUserId)
		if err != nil {
			t.Errorf("failed to get new token after staff modify staff role without permissions %v", err)
		}

		if slices.Contains(session.AvailableUserGroupRoles, types.SiteRoles_APP_GROUP_SCHEDULES.String()) {
			t.Error("staff modified staff role without having APP_GROUP_USERS permissions")
		}

		t.Logf("admin add APP_GROUP_USERS to staff role %s", staffRoleFullName)
		err = patchGroupAssignments(admin.TestToken, staffRoleFullName, types.SiteRoles_APP_GROUP_USERS.String())
		if err != nil {
			t.Errorf("admin add user editing ability to staff err %v", err)
		}

		token, session, err := getKeycloakToken(staff1.TestUserId)
		if err != nil {
			t.Errorf("failed to get new token after staff modify user role without permissions %v", err)
		}

		if !slices.Contains(session.AvailableUserGroupRoles, types.SiteRoles_APP_GROUP_USERS.String()) {
			t.Error("staff does not have APP_GROUP_USERS permissions after admin add")
		}

		staff1.TestToken = token

		t.Log("staff attempt promote user with permissions")
		err = patchGroupUser(staff1.TestToken, staff2.UserSession.UserSub, integrationTest.StaffRole.Id)
		if err != nil {
			t.Errorf("staff promotes user with permissions err %v", err)
		}

		_, session, err = getKeycloakToken(staff2.TestUserId)
		if err != nil {
			t.Errorf("failed to get new token after staff modify user role with permissions %v", err)
		}

		if !slices.Contains(session.AvailableUserGroupRoles, types.SiteRoles_APP_GROUP_SCHEDULES.String()) {
			t.Error("staff failed modify user role having APP_GROUP_USERS permissions")
		}
	})
}

// await patchGroupUser({ patchGroupUserRequest: { userId: id, roleId, roleName: name } }).unwrap();
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
