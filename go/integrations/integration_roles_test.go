package main

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationRoles(t *testing.T) {
	t.Run("admin can create roles", func(tt *testing.T) {
		admin := integrationTest.TestUsers[0]

		staffRoleRequest := &types.PostGroupRoleRequest{Name: "Staff"}
		staffRoleRequestBytes, err := protojson.Marshal(staffRoleRequest)
		if err != nil {
			t.Fatalf("error marshalling staff role request: %v", err)
		}

		staffRoleResponse := &types.PostGroupRoleResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/group/roles", staffRoleRequestBytes, nil, staffRoleResponse)
		if err != nil {
			t.Fatalf("error posting staff role request: %v", err)
		}
		if !util.IsUUID(staffRoleResponse.GetGroupRoleId()) {
			t.Fatalf("staff role id is not a uuid: %s", staffRoleResponse.GetGroupRoleId())
		}

		memberRoleRequest := &types.PostGroupRoleRequest{Name: "Member"}
		memberRoleRequestBytes, err := protojson.Marshal(memberRoleRequest)
		if err != nil {
			t.Fatalf("error marshalling member role request: %v", err)
		}

		memberRoleResponse := &types.PostGroupRoleResponse{}
		err = apiRequest(admin.TestToken, http.MethodPost, "/api/v1/group/roles", memberRoleRequestBytes, nil, memberRoleResponse)
		if err != nil {
			t.Fatalf("error posting member role request: %v", err)
		}

		groupRoleId := memberRoleResponse.GetGroupRoleId()

		if !util.IsUUID(groupRoleId) {
			t.Fatalf("member role id is not a uuid: %s", groupRoleId)
		}

		roles := make(map[string]*types.IGroupRole, 2)
		roles[groupRoleId] = &types.IGroupRole{
			Id:   groupRoleId,
			Name: staffRoleRequest.Name,
		}
		roles[groupRoleId] = &types.IGroupRole{
			Id:   groupRoleId,
			Name: memberRoleRequest.Name,
		}

		patchGroupRolesRequest := &types.PatchGroupRolesRequest{
			Roles:         roles,
			DefaultRoleId: groupRoleId,
		}
		patchGroupRolesRequestBytes, err := protojson.Marshal(patchGroupRolesRequest)
		if err != nil {
			t.Fatalf("error marshalling group roles request: %v", err)
		}

		patchGroupRolesResponse := &types.PatchGroupRolesResponse{}
		err = apiRequest(admin.TestToken, http.MethodPatch, "/api/v1/group/roles", patchGroupRolesRequestBytes, nil, patchGroupRolesResponse)
		if err != nil {
			t.Fatalf("error patching group roles request: %v", err)
		}

		integrationTest.Roles = roles
		integrationTest.MemberRole = roles[groupRoleId]
		integrationTest.StaffRole = roles[groupRoleId]
	})
}
