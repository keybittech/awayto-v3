package main

import (
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func testIntegrationBookings(t *testing.T) {
	admin := integrationTest.TestUsers[0]
	staff1 := integrationTest.TestUsers[1]
	member1 := integrationTest.TestUsers[4]
	member2 := integrationTest.TestUsers[5]
	member3 := integrationTest.TestUsers[6]

	t.Run("APP_GROUP_SCHEDULES is required by approver", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    integrationTest.Quote.Id,
				SlotDate:              integrationTest.Quote.SlotDate,
				ScheduleBracketSlotId: integrationTest.Quote.ScheduleBracketSlotId,
			},
		}

		_, err := postBooking(member1.TestToken, bookingRequests)
		if err == nil || !strings.Contains(err.Error(), "Forbidden") {
			if err != nil {
				t.Errorf("slot approval was not forbidden without APP_GROUP_SCHEDULES permission, error %v", err)
			} else {
				t.Error("booking post was successful without APP_GROUP_SCHEDULES")
			}
		}
	})

	t.Run("slot must be owned by approver", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    integrationTest.Quote.Id,
				SlotDate:              integrationTest.Quote.SlotDate,
				ScheduleBracketSlotId: integrationTest.Quote.ScheduleBracketSlotId,
			},
		}

		_, err := postBooking(admin.TestToken, bookingRequests)
		if err == nil {
			t.Error("slot was approved by non-owner approver")
		}
	})

	t.Run("different slots cannot be batch approved", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 2)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    integrationTest.Quote.Id,
				SlotDate:              integrationTest.Quote.SlotDate,
				ScheduleBracketSlotId: integrationTest.Quote.ScheduleBracketSlotId,
			},
		}
		bookingRequests[1] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quote.Id,
				SlotDate:              member1.Quote.SlotDate,
				ScheduleBracketSlotId: member1.Quote.ScheduleBracketSlotId,
			},
		}

		_, err := postBooking(staff1.TestToken, bookingRequests)
		if err == nil {
			t.Error("different booking approve was allowed")
		}
	})

	t.Run("staff can approve a quote, creating a booking", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 1)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    integrationTest.Quote.Id,
				SlotDate:              integrationTest.Quote.SlotDate,
				ScheduleBracketSlotId: integrationTest.Quote.ScheduleBracketSlotId,
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

	t.Run("slots can be batch approved, while rejecting some", func(t *testing.T) {
		bookingRequests := make([]*types.IBooking, 3)
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member1.Quote.Id,
				SlotDate:              member1.Quote.SlotDate,
				ScheduleBracketSlotId: member1.Quote.ScheduleBracketSlotId,
			},
		}
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member2.Quote.Id,
				SlotDate:              member2.Quote.SlotDate,
				ScheduleBracketSlotId: member2.Quote.ScheduleBracketSlotId,
			},
		}
		bookingRequests[0] = &types.IBooking{
			Quote: &types.IQuote{
				Id:                    member3.Quote.Id,
				SlotDate:              member3.Quote.SlotDate,
				ScheduleBracketSlotId: member3.Quote.ScheduleBracketSlotId,
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

	t.Run("slots cannot be requested which have already been approved", func(t *testing.T) {
		firstSlot := integrationTest.DateSlots[0]

		var serviceTierId string
		for _, tier := range integrationTest.GroupService.Service.Tiers {
			serviceTierId = tier.Id
			break
		}

		_, err := postQuote(admin.TestToken, serviceTierId, firstSlot, nil, nil)
		if err == nil {
			t.Errorf("user was able to request a booked slot")
		}
	})
}
