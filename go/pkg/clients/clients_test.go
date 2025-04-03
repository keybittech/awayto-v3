package clients

//
// type ClientsTestSetup struct {
// 	MockCtrl           *gomock.Controller
// 	MockAi             *mocks.MockIAi
// 	MockDatabase       *mocks.MockIDatabase
// 	MockDatabaseClient *mocks.MockIDatabaseClient
// 	MockDatabaseTx     *mocks.MockIDatabaseTx
// 	MockDatabaseRows   *mocks.MockIRows
// 	MockDatabaseRow    *mocks.MockIRow
// 	MockRedis          *mocks.MockIRedis
// 	MockRedisClient    *mocks.MockIRedisClient
// 	MockKeycloak       *mocks.MockIKeycloak
// 	MockSocket         *mocks.MockISocket
// }
//
// func SetupHandlersTest(t *testing.T) *ClientsTestSetup {
// 	ctrl := gomock.NewController(t)
//
// 	mockAi := mocks.NewMockIAi(ctrl)
//
// 	mockDatabase := mocks.NewMockIDatabase(ctrl)
// 	mockDatabaseClient := mocks.NewMockIDatabaseClient(ctrl)
// 	mockDatabaseTx := mocks.NewMockIDatabaseTx(ctrl)
//
// 	mockDatabaseRows := mocks.NewMockIRows(ctrl)
// 	mockDatabaseRow := mocks.NewMockIRow(ctrl)
//
// 	mockRedis := mocks.NewMockIRedis(ctrl)
// 	mockRedisClient := mocks.NewMockIRedisClient(ctrl)
//
// 	mockKeycloak := mocks.NewMockIKeycloak(ctrl)
//
// 	mockSock := mocks.NewMockISocket(ctrl)
//
// 	return &ClientsTestSetup{
// 		MockCtrl:           ctrl,
// 		MockAi:             mockAi,
// 		MockDatabase:       mockDatabase,
// 		MockDatabaseClient: mockDatabaseClient,
// 		MockDatabaseTx:     mockDatabaseTx,
// 		MockDatabaseRows:   mockDatabaseRows,
// 		MockDatabaseRow:    mockDatabaseRow,
// 		MockRedis:          mockRedis,
// 		MockRedisClient:    mockRedisClient,
// 		MockKeycloak:       mockKeycloak,
// 		MockSocket:         mockSock,
// 	}
// }
//
// func (hts *ClientsTestSetup) TearDown() {
// 	hts.MockCtrl.Finish()
// }
//
// type ClientsTestCase struct {
// 	name        string
// 	setupMocks  func(*ClientsTestSetup)
// 	handlerFunc func() (interface{}, error)
// 	expectedRes interface{}
// 	expectedErr error
// }
//
// func RunHandlerTests(t *testing.T, tests []ClientsTestCase) {
// 	for _, tt := range tests {
//
// 		t.Run(tt.name, func(t *testing.T) {
// 			hts := SetupHandlersTest(t)
// 			defer hts.TearDown()
//
// 			tt.setupMocks(hts)
//
// 			res, err := tt.handlerFunc()
//
// 			assert.Equal(t, tt.expectedErr, err)
// 			assert.Equal(t, tt.expectedRes, res)
// 		})
// 	}
// }
