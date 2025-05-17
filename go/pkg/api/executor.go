package api

import (
	"context"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type RequestExecutor func(ctx context.Context, dbc *clients.DatabaseClient, session *types.UserSession) (handlers.ReqInfo, func() error)

func TxExecutor(ctx context.Context, dbc *clients.DatabaseClient, session *types.UserSession) (handlers.ReqInfo, func() error) {
	poolTx, err := dbc.OpenPoolSessionTx(ctx, session)
	if err != nil {
		panic(util.ErrCheck(err))
	}

	reqInfo := handlers.ReqInfo{
		Tx: poolTx,
	}

	return reqInfo, func() error {
		defer poolTx.Rollback(ctx)

		err = dbc.ClosePoolSessionTx(ctx, poolTx)
		if err != nil {
			return util.ErrCheck(err)
		}

		return nil
	}
}

func BatchExecutor(ctx context.Context, dbc *clients.DatabaseClient, session *types.UserSession) (handlers.ReqInfo, func() error) {
	reqInfo := handlers.ReqInfo{
		Batch: util.NewBatchable(dbc.Pool, session.UserSub, session.GroupId, session.RoleBits),
	}

	return reqInfo, func() error { return nil } //noop
}
