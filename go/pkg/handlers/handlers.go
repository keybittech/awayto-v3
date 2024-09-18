package handlers

import (
	"av3api/pkg/clients"
)

type Handlers struct {
	Ai       clients.IAi
	Database clients.IDatabase
	Redis    clients.IRedis
	Keycloak clients.IKeycloak
	Socket   clients.ISocket
}
