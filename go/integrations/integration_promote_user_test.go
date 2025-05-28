package main_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func testIntegrationPromoteUser(t *testing.T) {
	admin := testutil.IntegrationTest.TestUsers[0]
	staff1 := testutil.IntegrationTest.TestUsers[1]
	staff2 := testutil.IntegrationTest.TestUsers[2]
	staff3 := testutil.IntegrationTest.TestUsers[3]
	member1 := testutil.IntegrationTest.TestUsers[4]

	staffRoleFullName := "/" + testutil.IntegrationTest.Group.Name + "/" + testutil.IntegrationTest.StaffRole.Name
	memberRoleFullName := "/" + testutil.IntegrationTest.Group.Name + "/" + testutil.IntegrationTest.MemberRole.Name

	var groupUsers []*types.IGroupUser
	t.Run("admin can list users", func(tt *testing.T) {
		groupUserResponse := &types.GetGroupUsersResponse{}
		err := admin.DoHandler(http.MethodGet, "/api/v1/group/users", nil, nil, groupUserResponse)
		if err != nil {
			t.Fatalf("error get group assignments request: %v", err)
		}
		groupUsers = groupUserResponse.GetGroupUsers()
	})

	getSubByUserEmail := func(userEmail string) string {
		for _, gu := range groupUsers {
			if gu.GetUserProfile().GetEmail() == userEmail {
				return gu.GetUserProfile().GetSub()
			}
		}
		return ""
	}

	t.Run("role assignments can be retrieved", func(tt *testing.T) {
		err := admin.DoHandler(http.MethodGet, "/api/v1/group/assignments", nil, nil, nil)
		if err != nil {
			t.Fatalf("error get group assignments request: %v", err)
		}
	})

	t.Run("user cannot update role permissions", func(tt *testing.T) {
		t.Logf("user add admin ability to member role %s", memberRoleFullName)
		err := member1.PatchGroupAssignments(memberRoleFullName, types.SiteRoles_APP_GROUP_ADMIN.String())
		if err == nil {
			t.Fatalf("user was able to add admin to their own role %v", err)
		} else {
			t.Logf("failed to patch admin to member as member %v", err)
		}
	})

	t.Run("admin can update role permissions", func(tt *testing.T) {
		t.Logf("admin add scheduling ability to staff role %s", staffRoleFullName)
		err := admin.PatchGroupAssignments(staffRoleFullName, types.SiteRoles_APP_GROUP_SCHEDULES.String())
		if err != nil {
			t.Fatalf("admin add scheduling ability to staff err %v", err)
		}

		t.Logf("admin add booking ability to member role %s", memberRoleFullName)
		err = admin.PatchGroupAssignments(memberRoleFullName, types.SiteRoles_APP_GROUP_BOOKINGS.String())
		if err != nil {
			t.Fatalf("admin add booking ability to member err %v", err)
		}
	})

	t.Run("user cannot change their own role", func(tt *testing.T) {
		t.Log("user attempt to modify their own role to staff")
		member1Sub := getSubByUserEmail(member1.GetTestEmail())
		err := member1.PatchGroupUser(member1Sub, testutil.IntegrationTest.StaffRole.GetRoleId())
		if err != nil && !strings.Contains(err.Error(), "403") {
			t.Fatalf("user patches self err %v", err)
		}

		profile, err := member1.GetProfileDetails()
		if err != nil {
			t.Fatalf("failed to get new profile after user patches self %v", err)
		}

		if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_ADMIN) > 0 {
			t.Fatal("user has admin role after trying to add admin role to themselves")
		}
	})

	t.Run("admin can promote users to staff", func(tt *testing.T) {
		t.Log("admin attempt to modify user to staff")
		staff1Sub := getSubByUserEmail(staff1.GetTestEmail())
		err := admin.PatchGroupUser(staff1Sub, testutil.IntegrationTest.GetStaffRole().GetRoleId())
		if err != nil {
			t.Fatalf("admin promotes staff err %v", err)
		}

		profile, err := staff1.GetProfileDetails()
		if err != nil {
			t.Fatalf("failed to get new details after admin promotes staff %v", err)
		}

		if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_SCHEDULES) == 0 {
			t.Fatal("staff does not have APP_GROUP_SCHEDULES after admin promotion")
		}
	})

	t.Run("APP_GROUP_USERS permission allows user role changes", func(tt *testing.T) {
		staff2Sub := getSubByUserEmail(staff2.GetTestEmail())
		err := staff1.PatchGroupUser(staff2Sub, testutil.IntegrationTest.StaffRole.GetRoleId())
		if err != nil && !strings.Contains(err.Error(), "403") {
			t.Fatalf("staff promotes staff without permissions was not 403: %v", err)
		}

		profile, err := staff2.GetProfileDetails()
		if err != nil {
			t.Fatalf("failed to get new details after staff modify staff role without permissions %v", err)
		}

		if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_SCHEDULES) > 0 {
			t.Fatal("staff modified staff role without having APP_GROUP_USERS permissions")
		}

		err = admin.PatchGroupAssignments(staffRoleFullName, types.SiteRoles_APP_GROUP_USERS.String())
		if err != nil {
			t.Fatalf("admin add user editing ability to staff err %v", err)
		}

		profile, err = staff1.GetProfileDetails()
		if err != nil {
			t.Fatalf("failed to get new details after admin modify staff role %v", err)
		}

		if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_USERS) == 0 {
			t.Fatal("staff does not have APP_GROUP_USERS permissions after admin add")
		}

		err = staff1.PatchGroupUser(staff2Sub, testutil.IntegrationTest.StaffRole.GetRoleId())
		if err != nil {
			t.Fatalf("staff promotes user with permissions err %v", err)
		}

		profile, err = staff2.GetProfileDetails()
		if err != nil {
			t.Fatalf("failed to get new token after staff modify user role with permissions %v", err)
		}

		if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_SCHEDULES) == 0 {
			t.Fatal("staff failed modify user role having APP_GROUP_USERS permissions")
		}

		staff3Sub := getSubByUserEmail(staff3.GetTestEmail())
		err = staff1.PatchGroupUser(staff3Sub, testutil.IntegrationTest.StaffRole.GetRoleId())
		if err != nil {
			t.Fatalf("staff promotes other user with permissions err %v", err)
		}
	})

	t.Run("new roles have been assigned", func(tt *testing.T) {
		// Update TestUser states
		staffs := make([]*testutil.TestUsersStruct, 3)
		staffs[0] = testutil.IntegrationTest.TestUsers[1]
		staffs[1] = testutil.IntegrationTest.TestUsers[2]
		staffs[2] = testutil.IntegrationTest.TestUsers[3]
		for _, staff := range staffs {
			profile, err := staff.GetProfileDetails()
			if err != nil {
				t.Fatalf("failed to get new staff details after role update staffid:%s %v", staff.TestUserId, err)
			}
			if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_SCHEDULES) == 0 {
				t.Fatalf("staff %s does not have APP_GROUP_SCHEDULES permissions, %d", staff.TestUserId, profile.GetRoleBits())
			}
			if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_USERS) == 0 {
				t.Fatalf("staff %s does not have APP_GROUP_USERS permissions, %d", staff.TestUserId, profile.GetRoleBits())
			}
			staff.Profile = profile
		}

		// Everyone starts with member role so these should now have APP_GROUP_BOOKINGS
		members := make([]*testutil.TestUsersStruct, 3)
		members[0] = testutil.IntegrationTest.TestUsers[4]
		members[1] = testutil.IntegrationTest.TestUsers[5]
		members[2] = testutil.IntegrationTest.TestUsers[6]
		for _, member := range members {
			profile, err := member.GetProfileDetails()
			if err != nil {
				t.Fatalf("failed to get new member token after role update memberid:%s %v", member.TestUserId, err)
			}
			if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_BOOKINGS) == 0 {
				t.Fatalf("member %s does not have APP_GROUP_BOOKINGS permissions, %d", member.TestUserId, profile.GetRoleBits())
			}
			member.Profile = profile
		}
	})
}
