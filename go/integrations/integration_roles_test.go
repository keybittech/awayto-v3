package main_test

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationRoles(t *testing.T) {
	t.Run("admin can create roles", func(tt *testing.T) {
		admin := testutil.IntegrationTest.TestUsers[0]

		staffRoleRequest := &types.PostGroupRoleRequest{Name: "Staff"}
		staffRoleRequestBytes, err := protojson.Marshal(staffRoleRequest)
		if err != nil {
			t.Fatalf("error marshalling staff role request: %v", err)
		}

		roles := make(map[string]*types.IGroupRole, 2)

		staffRoleResponse := &types.PostGroupRoleResponse{}
		err = admin.DoHandler(http.MethodPost, "/api/v1/group/roles", staffRoleRequestBytes, nil, staffRoleResponse)
		if err != nil {
			t.Fatalf("error posting staff role request: %v", err)
		}
		staffRoleId := staffRoleResponse.GetRoleId()
		staffGroupRoleId := staffRoleResponse.GetGroupRoleId()
		if !util.IsUUID(staffRoleId) {
			t.Fatalf("staffRoleId is not a uuid: %s", staffRoleId)
		}
		if !util.IsUUID(staffGroupRoleId) {
			t.Fatalf("staffGroupRoleId is not a uuid: %s", staffGroupRoleId)
		}
		roles[staffGroupRoleId] = &types.IGroupRole{
			Id:     staffGroupRoleId,
			RoleId: staffRoleId,
			Name:   staffRoleRequest.Name,
		}

		memberRoleRequest := &types.PostGroupRoleRequest{Name: "Member"}
		memberRoleRequestBytes, err := protojson.Marshal(memberRoleRequest)
		if err != nil {
			t.Fatalf("error marshalling member role request: %v", err)
		}

		memberRoleResponse := &types.PostGroupRoleResponse{}
		err = admin.DoHandler(http.MethodPost, "/api/v1/group/roles", memberRoleRequestBytes, nil, memberRoleResponse)
		if err != nil {
			t.Fatalf("error posting member role request: %v", err)
		}
		memberRoleId := memberRoleResponse.GetRoleId()
		memberGroupRoleId := memberRoleResponse.GetGroupRoleId()
		if !util.IsUUID(memberRoleId) {
			t.Fatalf("memberRoleId is not a uuid: %s", memberRoleId)
		}
		if !util.IsUUID(memberGroupRoleId) {
			t.Fatalf("memberGroupRoleId is not a uuid: %s", memberGroupRoleId)
		}
		roles[memberGroupRoleId] = &types.IGroupRole{
			Id:     memberGroupRoleId,
			RoleId: memberRoleId,
			Name:   memberRoleRequest.Name,
		}

		patchGroupRolesRequest := &types.PatchGroupRolesRequest{
			Roles:         roles,
			DefaultRoleId: memberRoleId,
		}
		patchGroupRolesRequestBytes, err := protojson.Marshal(patchGroupRolesRequest)
		if err != nil {
			t.Fatalf("error marshalling group roles request: %v", err)
		}

		patchGroupRolesResponse := &types.PatchGroupRolesResponse{}
		err = admin.DoHandler(http.MethodPatch, "/api/v1/group/roles", patchGroupRolesRequestBytes, nil, patchGroupRolesResponse)
		if err != nil {
			t.Fatalf("error patching group roles request: %v", err)
		}

		testutil.IntegrationTest.Roles = roles
		testutil.IntegrationTest.StaffRole = roles[staffGroupRoleId]
		testutil.IntegrationTest.MemberRole = roles[memberGroupRoleId]
	})
}
