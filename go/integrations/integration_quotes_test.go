package main_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testIntegrationQuotes(t *testing.T) {
	admin := testutil.IntegrationTest.TestUsers[0]
	staff1 := testutil.IntegrationTest.TestUsers[1]
	staff2 := testutil.IntegrationTest.TestUsers[2]
	member1 := testutil.IntegrationTest.TestUsers[4]
	member1.Quotes = make([]*types.IQuote, 10)
	member2 := testutil.IntegrationTest.TestUsers[5]
	member2.Quotes = make([]*types.IQuote, 10)
	member3 := testutil.IntegrationTest.TestUsers[6]
	member3.Quotes = make([]*types.IQuote, 10)

	t.Run("date slots are retrieved prior to getting a quote", func(tt *testing.T) {
		var err error
		testutil.IntegrationTest.DateSlots, err = admin.GetDateSlots(testutil.IntegrationTest.MasterSchedule.Id)
		if err != nil {
			t.Fatalf("date slot retrieval error: %v", err)
		}
	})

	var serviceTierId string
	for _, tier := range testutil.IntegrationTest.GroupService.Service.Tiers {
		serviceTierId = tier.Id
		break
	}

	t.Run("APP_GROUP_BOOKINGS is required to request quote", func(tt *testing.T) {
		firstSlot := testutil.IntegrationTest.DateSlots[0]
		_, err := staff2.PostQuote(serviceTierId, firstSlot, nil, nil)
		if err == nil {
			t.Fatalf("user created quote without APP_GROUP_BOOKINGS permissions")
		}

		quote, err := member1.PostQuote(serviceTierId, firstSlot, nil, nil)
		if err != nil {
			t.Fatalf("user can request quote error: %v", err)
		} else {
			t.Logf("quote created with id %s", quote.Id)
		}

		member1.Quotes[0] = quote
	})

	t.Run("APP_GROUP_SCHEDULES is required to disable a quote", func(tt *testing.T) {
		disableQuoteResponse := &types.DisableQuoteResponse{}
		err := staff1.DoHandler(http.MethodPatch, "/api/v1/quotes/disable/"+member1.Quotes[0].Id, nil, nil, disableQuoteResponse)
		if err != nil {
			t.Fatalf("error disabling quote request: %v", err)
		}

		if !disableQuoteResponse.Success {
			t.Fatalf("disabling the quote was not successful")
		}
		member1.Quotes[0] = nil
	})

	t.Run("multiple users can request the same slot", func(tt *testing.T) {
		if len(testutil.IntegrationTest.DateSlots) < 2 {
			t.Fatalf("date slots wrong len, expected 2, got: %d", len(testutil.IntegrationTest.DateSlots))
		}
		for i := range 2 {
			bracketSlot := testutil.IntegrationTest.DateSlots[i]

			for _, member := range []*testutil.TestUsersStruct{member1, member2, member3} {
				quote, err := member.PostQuote(serviceTierId, bracketSlot, nil, nil)
				if err != nil {
					t.Fatalf("multiple users %s can request quote error: %v", member.TestUserId, err)
				}

				if !util.IsUUID(quote.Id) {
					t.Fatalf("user #%s quote id invalid %s", member.TestUserId, quote.Id)
				}

				// Staff can see the quote info
				if quote, err = staff1.GetQuoteById(quote.Id); err != nil {
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
		dateSlots, err := admin.GetDateSlots(testutil.IntegrationTest.MasterSchedules[0].Id)
		if err != nil {
			t.Fatalf("date slot retrieval error: %v", err)
		}
		if len(dateSlots) < 2 {
			t.Fatalf("secondary date slots wrong len, wanted 2, got %d", len(dateSlots))
		}
		for i := range 2 {
			bracketSlot := dateSlots[i]

			for _, member := range []*testutil.TestUsersStruct{member1, member2, member3} {
				quote, err := member.PostQuote(serviceTierId, bracketSlot, nil, nil)
				if err != nil {
					t.Fatalf("secondary request quote error: %v", err)
				}

				testutil.IntegrationTest.Quotes = append(testutil.IntegrationTest.Quotes, quote)
			}
		}
	})
}
