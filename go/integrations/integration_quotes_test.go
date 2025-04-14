package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testIntegrationQuotes(t *testing.T) {
	admin := integrationTest.TestUsers[0]
	staff1 := integrationTest.TestUsers[1]
	staff2 := integrationTest.TestUsers[2]
	member1 := integrationTest.TestUsers[4]
	member1.Quotes = make([]*types.IQuote, 10)
	member2 := integrationTest.TestUsers[5]
	member2.Quotes = make([]*types.IQuote, 10)
	member3 := integrationTest.TestUsers[6]
	member3.Quotes = make([]*types.IQuote, 10)

	t.Run("date slots are retrieved prior to getting a quote", func(t *testing.T) {
		now := time.Now()
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		dateSlotsUrl := "/api/v1/group/schedules/" + integrationTest.MasterSchedule.Id + "/date/" + firstOfMonth
		dateSlotsResponse := &types.GetGroupScheduleByDateResponse{}
		err := apiRequest(admin.TestToken, http.MethodGet, dateSlotsUrl, nil, nil, dateSlotsResponse)
		if err != nil {
			t.Errorf("error get group date slots request, error: %v", err)
		}

		if len(dateSlotsResponse.GroupScheduleDateSlots) == 0 {
			t.Errorf("no date slots available to schedule %v", dateSlotsResponse)
		}

		integrationTest.DateSlots = dateSlotsResponse.GroupScheduleDateSlots
	})

	var serviceTierId string
	for _, tier := range integrationTest.GroupService.Service.Tiers {
		serviceTierId = tier.Id
		break
	}

	t.Run("APP_GROUP_BOOKINGS is required to request quote", func(t *testing.T) {
		firstSlot := integrationTest.DateSlots[0]
		_, err := postQuote(staff2.TestToken, serviceTierId, firstSlot, nil, nil)
		if err == nil {
			t.Error("user created quote without APP_GROUP_BOOKINGS permissions")
		}

		quote, err := postQuote(member1.TestToken, serviceTierId, firstSlot, nil, nil)
		if err != nil {
			t.Errorf("user can request quote error: %v", err)
		} else {
			t.Logf("quote created with id %s", quote.Id)
		}

		member1.Quotes[0] = quote
	})

	t.Run("APP_GROUP_SCHEDULES is required to disable a quote", func(t *testing.T) {
		disableQuoteResponse := &types.DisableQuoteResponse{}
		err := apiRequest(staff1.TestToken, http.MethodPatch, "/api/v1/quotes/disable/"+member1.Quotes[0].Id, nil, nil, disableQuoteResponse)
		if err != nil {
			t.Errorf("error disabling quote request: %v", err)
		}

		if !disableQuoteResponse.Success {
			t.Error("disabling the quote was not successful")
		}
		member1.Quotes[0] = nil
	})

	t.Run("multiple users can request the same slot", func(t *testing.T) {
		for i := range 2 {
			bracketSlot := integrationTest.DateSlots[i]

			for _, member := range []*TestUser{member1, member2, member3} {
				quote, err := postQuote(member.TestToken, serviceTierId, bracketSlot, nil, nil)
				if err != nil {
					t.Errorf("multiple users %d can request quote error: %v", member.TestUserId, err)
				}

				if !util.IsUUID(quote.Id) {
					t.Errorf("user #%d quote id invalid %s", member.TestUserId, quote.Id)
				}

				// Staff can see the quote info
				if _, err := getQuoteById(staff1.TestToken, quote.Id); err != nil {
					t.Errorf("user #%d quote id %s bad get %v", member.TestUserId, quote.Id, err)
				}

				if _, err := time.Parse("2006-01-02", quote.SlotDate); err != nil {
					t.Errorf("user %d quote %d slot date invalid %s", member.TestUserId, i, quote.SlotDate)
				}

				if !util.IsUUID(quote.ScheduleBracketSlotId) {
					t.Errorf("user %d quote %d sbs id invalid %s", member.TestUserId, i, quote.SlotDate)
				}

				if quote.ScheduleBracketSlotId != bracketSlot.ScheduleBracketSlotId {
					t.Errorf("user %d quote %d sbs id %s doesn't match requested slot id %s", member.TestUserId, i, quote.SlotDate, bracketSlot.ScheduleBracketSlotId)
				}

				member.Quotes[i] = quote
			}
		}
	})
}
