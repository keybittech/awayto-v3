package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type Handlers struct {
	Functions map[string]ProtoHandler
	Options   map[string]*util.HandlerOptions
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

	h.Options = make(map[string]*util.HandlerOptions, len(h.Functions))

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}

		services := fd.Services().Get(0)

		for i := 0; i <= services.Methods().Len()-1; i++ {
			serviceMethod := services.Methods().Get(i)
			handlerOpts := util.ParseHandlerOptions(serviceMethod)

			protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
				if mt.Descriptor().FullName() == serviceMethod.Input().FullName() {
					handlerOpts.ServiceMethodType = mt
					return false
				}
				return true
			})

			h.Options[handlerOpts.ServiceMethodName] = handlerOpts
		}

		return true
	})

	util.ParseInvalidations(h.Options)

	return h
}

type ReqInfo struct {
	Ctx     context.Context
	W       http.ResponseWriter
	Req     *http.Request
	Session *types.UserSession
	Tx      *clients.PoolTx
	Batch   *util.Batchable
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
