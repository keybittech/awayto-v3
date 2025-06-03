package handlers

import (
	"context"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/proto"
)

type ReqInfo struct {
	Ctx     context.Context
	W       http.ResponseWriter
	Req     *http.Request
	Session *types.ConcurrentUserSession
	Tx      *clients.PoolTx
	Batch   *util.Batchable
}

type TypedProtoHandler[ReqMsg, ResMsg proto.Message] func(info ReqInfo, message ReqMsg) (ResMsg, error)

type ProtoHandler func(info ReqInfo, message proto.Message) (proto.Message, error)

type Handlers struct {
	Functions map[string]ProtoHandler
	Options   map[string]*util.HandlerOptions
	LLM       *clients.LLM
	Database  *clients.Database
	Redis     *clients.Redis
	Keycloak  *clients.Keycloak
	Socket    *clients.Socket
	Cache     *util.Cache
}

func NewHandlers() *Handlers {
	h := &Handlers{
		Functions: make(map[string]ProtoHandler),
		LLM:       clients.InitLLM(),
		Database:  clients.InitDatabase(),
		Redis:     clients.InitRedis(),
		Keycloak:  clients.InitKeycloak(),
		Socket:    clients.InitSocket(),
		Cache:     util.NewCache(),
		Options:   util.GenerateOptions(),
	}
	registerHandlers(h)
	return h
}
