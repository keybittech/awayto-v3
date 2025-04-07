package handlers

import (
	"net/http"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"

	"github.com/golang/mock/gomock"
)

type HandlersTestSetup struct {
	MockCtrl           *gomock.Controller
	MockAi             *interfaces.MockIAi
	MockDatabase       *interfaces.MockIDatabase
	MockDatabaseClient *interfaces.MockIDatabaseClient
	MockDatabaseTx     *interfaces.MockIDatabaseTx
	MockDatabaseRows   *interfaces.MockIRows
	MockDatabaseRow    *interfaces.MockIRow
	MockRedis          *interfaces.MockIRedis
	MockRedisClient    *interfaces.MockIRedisClient
	MockKeycloak       *interfaces.MockIKeycloak
	MockSocket         *interfaces.MockISocket
	Handlers           *Handlers
	UserSession        *types.UserSession
}

//	func SetupHandlersTest(t *testing.T) *HandlersTestSetup {
//		ctrl := gomock.NewController(t)
//
//		mockAi := interfaces.NewMockIAi(ctrl)
//
//		mockDatabase := interfaces.NewMockIDatabase(ctrl)
//		mockDatabaseClient := interfaces.NewMockIDatabaseClient(ctrl)
//		mockDatabaseTx := interfaces.NewMockIDatabaseTx(ctrl)
//
//		mockDatabaseRows := interfaces.NewMockIRows(ctrl)
//		mockDatabaseRow := interfaces.NewMockIRow(ctrl)
//
//		mockRedis := interfaces.NewMockIRedis(ctrl)
//		mockRedisClient := interfaces.NewMockIRedisClient(ctrl)
//
//		mockKeycloak := interfaces.NewMockIKeycloak(ctrl)
//
//		mockSock := interfaces.NewMockISocket(ctrl)
//
//		handlers := &Handlers{
//			Ai:       mockAi,
//			Database: mockDatabase,
//			Redis:    mockRedis,
//			Keycloak: mockKeycloak,
//			Socket:   mockSock,
//		}
//
//		session := &types.UserSession{
//			UserSub:   "test-user-sub",
//			UserEmail: "test@email.com",
//			GroupId:   "test-group-id",
//			GroupSub:  "test-group-sub",
//		}
//
//		return &HandlersTestSetup{
//			MockCtrl:           ctrl,
//			MockAi:             mockAi,
//			MockDatabase:       mockDatabase,
//			MockDatabaseClient: mockDatabaseClient,
//			MockDatabaseTx:     mockDatabaseTx,
//			MockDatabaseRows:   mockDatabaseRows,
//			MockDatabaseRow:    mockDatabaseRow,
//			MockRedis:          mockRedis,
//			MockRedisClient:    mockRedisClient,
//			MockKeycloak:       mockKeycloak,
//			MockSocket:         mockSock,
//			Handlers:           handlers,
//			UserSession:        session,
//		}
//	}
//
//	func (hts *HandlersTestSetup) TearDown() {
//		hts.MockCtrl.Finish()
//	}
type HandlersTestCase struct {
	name        string
	setupMocks  func(*HandlersTestSetup)
	handlerFunc func(h *Handlers, w http.ResponseWriter, r *http.Request, session *types.UserSession, tx *interfaces.MockIDatabaseTx) (interface{}, error)
	expectedRes interface{}
	expectedErr error
}

func RunHandlerTests(t *testing.T, tests []HandlersTestCase) {

	// for _, tt := range tests {
	//
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		hts := SetupHandlersTest(t)
	// 		defer hts.TearDown()
	//
	// 		tt.setupMocks(hts)
	//
	// 		rr := httptest.NewRecorder()
	//
	// 		req := httptest.NewRequest("GET", "/testing", nil)
	//
	// 		res, err := tt.handlerFunc(hts.Handlers, rr, req, hts.UserSession, hts.MockDatabaseTx)
	//
	// 		assert.Equal(t, tt.expectedErr, err)
	// 		assert.Equal(t, tt.expectedRes, res)
	// 	})
	//
	// }

}
