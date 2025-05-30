package main_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testIntegrationSchedule(t *testing.T) {
	admin := testutil.IntegrationTest.TestUsers[0]

	var scheduleUnitId, bracketUnitId, slotUnitId string
	name := "test schedule"
	timezone := "America/Los_Angeles"
	slotDuration := int32(30)
	st := &timestamppb.Timestamp{
		Seconds: time.Date(2023, time.March, 03, 0, 0, 0, 0, time.UTC).Unix(),
	}
	startDate := st.AsTime().Format(time.RFC3339)

	et := &timestamppb.Timestamp{
		Seconds: time.Date(2033, time.March, 03, 0, 0, 0, 0, time.UTC).Unix(),
	}
	endDate := et.AsTime().Format(time.RFC3339)

	t.Run("admin can get lookups and generate a schedule", func(tt *testing.T) {
		testutil.IntegrationTest.MasterSchedule = &types.ISchedule{
			Name:         name + " Master",
			Timezone:     timezone,
			SlotDuration: slotDuration,
			StartDate:    startDate,
			EndDate:      endDate,
		}

		lookupsResponse := &types.GetLookupsResponse{}
		err := admin.DoHandler(http.MethodGet, "/api/v1/lookup", nil, nil, lookupsResponse)
		if err != nil {
			t.Fatalf("error getting lookups request: %v", err)
		}

		if lookupsResponse.TimeUnits == nil {
			t.Fatalf("did not get integration time units")
		}

		for _, tu := range lookupsResponse.TimeUnits {
			if tu.Name == "week" {
				scheduleUnitId = tu.Id
				testutil.IntegrationTest.MasterSchedule.ScheduleTimeUnitId = tu.Id
			} else if tu.Name == "hour" {
				bracketUnitId = tu.Id
				testutil.IntegrationTest.MasterSchedule.BracketTimeUnitId = tu.Id
			} else if tu.Name == "minute" {
				slotUnitId = tu.Id
				testutil.IntegrationTest.MasterSchedule.SlotTimeUnitId = tu.Id
			}
		}
	})

	t.Run("master schedule can be created and attached to the group", func(tt *testing.T) {
		schedule, err := admin.PostSchedule(&types.PostScheduleRequest{
			AsGroup:            true,
			Name:               name + " Master Creation Test",
			StartDate:          startDate,
			EndDate:            endDate,
			ScheduleTimeUnitId: scheduleUnitId,
			BracketTimeUnitId:  bracketUnitId,
			SlotTimeUnitId:     slotUnitId,
			SlotDuration:       slotDuration,
		})
		if err != nil {
			t.Fatalf("failed to post master schedule %v", err)
		}
		if !util.IsUUID(schedule.Id) {
			t.Fatalf("master schedule id is not uuid")
		}

		err = admin.PostGroupSchedule(schedule.Id)
		if err != nil {
			t.Fatalf("master schedule creation attach group err: %v", err)
		}

		groupMasterSchedule, err := admin.GetMasterScheduleById(schedule.Id)
		if err != nil {
			t.Fatalf("master schedule creation err: %v", err)
		}
		if groupMasterSchedule.Schedule.Id == "" {
			t.Fatalf("no master schedule > schedule id: %v", groupMasterSchedule.Schedule)
		}

		testutil.IntegrationTest.MasterSchedules = append(testutil.IntegrationTest.MasterSchedules, groupMasterSchedule.Schedule)
		testutil.IntegrationTest.GroupSchedules = append(testutil.IntegrationTest.GroupSchedules, groupMasterSchedule)
	})
}
