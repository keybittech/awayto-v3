package main_test

import (
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testIntegrationBookings(t *testing.T) {
	admin := testutil.IntegrationTest.TestUsers[0]
	staff1 := testutil.IntegrationTest.TestUsers[1]
	member1 := testutil.IntegrationTest.TestUsers[4]
	member2 := testutil.IntegrationTest.TestUsers[5]
	member3 := testutil.IntegrationTest.TestUsers[6]

	var serviceTierId string
	for _, tier := range testutil.IntegrationTest.GroupService.Service.Tiers {
		serviceTierId = tier.Id
		break
	}

	t.Run("APP_GROUP_SCHEDULES is required by approver", func(tt *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[0].Id,
				SlotDate:              member1.Quotes[0].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
			},
		}

		_, err := member1.PostBooking(bookingRequests)
		if err != nil && !strings.Contains(err.Error(), "403") {
			t.Fatalf("booking post without permissions was not 403, %v", err)
		}
	})

	t.Run("slot must be owned by approver", func(tt *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[0].Id,
				SlotDate:              member1.Quotes[0].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
			},
		}

		_, err := admin.PostBooking(bookingRequests)
		if err == nil {
			t.Fatalf("slot was approved by non-owner approver")
		}
	})

	t.Run("different sbsids cannot be batch approved", func(tt *testing.T) {
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

		_, err := staff1.PostBooking(bookingRequests)
		if err == nil {
			t.Fatalf("different booking approve was allowed, id 1 %s, id 2 %s", member1.Quotes[0].ScheduleBracketSlotId, member1.Quotes[1].ScheduleBracketSlotId)
		}
	})

	t.Run("staff can create a booking from quote info, reserving slots", func(tt *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quotes[0].Id,
				SlotDate:              member1.Quotes[0].SlotDate,
				ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
			},
		}
		bookings, err := staff1.PostBooking(bookingRequests)
		if err != nil {
			t.Fatalf("single booking approve error %v", err)
		}

		if len(bookings) != 1 {
			t.Fatalf("expected %d bookings, received %d", 1, len(bookings))
		}
	})

	t.Run("reserved slots are not usable in future quotes", func(tt *testing.T) {
		reservedSlot := &types.IGroupScheduleDateSlots{
			StartDate:             member1.Quotes[0].SlotDate,
			ScheduleBracketSlotId: member1.Quotes[0].ScheduleBracketSlotId,
		}
		_, err := member3.PostQuote(serviceTierId, reservedSlot, nil, nil)
		if err == nil {
			t.Fatalf("user was able to request a booked slot")
		}
	})

	t.Run("slots can be batch approved", func(tt *testing.T) {
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

		bookings, err := staff1.PostBooking(bookingRequests)
		if err != nil {
			t.Fatalf("batch booking approve error %v", err)
		}

		if len(bookings) != 2 {
			t.Fatalf("expected %d bookings, received %d", 2, len(bookings))
		}

		if !util.IsUUID(bookings[0].Id) {
			t.Fatalf("booking id 1 is not a uuid %s", bookings[0].Id)
		}

		if !util.IsUUID(bookings[1].Id) {
			t.Fatalf("booking id 2 is not a uuid %s", bookings[1].Id)
		}

		testutil.IntegrationTest.Bookings = append(testutil.IntegrationTest.Bookings, bookings...)
	})

	t.Run("master schedules can be disabled, preserving all records", func(tt *testing.T) {

	})

	t.Run("master schedules can be deleted, destroying all records", func(tt *testing.T) {

	})
}
