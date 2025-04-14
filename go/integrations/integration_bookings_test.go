package main

import (
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testIntegrationBookings(t *testing.T) {
	admin := integrationTest.TestUsers[0]
	staff1 := integrationTest.TestUsers[1]
	member1 := integrationTest.TestUsers[4]
	member2 := integrationTest.TestUsers[5]
	member3 := integrationTest.TestUsers[6]

	var serviceTierId string
	for _, tier := range integrationTest.GroupService.Service.Tiers {
		serviceTierId = tier.Id
		break
	}

	t.Run("APP_GROUP_SCHEDULES is required by approver", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[0].Id,
				SlotDate:              member1.Quotes[0].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
			},
		}

		_, err := postBooking(member1.TestToken, bookingRequests)
		if err != nil && !strings.Contains(err.Error(), "Forbidden") {
			t.Errorf("booking post without permissions was not Forbidden, %v", err)
		}
	})

	t.Run("slot must be owned by approver", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[0].Id,
				SlotDate:              member1.Quotes[0].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
			},
		}

		_, err := postBooking(admin.TestToken, bookingRequests)
		if err == nil {
			t.Error("slot was approved by non-owner approver")
		}
	})

	t.Run("different sbsids cannot be batch approved", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 2)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[0].Id,
				SlotDate:              member1.Quotes[0].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
			},
		}
		bookingRequests[1] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[1].Id,
				SlotDate:              member1.Quotes[1].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[1].ScheduleBracketSlotId,
			},
		}

		_, err := postBooking(staff1.TestToken, bookingRequests)
		if err == nil {
			t.Errorf("different booking approve was allowed, id 1 %s, id 2 %s", member1.Quotes[0].ScheduleBracketSlotId, member1.Quotes[1].ScheduleBracketSlotId)
		}
	})

	t.Run("staff can create a booking from quote info, reserving slots", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[0].Id,
				SlotDate:              member1.Quotes[0].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
			},
		}
		bookings, err := postBooking(staff1.TestToken, bookingRequests)
		if err != nil {
			t.Errorf("single booking approve error %v", err)
		}

		if len(bookings) != 1 {
			t.Errorf("expected %d bookings, received %d", 1, len(bookings))
		}
	})

	t.Run("reserved slots are not usable in future quotes", func(t *testing.T) {
		reservedSlot := &types.IGroupScheduleDateSlots{
			StartDate:             member1.Quotes[0].SlotDate,
			ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
		}
		_, err := postQuote(member3.TestToken, serviceTierId, reservedSlot, nil, nil)
		if err == nil {
			t.Errorf("user was able to request a booked slot")
		}
	})

	t.Run("slots can be batch approved", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 2)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[1].Id,
				SlotDate:              member1.Quotes[1].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[1].ScheduleBracketSlotId,
			},
		}
		bookingRequests[1] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member2.Quotes[1].Id,
				SlotDate:              member2.Quotes[1].SlotDate,
				ScheduleBracketSlotId: member2.Quotes[1].ScheduleBracketSlotId,
			},
		}

		bookings, err := postBooking(staff1.TestToken, bookingRequests)
		if err != nil {
			t.Errorf("batch booking approve error %v", err)
		}

		if len(bookings) != 2 {
			t.Errorf("expected %d bookings, received %d", 2, len(bookings))
		}

		if !util.IsUUID(bookings[0].Id) {
			t.Errorf("booking id 1 is not a uuid %s", bookings[0].Id)
		}

		if !util.IsUUID(bookings[1].Id) {
			t.Errorf("booking id 2 is not a uuid %s", bookings[1].Id)
		}
	})

	t.Run("master schedules can be disabled, preserving all records", func(t *testing.T) {

	})

	t.Run("master schedules can be deleted, destroying all records", func(t *testing.T) {

	})

	failCheck(t)
}
