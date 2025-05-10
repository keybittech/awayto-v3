package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/proto"
)

type Handlers struct {
	Functions map[string]ProtoHandler
	Ai        *clients.Ai
	Database  *clients.Database
	Redis     *clients.Redis
	Keycloak  *clients.Keycloak
	Socket    *clients.Socket
	Cache     *HandlerCache
}

func NewHandlers() *Handlers {
	h := &Handlers{
		Functions: make(map[string]ProtoHandler),
		Ai:        clients.InitAi(),
		Database:  clients.InitDatabase(),
		Redis:     clients.InitRedis(),
		Keycloak:  clients.InitKeycloak(),
		Socket:    clients.InitSocket(),
		Cache:     &HandlerCache{},
	}

	h.registerHandlers()

	return h
}

type ReqInfo struct {
	Ctx     context.Context
	W       http.ResponseWriter
	Req     *http.Request
	Session *types.UserSession
	Tx      *clients.PoolTx
}

type TypedProtoHandler[ReqMsg, ResMsg proto.Message] func(info ReqInfo, message ReqMsg) (ResMsg, error)

type ProtoHandler func(info ReqInfo, messages proto.Message) (proto.Message, error)

func Register[ReqMsg, ResMsg proto.Message](handler TypedProtoHandler[ReqMsg, ResMsg]) ProtoHandler {
	return func(info ReqInfo, message proto.Message) (proto.Message, error) {
		msg, ok := message.(ReqMsg)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return handler(info, msg)
	}
}
