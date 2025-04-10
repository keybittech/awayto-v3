package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
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

func TestHandlers_PostBooking(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PostBookingRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostBookingResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostBooking(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostBooking(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostBooking(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchBooking(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PatchBookingRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchBookingResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchBooking(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchBooking(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchBooking(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetBookings(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetBookingsRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetBookingsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetBookings(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetBookings(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetBookings(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetBookingById(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetBookingByIdRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetBookingByIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetBookingById(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetBookingById(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetBookingById(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetBookingFiles(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.GetBookingFilesRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetBookingFilesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetBookingFiles(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetBookingFiles(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetBookingFiles(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchBookingRating(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.PatchBookingRatingRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchBookingRatingResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchBookingRating(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchBookingRating(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchBookingRating(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteBooking(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.DeleteBookingRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteBookingResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteBooking(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteBooking(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteBooking(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}

func TestHandlers_DisableBooking(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		req     *http.Request
		data    *types.DisableBookingRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DisableBookingResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DisableBooking(tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DisableBooking(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DisableBooking(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.req, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}
