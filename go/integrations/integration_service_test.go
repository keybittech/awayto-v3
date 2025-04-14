package main

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationService(t *testing.T) {
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
			CreatedOn: tierId,
			Name:      "test tier",
			Addons:    serviceAddons,
			Order:     1,
		}

		integrationTest.MasterService = &types.IService{
			Name:  "test service",
			Tiers: tiers,
		}
	})
}
