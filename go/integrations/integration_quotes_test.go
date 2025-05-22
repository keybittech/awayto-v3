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

	t.Run("date slots are retrieved prior to getting a quote", func(tt *testing.T) {
		var err error
		integrationTest.DateSlots, err = getDateSlots(admin.TestToken, integrationTest.MasterSchedule.Id)
		if err != nil {
			t.Fatalf("date slot retrieval error: %v", err)
		}
	})

	var serviceTierId string
	for _, tier := range integrationTest.GroupService.Service.Tiers {
		serviceTierId = tier.Id
		break
	}

	t.Run("APP_GROUP_BOOKINGS is required to request quote", func(tt *testing.T) {
		firstSlot := integrationTest.DateSlots[0]
		_, err := postQuote(staff2.TestToken, serviceTierId, firstSlot, nil, nil)
		if err == nil {
			t.Fatalf("user created quote without APP_GROUP_BOOKINGS permissions")
		}

		quote, err := postQuote(member1.TestToken, serviceTierId, firstSlot, nil, nil)
		if err != nil {
			t.Fatalf("user can request quote error: %v", err)
		} else {
			t.Logf("quote created with id %s", quote.Id)
		}

		member1.Quotes[0] = quote
	})

	t.Run("APP_GROUP_SCHEDULES is required to disable a quote", func(tt *testing.T) {
		disableQuoteResponse := &types.DisableQuoteResponse{}
		err := apiRequest(staff1.TestToken, http.MethodPatch, "/api/v1/quotes/disable/"+member1.Quotes[0].Id, nil, nil, disableQuoteResponse)
		if err != nil {
			t.Fatalf("error disabling quote request: %v", err)
		}

		if !disableQuoteResponse.Success {
			t.Fatalf("disabling the quote was not successful")
		}
		member1.Quotes[0] = nil
	})

	t.Run("multiple users can request the same slot", func(tt *testing.T) {
		if len(integrationTest.DateSlots) < 2 {
			t.Fatalf("date slots wrong len, expected 2, got: %d", len(integrationTest.DateSlots))
		}
		for i := range 2 {
			bracketSlot := integrationTest.DateSlots[i]

			for _, member := range []*types.TestUser{member1, member2, member3} {
				quote, err := postQuote(member.TestToken, serviceTierId, bracketSlot, nil, nil)
				if err != nil {
					t.Fatalf("multiple users %s can request quote error: %v", member.TestUserId, err)
				}

				if !util.IsUUID(quote.Id) {
					t.Fatalf("user #%s quote id invalid %s", member.TestUserId, quote.Id)
				}

				// Staff can see the quote info
				if _, err := getQuoteById(staff1.TestToken, quote.Id); err != nil {
					t.Fatalf("user #%s quote id %s bad get %v", member.TestUserId, quote.Id, err)
				}

				if _, err := time.Parse("2006-01-02", quote.SlotDate); err != nil {
					t.Fatalf("user %s quote %d slot date invalid %s", member.TestUserId, i, quote.SlotDate)
				}

				if !util.IsUUID(quote.ScheduleBracketSlotId) {
					t.Fatalf("user %s quote %d sbs id invalid %s", member.TestUserId, i, quote.SlotDate)
				}

				if quote.ScheduleBracketSlotId != bracketSlot.ScheduleBracketSlotId {
					t.Fatalf("user %s quote %d sbs id %s doesn't match requested slot id %s", member.TestUserId, i, quote.SlotDate, bracketSlot.ScheduleBracketSlotId)
				}

				member.Quotes[i] = quote
			}
		}
	})

	t.Run("secondary quote creation", func(tt *testing.T) {
		dateSlots, err := getDateSlots(admin.TestToken, integrationTest.MasterSchedules[0].Id)
		if err != nil {
			t.Fatalf("date slot retrieval error: %v", err)
		}
		if len(dateSlots) < 2 {
			t.Fatalf("secondary date slots wrong len, wanted 2, got %d", len(dateSlots))
		}
		for i := range 2 {
			bracketSlot := dateSlots[i]

			for _, member := range []*types.TestUser{member1, member2, member3} {
				quote, err := postQuote(member.TestToken, serviceTierId, bracketSlot, nil, nil)
				if err != nil {
					t.Fatalf("secondary request quote error: %v", err)
				}

				integrationTest.Quotes = append(integrationTest.Quotes, quote)
			}
		}
	})
}
