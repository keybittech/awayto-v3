package handlers

import (
	"errors"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/proto"
)

type Handlers struct {
	Functions map[string]HandlerWrapper
	Ai        *clients.Ai
	Database  *clients.Database
	Redis     *clients.Redis
	Keycloak  *clients.Keycloak
	Socket    *clients.Socket
}

func NewHandlers() *Handlers {
	h := &Handlers{
		Ai:       clients.InitAi(),
		Database: clients.InitDatabase(),
		Redis:    clients.InitRedis(),
		Keycloak: clients.InitKeycloak(),
		Socket:   clients.InitSocket(),
	}

	h.registerHandlers()

	return h
}

type TypedProtoHandler[
	Req proto.Message,
	Res proto.Message,
] func(w http.ResponseWriter, req *http.Request, message Req, session *types.UserSession, tx *clients.PoolTx) (Res, error)

type HandlerWrapper func(w http.ResponseWriter, req *http.Request, message proto.Message, session *types.UserSession, tx *clients.PoolTx) (proto.Message, error)

func Register[
	Req, Res proto.Message,
](handler TypedProtoHandler[Req, Res]) HandlerWrapper {
	return func(w http.ResponseWriter, req *http.Request, message proto.Message, session *types.UserSession, tx *clients.PoolTx) (proto.Message, error) {
		msg, ok := message.(Req)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return handler(w, req, msg, session, tx)
	}
}
