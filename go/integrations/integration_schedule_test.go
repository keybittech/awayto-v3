package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testIntegrationSchedule(t *testing.T) {
	t.Run("admin can get lookups and generate a schedule", func(t *testing.T) {
		admin := integrationTest.TestUsers[0]

		startTime := time.Date(2023, time.March, 03, 0, 0, 0, 0, time.UTC).Unix()
		endTime := time.Date(2033, time.March, 03, 0, 0, 0, 0, time.UTC).Unix()

		integrationTest.MasterSchedule = &types.ISchedule{
			Name:         "test schedule",
			Timezone:     "America/Los_Angeles",
			SlotDuration: 30,
			StartTime:    &timestamppb.Timestamp{Seconds: startTime},
			EndTime:      &timestamppb.Timestamp{Seconds: endTime},
		}

		lookupsResponse := &types.GetLookupsResponse{}
		err := apiRequest(admin.TestToken, http.MethodGet, "/api/v1/lookup", nil, nil, lookupsResponse)
		if err != nil {
			t.Errorf("error getting lookups request: %v", err)
		}

		if lookupsResponse.TimeUnits == nil {
			t.Error("did not get integration time units")
		}

		for _, tu := range lookupsResponse.TimeUnits {
			if tu.Name == "week" {
				integrationTest.MasterSchedule.ScheduleTimeUnitId = tu.Id
			} else if tu.Name == "hour" {
				integrationTest.MasterSchedule.BracketTimeUnitId = tu.Id
			} else if tu.Name == "minute" {
				integrationTest.MasterSchedule.SlotTimeUnitId = tu.Id
			}
		}
	})

	failCheck(t)
}
