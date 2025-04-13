package integrations

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationOnboarding(t *testing.T) {
	t.Run("admin can complete onboarding with group, roles, service, schedule", func(t *testing.T) {
		admin := integrationTest.TestUsers[0]

		onboardingRequest := &types.CompleteOnboardingRequest{
			Service:  integrationTest.MasterService,
			Schedule: integrationTest.MasterSchedule,
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

		masterService, err := getServiceById(admin.TestToken, onboardingResponse.ServiceId)
		if err != nil {
			t.Errorf("service by id err: %v", err)
		}

		masterGroupSchedule, err := getMasterScheduleById(admin.TestToken, onboardingResponse.ScheduleId)
		if err != nil {
			t.Errorf("master schedule by id err: %v", err)
		}

		integrationTest.MasterService = masterService
		integrationTest.MasterSchedule = masterGroupSchedule.Schedule

		integrationTest.GroupService = &types.IGroupService{
			Id:      onboardingResponse.GroupServiceId,
			GroupId: integrationTest.Group.Id,
			Service: integrationTest.MasterService,
		}

		integrationTest.GroupSchedule = masterGroupSchedule
	})

	failCheck(t)
}
