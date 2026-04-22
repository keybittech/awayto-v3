package main_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationService(t *testing.T) {
	var formId string
	t.Run("admin can create a form and add it to a service", func(tt *testing.T) {
		admin := testutil.IntegrationTest.TestUsers[0]

		groupRoleId := testutil.IntegrationTest.StaffRole.GetId()

		formPayload := map[string]any{
			"name": "test",
			"groupForm": map[string]any{
				"form": map[string]any{
					"name": "test",
					"version": map[string]any{
						"form": map[string]any{
							"0": []map[string]any{
								{
									"i": "37cd98f7-95cc-4c92-b405-309e866afbef",
									"l": "ffffffff",
									"t": "text",
								},
								{
									"i": "78593dae-e8c6-4eef-8b26-abb788dcbb0c",
									"l": "nnnnnnnnnnnn",
									"t": "text",
								},
								{
									"i": "8c56712a-b2a7-4791-80aa-56531bb31107",
									"l": "bnbbbbbbbbbbbbbbbb",
									"t": "multi-select",
									"r": true,
									"v": []any{},
									"o": []map[string]any{
										{
											"i": "ba668bb4-2c7c-4897-bec8-06d1617da55f",
											"l": "aaaaaa",
											"v": "aaaaaa",
										},
										{
											"i": "d2e887fe-c034-4635-b1a9-68a70714f16c",
											"l": "vvvvvvvvv",
											"v": "vvvvvvvvv",
										},
										{
											"i": "6c2640d3-9592-425c-8fb0-6e2268dc0ff2",
											"l": "cccccc",
											"v": "cccccc",
										},
									},
								},
							},
						},
					},
				},
			},
			"groupRoleIds": []string{groupRoleId},
		}

		formRequestBytes, err := json.Marshal(formPayload)
		if err != nil {
			t.Fatalf("error marshalling form request: %v", err)
		}

		postGroupFormResponse := &types.PostGroupFormResponse{}
		err = admin.DoHandler(http.MethodPost, "/api/v1/group/forms", formRequestBytes, nil, postGroupFormResponse)
		if err != nil {
			t.Fatalf("error posting form request: %v", err)
		}

		formId = postGroupFormResponse.Id
		if !util.IsUUID(formId) {
			t.Fatalf("formId is not a uuid: %s", formId)
		}
	})

	t.Run("admin can create service addons and generate a schedule", func(tt *testing.T) {

		admin := testutil.IntegrationTest.TestUsers[0]

		postServiceAddon1Request := &types.PostServiceAddonRequest{Name: "Detailed Discussion"}
		postServiceAddon1RequestBytes, err := protojson.Marshal(postServiceAddon1Request)
		if err != nil {
			t.Fatalf("error marshalling addon 1 request: %v", err)
		}

		postServiceAddon1Response := &types.PostServiceAddonResponse{}
		err = admin.DoHandler(http.MethodPost, "/api/v1/service_addons", postServiceAddon1RequestBytes, nil, postServiceAddon1Response)
		if err != nil {
			t.Fatalf("error requesting addon 1 request: %v", err)
		}

		if !util.IsUUID(postServiceAddon1Response.Id) {
			t.Fatalf("addon 1 id is not a uuid: %s", postServiceAddon1Response.Id)
		}

		postServiceAddon2Request := &types.PostServiceAddonRequest{Name: "Suggestions & Feedback"}
		postServiceAddon2RequestBytes, err := protojson.Marshal(postServiceAddon2Request)
		if err != nil {
			t.Fatalf("error marshalling addon 2 request: %v", err)
		}

		postServiceAddon2Response := &types.PostServiceAddonResponse{}
		err = admin.DoHandler(http.MethodPost, "/api/v1/service_addons", postServiceAddon2RequestBytes, nil, postServiceAddon2Response)
		if err != nil {
			t.Fatalf("error posting addon 2 request: %v", err)
		}

		if !util.IsUUID(postServiceAddon2Response.Id) {
			t.Fatalf("addon 2 id is not a uuid: %s", postServiceAddon2Response.Id)
		}

		serviceAddons := make(map[string]*types.IServiceAddon, 2)
		serviceAddons[postServiceAddon1Response.Id] = &types.IServiceAddon{
			Id:    postServiceAddon1Response.Id,
			Name:  postServiceAddon1Request.Name,
			Order: 1,
		}
		serviceAddons[postServiceAddon2Response.Id] = &types.IServiceAddon{
			Id:    postServiceAddon2Response.Id,
			Name:  postServiceAddon2Request.Name,
			Order: 2,
		}

		tiers := make(map[string]*types.IServiceTier, 1)
		tierId := strconv.Itoa(int(time.Now().UnixMilli()))
		time.Sleep(time.Millisecond)

		tiers[tierId] = &types.IServiceTier{
			Id:        tierId,
			CreatedOn: "test",
			Name:      "General",
			Addons:    serviceAddons,
			Order:     1,
		}

		testutil.IntegrationTest.MasterService = &types.IService{
			Name:      "1-on-1 Tutoring",
			Tiers:     tiers,
			IntakeIds: []string{formId},
			SurveyIds: []string{formId},
		}
	})
}
