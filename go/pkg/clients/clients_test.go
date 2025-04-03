package clients

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/stretchr/testify/assert"
)

func SetupClientsTest(t *testing.T) *interfaces.DefaultTestSetup {
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

type ClientsTestCase struct {
	name        string
	setupMocks  func(*interfaces.DefaultTestSetup)
	params      []interface{}
	handlerFunc func(...interface{}) (interface{}, interface{}, error)
	expectedRes interface{}
	expectedErr error
}

func RunClientsTests(t *testing.T, tests []ClientsTestCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hts := SetupClientsTest(t)
			defer hts.TearDown()

			tt.setupMocks(hts)

			res1, res2, err := tt.handlerFunc(tt.params...)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedRes, []interface{}{res1, res2})
		})
	}
}
