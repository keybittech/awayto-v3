package main_test

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationGroup(t *testing.T) {
	t.Run("admin can create a group", func(tt *testing.T) {
		admin := testutil.IntegrationTest.TestUsers[0]
		groupRequest := &types.PostGroupRequest{
			Name:           "the_test_group_" + string(admin.TestUserId),
			DisplayName:    "The Test Group #" + string(admin.TestUserId),
			Ai:             true,
			Purpose:        "integration testing group",
			AllowedDomains: "",
		}
		requestBytes, err := protojson.Marshal(groupRequest)
		if err != nil {
			t.Fatalf("error marshalling group request: %v", err)
		}

		postGroupResponse := &types.PostGroupResponse{}
		err = admin.DoHandler(http.MethodPost, "/api/v1/group", requestBytes, nil, postGroupResponse)
		if err != nil {
			t.Fatalf("error posting group: %v", err)
		}

		if len(postGroupResponse.Code) != 8 {
			t.Fatalf("group code is not 8 length: %s", postGroupResponse.Code)
		}

		testutil.IntegrationTest.Group = &types.IGroup{
			Code:           postGroupResponse.Code,
			Name:           groupRequest.Name,
			DisplayName:    groupRequest.DisplayName,
			Ai:             groupRequest.Ai,
			Purpose:        groupRequest.Purpose,
			AllowedDomains: groupRequest.AllowedDomains,
		}

		profile, err := admin.GetProfileDetails()
		if err != nil {
			t.Fatal(err)
		}

		if profile.GetRoleBits()&int32(types.SiteRoles_APP_GROUP_ADMIN) == 0 {
			t.Fatal("admin doesn't have admin role")
		}
	})
}
