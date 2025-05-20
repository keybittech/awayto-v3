package main

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationGroup(t *testing.T) {
	t.Run("admin can create a group", func(t *testing.T) {
		admin := integrationTest.TestUsers[0]
		groupRequest := &types.PostGroupRequest{
			Name:           "the_test_group_" + string(admin.TestUserId),
			DisplayName:    "The Test Group #" + string(admin.TestUserId),
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

		if len(postGroupResponse.Code) != 8 {
			t.Errorf("group code is not 8 length: %s", postGroupResponse.Code)
		}

		integrationTest.Group = &types.IGroup{
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

		if session.RoleBits&types.SiteRoles_APP_GROUP_ADMIN == 0 {
			t.Error("admin doesn't have admin role")
		}

		admin.TestToken = token
		admin.UserSession = session
	})
}
