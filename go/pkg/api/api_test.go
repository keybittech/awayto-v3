package api

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
)

func SetupHandlersTest(t *testing.T) *interfaces.DefaultTestSetup {
	ctrl := gomock.NewController(t)

	mockAi := interfaces.NewMockIAi(ctrl)

	mockDatabase := interfaces.NewMockIDatabase(ctrl)
	mockDatabaseClient := interfaces.NewMockIDatabaseClient(ctrl)
	mockDatabaseTx := interfaces.NewMockIDatabaseTx(ctrl)

	mockDatabaseRows := interfaces.NewMockIRows(ctrl)
	mockDatabaseRow := interfaces.NewMockIRow(ctrl)

	mockRedis := interfaces.NewMockIRedis(ctrl)
	mockRedisClient := interfaces.NewMockIRedisClient(ctrl)

	mockKeycloak := interfaces.NewMockIKeycloak(ctrl)

	mockSock := interfaces.NewMockISocket(ctrl)

	return &interfaces.DefaultTestSetup{
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
	}
}

func RunHandlerTests(t *testing.T, tests []interfaces.DefaultTestCase) {
	for _, tt := range tests {

		t.Run(tt.Name, func(t *testing.T) {
			hts := SetupHandlersTest(t)
			defer hts.TearDown()

			tt.SetupMocks(hts)

			// res, err := tt.()

			// assert.Equal(t, tt.ExpectedErr, err)
			// assert.Equal(t, tt.ExpectedRes, res)
		})
	}
}
