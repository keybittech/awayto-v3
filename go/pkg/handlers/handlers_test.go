package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type HandlersTestSetup struct {
	MockCtrl           *gomock.Controller
	MockAi             *mocks.MockIAi
	MockDatabase       *mocks.MockIDatabase
	MockDatabaseClient *mocks.MockIDatabaseClient
	MockDatabaseTx     *mocks.MockIDatabaseTx
	MockDatabaseRows   *mocks.MockIRows
	MockDatabaseRow    *mocks.MockIRow
	MockRedis          *mocks.MockIRedis
	MockRedisClient    *mocks.MockIRedisClient
	MockKeycloak       *mocks.MockIKeycloak
	MockSocket         *mocks.MockISocket
	Handlers           *Handlers
	UserSession        *clients.UserSession
}

func SetupHandlersTest(t *testing.T) *HandlersTestSetup {
	ctrl := gomock.NewController(t)

	mockAi := mocks.NewMockIAi(ctrl)

	mockDatabase := mocks.NewMockIDatabase(ctrl)
	mockDatabaseClient := mocks.NewMockIDatabaseClient(ctrl)
	mockDatabaseTx := mocks.NewMockIDatabaseTx(ctrl)

	mockDatabaseRows := mocks.NewMockIRows(ctrl)
	mockDatabaseRow := mocks.NewMockIRow(ctrl)

	mockRedis := mocks.NewMockIRedis(ctrl)
	mockRedisClient := mocks.NewMockIRedisClient(ctrl)

	mockKeycloak := mocks.NewMockIKeycloak(ctrl)

	mockSock := mocks.NewMockISocket(ctrl)

	handlers := &Handlers{
		Ai:       mockAi,
		Database: mockDatabase,
		Redis:    mockRedis,
		Keycloak: mockKeycloak,
		Socket:   mockSock,
	}

	session := &clients.UserSession{
		UserSub:   "test-user-sub",
		UserEmail: "test@email.com",
		GroupId:   "test-group-id",
		GroupSub:  "test-group-sub",
	}

	return &HandlersTestSetup{
		MockCtrl:           ctrl,
		MockAi:             mockAi,
		MockDatabase:       mockDatabase,
		MockDatabaseClient: mockDatabaseClient,
		MockDatabaseTx:     mockDatabaseTx,
		MockDatabaseRows:   mockDatabaseRows,
		MockDatabaseRow:    mockDatabaseRow,
		MockRedis:          mockRedis,
		MockRedisClient:    mockRedisClient,
		MockKeycloak:       mockKeycloak,
		MockSocket:         mockSock,
		Handlers:           handlers,
		UserSession:        session,
	}
}

func (hts *HandlersTestSetup) TearDown() {
	hts.MockCtrl.Finish()
}

type HandlersTestCase struct {
	name        string
	setupMocks  func(*HandlersTestSetup)
	handlerFunc func(h *Handlers, w http.ResponseWriter, r *http.Request, session *clients.UserSession, tx *mocks.MockIDatabaseTx) (interface{}, error)
	expectedRes interface{}
	expectedErr error
}

func RunHandlerTests(t *testing.T, tests []HandlersTestCase) {

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			hts := SetupHandlersTest(t)
			defer hts.TearDown()

			tt.setupMocks(hts)

			rr := httptest.NewRecorder()

			req := httptest.NewRequest("GET", "/testing", nil)

			res, err := tt.handlerFunc(hts.Handlers, rr, req, hts.UserSession, hts.MockDatabaseTx)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedRes, res)
		})

	}

}
