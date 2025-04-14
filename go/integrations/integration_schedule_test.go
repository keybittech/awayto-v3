package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testIntegrationSchedule(t *testing.T) {
	admin := integrationTest.TestUsers[0]

	var scheduleUnitId, bracketUnitId, slotUnitId string
	name := "test schedule"
	timezone := "America/Los_Angeles"
	slotDuration := int32(30)
	startTime := &timestamppb.Timestamp{
		Seconds: time.Date(2023, time.March, 03, 0, 0, 0, 0, time.UTC).Unix(),
	}
	endTime := &timestamppb.Timestamp{
		Seconds: time.Date(2033, time.March, 03, 0, 0, 0, 0, time.UTC).Unix(),
	}

	t.Run("admin can get lookups and generate a schedule", func(t *testing.T) {

		integrationTest.MasterSchedule = &types.ISchedule{
			Name:         name + " Master",
			Timezone:     timezone,
			SlotDuration: slotDuration,
			StartTime:    startTime,
			EndTime:      endTime,
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
				scheduleUnitId = tu.Id
				integrationTest.MasterSchedule.ScheduleTimeUnitId = tu.Id
			} else if tu.Name == "hour" {
				bracketUnitId = tu.Id
				integrationTest.MasterSchedule.BracketTimeUnitId = tu.Id
			} else if tu.Name == "minute" {
				slotUnitId = tu.Id
				integrationTest.MasterSchedule.SlotTimeUnitId = tu.Id
			}
		}
	})

	t.Run("master schedule can be created", func(t *testing.T) {
		schedule, err := postSchedule(admin.TestToken, &types.PostScheduleRequest{
			AsGroup:            true,
			Name:               name + " Master Creation Test",
			StartTime:          startTime,
			EndTime:            endTime,
			ScheduleTimeUnitId: scheduleUnitId,
			BracketTimeUnitId:  bracketUnitId,
			SlotTimeUnitId:     slotUnitId,
			SlotDuration:       slotDuration,
		})
		if err != nil {
			t.Errorf("failed to post master schedule %v", err)
		}
		if !util.IsUUID(schedule.Id) {
			t.Error("master schedule id is not uuid")
		}

		integrationTest.Schedules = append(integrationTest.Schedules, schedule)
		t.Logf("CREATE MASTER SCHEDULE %s", integrationTest.Schedules[0].Id)
	})

	failCheck(t)
}
