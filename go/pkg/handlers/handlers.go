package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/clients"
)

type Handlers struct {
	Ai       clients.IAi
	Database clients.IDatabase
	Redis    clients.IRedis
	Keycloak clients.IKeycloak
	Socket   clients.ISocket
}
