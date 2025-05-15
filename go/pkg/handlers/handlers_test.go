package handlers

import (
	"errors"
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

func init() {
	jsonBytes, err := os.ReadFile(filepath.Join(os.Getenv("PROJECT_DIR"), "go", "integrations", "integration_results.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = protojson.Unmarshal(jsonBytes, integrationTest)
	if err != nil {
		log.Fatal(err)
	}
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func setupTestEnv(useTx bool) (*Handlers, ReqInfo, func(), error) {
	h := NewHandlers()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)

	session := integrationTest.TestUsers[0].UserSession
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
		reqInfo.Batch = util.NewBatchable(h.Database.DatabaseClient.Pool, session.UserSub, session.GroupId, session.RoleBits)
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

// Mock implementations for testing
type mockProtoMessage struct {
	proto.Message
	data string
}

func (m *mockProtoMessage) Reset()         {}
func (m *mockProtoMessage) String() string { return m.data }
func (m *mockProtoMessage) ProtoMessage()  {}

func TestRegister(t *testing.T) {

	tests := []struct {
		name         string
		inputMessage proto.Message
		expectError  bool
		expectedData string
	}{
		{
			name:         "Valid message type",
			inputMessage: &mockProtoMessage{data: "test"},
			expectError:  false,
			expectedData: "response",
		},
		{
			name:         "Invalid message type",
			inputMessage: &mockProtoMessage{},
			expectError:  true,
			expectedData: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typedHandler := func(info ReqInfo, message *mockProtoMessage) (*mockProtoMessage, error) {
				if message.data == "" {
					return nil, errors.New("invalid request type")
				}
				return &mockProtoMessage{data: "response"}, nil
			}

			handler := Register(typedHandler)

			result, err := handler(ReqInfo{}, tt.inputMessage)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				resultMsg, ok := result.(*mockProtoMessage)
				if !ok {
					t.Errorf("Result is not of expected type")
				} else if resultMsg.data != tt.expectedData {
					t.Errorf("Got result data %s, expected %s", resultMsg.data, tt.expectedData)
				}
			}
		})
	}
}
