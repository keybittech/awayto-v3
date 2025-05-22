package main

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func testIntegrationUserSchedule(t *testing.T) {
	staff1 := integrationTest.TestUsers[1]
	staff2 := integrationTest.TestUsers[2]
	member1 := integrationTest.TestUsers[4]

	brackets := make(map[string]*types.IScheduleBracket, 1)
	services := make(map[string]*types.IService, 1)
	slots := make(map[string]*types.IScheduleBracketSlot, 2)

	bracketId := strconv.Itoa(int(time.Now().UnixMilli()))
	time.Sleep(time.Millisecond)

	serviceId := integrationTest.MasterService.Id
	services[serviceId] = integrationTest.MasterService

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
		_, err := postSchedule(member1.TestToken, &types.PostScheduleRequest{})
		if err == nil || !strings.Contains(err.Error(), "403") {
			t.Fatal("user request to create schedule without permissions was not 403")
		}
	})

	t.Run("user can create a personal schedule using a group schedule id", func(tt *testing.T) {
		schedule, err := postSchedule(staff1.TestToken, &types.PostScheduleRequest{
			Brackets:           brackets,
			GroupScheduleId:    integrationTest.MasterSchedule.Id,
			Name:               integrationTest.MasterSchedule.Name,
			StartDate:          integrationTest.MasterSchedule.StartDate,
			EndDate:            integrationTest.MasterSchedule.EndDate,
			ScheduleTimeUnitId: integrationTest.MasterSchedule.ScheduleTimeUnitId,
			BracketTimeUnitId:  integrationTest.MasterSchedule.BracketTimeUnitId,
			SlotTimeUnitId:     integrationTest.MasterSchedule.SlotTimeUnitId,
			SlotDuration:       integrationTest.MasterSchedule.SlotDuration,
		})
		if err != nil {
			t.Fatalf("staff post schedule err %v", err)
		}

		t.Logf("created user schedule with id %s", schedule.Id)

		integrationTest.UserSchedule = schedule
	})

	t.Run("secondary user schedule creation", func(tt *testing.T) {
		schedule, err := postSchedule(staff2.TestToken, &types.PostScheduleRequest{
			Brackets:           brackets,
			GroupScheduleId:    integrationTest.MasterSchedules[0].Id,
			Name:               integrationTest.MasterSchedules[0].Name,
			StartDate:          integrationTest.MasterSchedules[0].StartDate,
			EndDate:            integrationTest.MasterSchedules[0].EndDate,
			ScheduleTimeUnitId: integrationTest.MasterSchedules[0].ScheduleTimeUnitId,
			BracketTimeUnitId:  integrationTest.MasterSchedules[0].BracketTimeUnitId,
			SlotTimeUnitId:     integrationTest.MasterSchedules[0].SlotTimeUnitId,
			SlotDuration:       integrationTest.MasterSchedules[0].SlotDuration,
		})
		if err != nil {
			t.Fatalf("secondary staff post schedule err %v", err)
		}

		integrationTest.UserSchedules = append(integrationTest.UserSchedules, schedule)
	})
}
