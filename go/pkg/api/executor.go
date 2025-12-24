package api

import (
	"context"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/handlers"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

type RequestExecutor func(ctx context.Context, w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession, dbc *clients.DatabaseClient) (handlers.ReqInfo, func(error) error)

func TxExecutor(ctx context.Context, w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession, dbc *clients.DatabaseClient) (handlers.ReqInfo, func(error) error) {
	poolTx, err := dbc.OpenPoolSessionTx(ctx, session)
	if err != nil {
		panic(util.ErrCheck(err))
	}

	reqInfo := handlers.ReqInfo{
		Ctx:     ctx,
		W:       w,
		Req:     req,
		Session: session,
		Tx:      poolTx,
	}

	return reqInfo, func(handlerErr error) error {
		defer poolTx.Rollback(ctx)

		// if an err occurred in handler, immediately return it + rollback
		// don't ErrCheck as the handlers should be doing so
		if handlerErr != nil {
			return handlerErr
		}

		err = dbc.ClosePoolSessionTx(ctx, poolTx)
		if err != nil {
			return util.ErrCheck(err)
		}

		return nil
	}
}

func BatchExecutor(ctx context.Context, w http.ResponseWriter, req *http.Request, session *types.ConcurrentUserSession, dbc *clients.DatabaseClient) (handlers.ReqInfo, func(error) error) {
	batch := util.NewBatchable(dbc.Pool, session.GetUserSub(), session.GetGroupId(), session.GetRoleBits())
	reqInfo := handlers.ReqInfo{
		Ctx:     ctx,
		W:       w,
		Req:     req,
		Session: session,
		Batch:   batch,
	}

	return reqInfo, func(_ error) error { return nil } //noop
}
