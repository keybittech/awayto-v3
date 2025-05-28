package main_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func testIntegrationUserSchedule(t *testing.T) {
	staff1 := testutil.IntegrationTest.TestUsers[1]
	staff2 := testutil.IntegrationTest.TestUsers[2]
	member1 := testutil.IntegrationTest.TestUsers[4]

	brackets := make(map[string]*types.IScheduleBracket, 1)
	services := make(map[string]*types.IService, 1)
	slots := make(map[string]*types.IScheduleBracketSlot, 2)

	bracketId := strconv.Itoa(int(time.Now().UnixMilli()))
	time.Sleep(time.Millisecond)

	serviceId := testutil.IntegrationTest.MasterService.Id
	services[serviceId] = testutil.IntegrationTest.MasterService

	slot1Id := strconv.Itoa(int(time.Now().UnixMilli()))
	time.Sleep(time.Millisecond)

	slots[slot1Id] = &types.IScheduleBracketSlot{
		Id:                slot1Id,
		StartTime:         "P2DT1H",
		ScheduleBracketId: bracketId,
	}

	slot2Id := strconv.Itoa(int(time.Now().UnixMilli()))
	time.Sleep(time.Millisecond)

	slots[slot2Id] = &types.IScheduleBracketSlot{
		Id:                slot2Id,
		StartTime:         "P3DT4H",
		ScheduleBracketId: bracketId,
	}

	brackets[bracketId] = &types.IScheduleBracket{
		Id:         bracketId,
		Automatic:  false,
		Duration:   15,
		Multiplier: 100,
		Services:   services,
		Slots:      slots,
	}

	t.Run("APP_GROUP_SCHEDULE permission is required to create a personal schedule", func(tt *testing.T) {
		_, err := member1.PostSchedule(&types.PostScheduleRequest{})
		if err == nil || !strings.Contains(err.Error(), "403") {
			t.Fatal("user request to create schedule without permissions was not 403")
		}
	})

	t.Run("user can create a personal schedule using a group schedule id", func(tt *testing.T) {
		schedule, err := staff1.PostSchedule(&types.PostScheduleRequest{
			Brackets:           brackets,
			GroupScheduleId:    testutil.IntegrationTest.MasterSchedule.Id,
			Name:               testutil.IntegrationTest.MasterSchedule.Name,
			StartDate:          testutil.IntegrationTest.MasterSchedule.StartDate,
			EndDate:            testutil.IntegrationTest.MasterSchedule.EndDate,
			ScheduleTimeUnitId: testutil.IntegrationTest.MasterSchedule.ScheduleTimeUnitId,
			BracketTimeUnitId:  testutil.IntegrationTest.MasterSchedule.BracketTimeUnitId,
			SlotTimeUnitId:     testutil.IntegrationTest.MasterSchedule.SlotTimeUnitId,
			SlotDuration:       testutil.IntegrationTest.MasterSchedule.SlotDuration,
		})
		if err != nil {
			t.Fatalf("staff post schedule err %v", err)
		}

		t.Logf("created user schedule with id %s", schedule.Id)

		testutil.IntegrationTest.UserSchedule = schedule
	})

	t.Run("secondary user schedule creation", func(tt *testing.T) {
		schedule, err := staff2.PostSchedule(&types.PostScheduleRequest{
			Brackets:           brackets,
			GroupScheduleId:    testutil.IntegrationTest.MasterSchedules[0].Id,
			Name:               testutil.IntegrationTest.MasterSchedules[0].Name,
			StartDate:          testutil.IntegrationTest.MasterSchedules[0].StartDate,
			EndDate:            testutil.IntegrationTest.MasterSchedules[0].EndDate,
			ScheduleTimeUnitId: testutil.IntegrationTest.MasterSchedules[0].ScheduleTimeUnitId,
			BracketTimeUnitId:  testutil.IntegrationTest.MasterSchedules[0].BracketTimeUnitId,
			SlotTimeUnitId:     testutil.IntegrationTest.MasterSchedules[0].SlotTimeUnitId,
			SlotDuration:       testutil.IntegrationTest.MasterSchedules[0].SlotDuration,
		})
		if err != nil {
			t.Fatalf("secondary staff post schedule err %v", err)
		}

		testutil.IntegrationTest.UserSchedules = append(testutil.IntegrationTest.UserSchedules, schedule)
	})
}
