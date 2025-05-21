package handlers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var integrationTest = &types.IntegrationTest{}

func TestMain(m *testing.M) {
	util.ParseEnv()

	jsonBytes, err := os.ReadFile(filepath.Join(util.E_PROJECT_DIR, "go", "integrations", "integration_results.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = protojson.Unmarshal(jsonBytes, integrationTest)
	if err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func setupTestEnv(useTx bool) (*Handlers, ReqInfo, func(), error) {
	h := NewHandlers()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)

	session := types.NewConcurrentUserSession(integrationTest.TestUsers[0].UserSession)
	ctx := req.Context()

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
			if got.Ai == nil {
				t.Error("Ai client was not initialized")
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
