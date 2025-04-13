package main

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func testIntegrationPromoteUser(t *testing.T) {
	admin := integrationTest.TestUsers[0]
	staff1 := integrationTest.TestUsers[1]
	staff2 := integrationTest.TestUsers[2]
	staff3 := integrationTest.TestUsers[3]
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
	})

	t.Run("APP_GROUP_USERS permission allows user role changes", func(t *testing.T) {
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

		err = patchGroupUser(staff1.TestToken, staff2.UserSession.UserSub, integrationTest.StaffRole.Id)
		if err != nil {
			t.Errorf("staff promotes user with permissions err %v", err)
		}

		err = patchGroupUser(admin.TestToken, staff3.UserSession.UserSub, integrationTest.StaffRole.Id)
		if err != nil {
			t.Errorf("staff promotes other user with permissions err %v", err)
		}

		token, session, err = getKeycloakToken(staff2.TestUserId)
		if err != nil {
			t.Errorf("failed to get new token after staff modify user role with permissions %v", err)
		}

		if !slices.Contains(session.AvailableUserGroupRoles, types.SiteRoles_APP_GROUP_SCHEDULES.String()) {
			t.Error("staff failed modify user role having APP_GROUP_USERS permissions")
		}

		staff2.TestToken = token
	})

	t.Run("new roles have been assigned", func(t *testing.T) {
		// Update TestUser states
		staffs := make([]*TestUser, 3)
		staffs[0] = integrationTest.TestUsers[1]
		staffs[1] = integrationTest.TestUsers[2]
		staffs[2] = integrationTest.TestUsers[3]
		for i, staff := range staffs {
			token, session, err := getKeycloakToken(staff.TestUserId)
			if err != nil {
				t.Errorf("failed to get new staff token after role update staffid:%d %v", staff.TestUserId, err)
			}
			if strings.Index(session.Roles, types.SiteRoles_APP_GROUP_SCHEDULES.String()) == -1 {
				t.Errorf("staff %d does not have APP_GROUP_SCHEDULES permissions, %s", staff.TestUserId, session.Roles)
			}
			if strings.Index(session.Roles, types.SiteRoles_APP_GROUP_USERS.String()) == -1 {
				t.Errorf("staff %d does not have APP_GROUP_USERS permissions, %s", staff.TestUserId, session.Roles)
			}
			integrationTest.TestUsers[i+1].TestToken = token
			integrationTest.TestUsers[i+1].UserSession = session
		}

		members := make([]*TestUser, 3)
		members[0] = integrationTest.TestUsers[4]
		members[1] = integrationTest.TestUsers[5]
		members[2] = integrationTest.TestUsers[6]
		for i, member := range members {
			token, session, err := getKeycloakToken(member.TestUserId)
			if err != nil {
				t.Errorf("failed to get new member token after role update memberid:%d %v", member.TestUserId, err)
			}
			if strings.Index(session.Roles, types.SiteRoles_APP_GROUP_BOOKINGS.String()) == -1 {
				t.Errorf("member %d does not have APP_GROUP_BOOKINGS permissions, %s", member.TestUserId, session.Roles)
			}
			integrationTest.TestUsers[i+4].TestToken = token
			integrationTest.TestUsers[i+4].UserSession = session
		}
	})

	failCheck(t)
}
