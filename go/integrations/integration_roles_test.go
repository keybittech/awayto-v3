package integrations

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationRoles(t *testing.T) {
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

	failCheck(t)
}
