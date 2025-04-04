package handlers

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"

	"github.com/golang/mock/gomock"
)

func TestBooking(t *testing.T) {
	bookingId := "booking-id"
	quoteId := "quote-id"
	slotDate := "slot-date"
	sbsId := "sbs-id"

	tests := []HandlersTestCase{
		{
			name: "PostBooking success",
			setupMocks: func(hts *HandlersTestSetup) {
				// Configure userSession
				// mockSession := &types.UserSession{UserSub: "user-sub"}
				// hts.MockRedis.EXPECT().ReqSession(gomock.Any()).Return(mockSession)

				// Do DB Tx
				// hts.MockDatabase.EXPECT().Client().Return(hts.MockDatabaseClient)
				// hts.MockDatabaseClient.EXPECT().Begin().Return(hts.MockDatabaseTx, nil)
				// hts.MockDatabaseTx.EXPECT().Rollback().Return(nil).AnyTimes()
				hts.MockDatabaseTx.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(hts.MockDatabaseRow)
				hts.MockDatabaseRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...interface{}) error {
					if len(dest) > 0 {
						*dest[0].(*string) = bookingId
					}
					return nil
				}).Times(1)

				// Delete Redis item
				hts.MockRedis.EXPECT().Client().Return(hts.MockRedisClient)
				hts.MockRedisClient.EXPECT().Del(gomock.Any(), gomock.Any())

				// Commit TX
				// hts.MockDatabaseTx.EXPECT().Commit().Return(nil)
			},
			handlerFunc: func(h *Handlers, w http.ResponseWriter, r *http.Request, session *types.UserSession, tx *interfaces.MockIDatabaseTx) (interface{}, error) {
				data := &types.PostBookingRequest{
					Bookings: []*types.IBooking{
						{
							Quote: &types.IQuote{
								Id:                    quoteId,
								SlotDate:              slotDate,
								ScheduleBracketSlotId: sbsId,
							},
						},
					},
				}

				return h.PostBooking(w, r, data, session, tx)
			},
			expectedRes: &types.PostBookingResponse{
				Bookings: []*types.IBooking{
					{
						Id: bookingId,
					},
				},
			},
		},
	}

	RunHandlerTests(t, tests)
}
