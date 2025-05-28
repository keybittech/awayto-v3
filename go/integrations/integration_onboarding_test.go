package main_test

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func testIntegrationOnboarding(t *testing.T) {
	t.Run("admin can complete onboarding with group, roles, service, schedule", func(tt *testing.T) {
		admin := testutil.IntegrationTest.TestUsers[0]

		onboardingRequest := &types.CompleteOnboardingRequest{
			Service:  testutil.IntegrationTest.MasterService,
			Schedule: testutil.IntegrationTest.MasterSchedule,
		}
		onboardingRequestBytes, err := protojson.Marshal(onboardingRequest)
		if err != nil {
			t.Fatalf("error marshalling onboarding request: %v", err)
		}

		onboardingResponse := &types.CompleteOnboardingResponse{}
		err = admin.DoHandler(http.MethodPost, "/api/v1/group/onboard", onboardingRequestBytes, nil, onboardingResponse)
		if err != nil {
			t.Fatalf("error posting onboarding request: %v", err)
		}

		if !util.IsUUID(onboardingResponse.ServiceId) {
			t.Fatalf("service id is not a uuid: %s", onboardingResponse.ServiceId)
		}

		if !util.IsUUID(onboardingResponse.GroupServiceId) {
			t.Fatalf("group service 2 id is not a uuid: %s", onboardingResponse.GroupServiceId)
		}

		if !util.IsUUID(onboardingResponse.ScheduleId) {
			t.Fatalf("schedule 2 id is not a uuid: %s", onboardingResponse.ScheduleId)
		}

		if !util.IsUUID(onboardingResponse.GroupScheduleId) {
			t.Fatalf("group schedule 2 id is not a uuid: %s", onboardingResponse.GroupScheduleId)
		}

		masterService, err := admin.GetServiceById(onboardingResponse.ServiceId)
		if err != nil {
			t.Fatalf("service by id err: %v", err)
		}

		masterGroupSchedule, err := admin.GetMasterScheduleById(onboardingResponse.ScheduleId)
		if err != nil {
			t.Fatalf("master schedule by id err: %v", err)
		}

		testutil.IntegrationTest.MasterService = masterService
		testutil.IntegrationTest.MasterSchedule = masterGroupSchedule.Schedule

		testutil.IntegrationTest.GroupService = &types.IGroupService{
			Id:      onboardingResponse.GroupServiceId,
			GroupId: testutil.IntegrationTest.Group.Id,
			Service: testutil.IntegrationTest.MasterService,
		}

		testutil.IntegrationTest.GroupSchedule = masterGroupSchedule
	})
}
