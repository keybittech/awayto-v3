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
	staff2 := integrationTest.TestUsers[2]
	member1 := integrationTest.TestUsers[4]
	member2 := integrationTest.TestUsers[5]
	member3 := integrationTest.TestUsers[6]

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

	firstSlot := integrationTest.DateSlots[0]
	secondSlot := integrationTest.DateSlots[1]

	var serviceTierId string
	for _, tier := range integrationTest.GroupService.Service.Tiers {
		serviceTierId = tier.Id
		break
	}

	t.Run("APP_GROUP_BOOKINGS is required to request quote", func(t *testing.T) {
		_, err := postQuote(staff2.TestToken, serviceTierId, firstSlot, nil, nil)
		if err == nil {
			t.Error("user created quote without APP_GROUP_BOOKINGS permissions")
		}

		integrationTest.Quote, err = postQuote(admin.TestToken, serviceTierId, firstSlot, nil, nil)
		if err != nil {
			t.Errorf("user can request quote error: %v", err)
		} else {
			t.Logf("quote created with id %s", integrationTest.Quote.Id)
		}
	})

	t.Run("multiple users can request the same slot", func(t *testing.T) {
		for _, member := range []*TestUser{member1, member2, member3} {
			var err error
			member.Quote, err = postQuote(member.TestToken, serviceTierId, secondSlot, nil, nil)
			if err != nil {
				t.Errorf("multiple users %d can request quote error: %v", member.TestUserId, err)
			}

			if !util.IsUUID(member.Quote.Id) {
				t.Errorf("user #%d quote id invalid %s", member.TestUserId, member.Quote.Id)
			}

			_, err = time.Parse("2006-01-02", member.Quote.SlotDate)
			if err != nil {
				t.Errorf("user %d quote slot date invalid %s", member.TestUserId, member.Quote.SlotDate)
			}

			if !util.IsUUID(member.Quote.ScheduleBracketSlotId) {
				t.Errorf("user %d quote sbs id invalid %s", member.TestUserId, member.Quote.SlotDate)
			}
		}
	})

	failCheck(t)
}
