package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"buf.build/go/protovalidate"
	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	util.ParseEnv()

	cmd, err := testutil.StartTestServer()
	if err != nil {
		panic(err)
	}
	if cmd != nil {
		defer func() {
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("Failed to close server: %v", util.ErrCheck(err))
			}
		}()
	}

	testutil.LoadIntegrations()

	m.Run()
}

func setupTestEnv(useTx bool) (*Handlers, ReqInfo, func(), error) {
	h := NewHandlers()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)

	testUser := testutil.IntegrationTest.TestUsers[0]

	_, err := testUser.Login()
	if err != nil {
		return nil, ReqInfo{}, nil, util.ErrCheck(fmt.Errorf("handlers test could not login as test user, %v", err))
	}

	err = testUser.GetVaultKey()
	if err != nil {
		return nil, ReqInfo{}, nil, util.ErrCheck(fmt.Errorf("handlers test could not get vault key: %v", err))
	}

	ctx := req.Context()

	workerDbSession := &clients.DbSession{
		Pool: h.Database.Client().Pool,
		ConcurrentUserSession: types.NewConcurrentUserSession(&types.UserSession{
			UserSub: "worker",
		}),
	}

	row, done, err := workerDbSession.SessionBatchQueryRow(ctx, `
		SELECT sub 
		FROM dbtable_schema.users 
		WHERE email = $1
	`, testUser.GetProfile().GetEmail())
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	var userSub string
	err = row.Scan(&userSub)
	if err != nil {
		done()
		log.Fatal(util.ErrCheck(err))
	}
	done()

	session, err := testUser.GetUserSession(h.Database.DatabaseClient.Pool)
	if err != nil {
		return nil, ReqInfo{}, nil, util.ErrCheck(err)
	}

	session.SetUserSub(userSub)

	reqInfo := ReqInfo{
		Ctx:     ctx,
		W:       w,
		Req:     req,
		Session: session,
	}

	if useTx {
		poolTx, err := h.Database.DatabaseClient.OpenPoolSessionTx(ctx, session)
		if err != nil {
			return nil, ReqInfo{}, nil, util.ErrCheck(err)
		}
		reqInfo.Tx = poolTx
		return h, reqInfo, func() { poolTx.Rollback(ctx) }, nil
	} else {
		reqInfo.Batch = util.NewBatchable(h.Database.DatabaseClient.Pool, session.GetUserSub(), session.GetGroupId(), session.GetRoleBits())
		return h, reqInfo, func() {}, nil
	}
}

func validateData(t *testing.T, data proto.Message, name string, expectErr bool) bool {
	err := protovalidate.Validate(data)

	if (err != nil) != expectErr {
		t.Errorf("%s: Validation error got: %v,  expected: %v", name, (err != nil), expectErr)
		if err != nil {
			t.Logf("Validation errors: %v", err)
		}
		return false
	}

	if expectErr {
		return true
	}

	return false
}

func TestNewHandlers(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Initializes handlers"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHandlers()
			if got == nil {
				t.Error("NewHandlers() returned nil")
			}
			if got.Functions == nil {
				t.Error("Functions map was not initialized")
			}
			if got.LLM == nil {
				t.Error("LLM client was not initialized")
			}
			if got.Database == nil {
				t.Error("Database client was not initialized")
			}
			if got.Redis == nil {
				t.Error("Redis client was not initialized")
			}
			if got.Keycloak == nil {
				t.Error("Keycloak client was not initialized")
			}
			if got.Socket == nil {
				t.Error("Socket client was not initialized")
			}
		})
	}
}

func FuzzPostSchedule(f *testing.F) {
	handlers := NewHandlers()
	testUser := testutil.IntegrationTest.TestUsers[0]
	session, err := testUser.GetUserSession(handlers.Database.DatabaseClient.Pool)
	if err != nil {
		f.Fatalf("failed to get session during fuzz, err: %v", err)
	}
	// Create fake context so we can use our regular http handlers
	fakeReq, err := http.NewRequest("GET", "handlers://fuzz", nil)
	if err != nil {
		f.Fatalf("failed to create fuzz request %v", err)
	}

	groupScheuleId := testutil.IntegrationTest.GetMasterSchedule().GetId()
	scheduleTimeUnitId := testutil.IntegrationTest.GetMasterSchedule().GetScheduleTimeUnitId()
	bracketTimeUnitId := testutil.IntegrationTest.GetMasterSchedule().GetBracketTimeUnitId()
	slotTimeUnitId := testutil.IntegrationTest.GetMasterSchedule().GetSlotTimeUnitId()

	f.Add("my schedule", "03-03-2025", "04-02-2025", int32(30))

	f.Fuzz(func(t *testing.T, name, startDate, endDate string, slotDuration int32) {

		pb := &types.PostScheduleRequest{
			Name:               name,
			StartDate:          startDate,
			EndDate:            endDate,
			ScheduleTimeUnitId: scheduleTimeUnitId,
			BracketTimeUnitId:  bracketTimeUnitId,
			SlotTimeUnitId:     slotTimeUnitId,
			GroupScheduleId:    groupScheuleId,
			SlotDuration:       slotDuration,
			// ApprovalMode
		}

		err = protovalidate.Validate(pb)
		if err != nil {
			t.Skipf("validator caught issue %v", err)
		}

		ctx, cancel := context.WithTimeout(fakeReq.Context(), 5*time.Second)
		defer cancel()

		reqInfo := ReqInfo{
			Ctx:     ctx,
			W:       nil,
			Req:     fakeReq,
			Session: session,
		}

		handlers.PostSchedule(reqInfo, pb)
	})
}
