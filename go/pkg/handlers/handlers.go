package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
)

type Handlers struct {
	Ai       interfaces.IAi
	Database interfaces.IDatabase
	Redis    interfaces.IRedis
	Keycloak interfaces.IKeycloak
	Socket   interfaces.ISocket
}
