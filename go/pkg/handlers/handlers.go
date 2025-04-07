package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/clients"
)

type Handlers struct {
	Ai       *clients.Ai
	Database *clients.Database
	Redis    *clients.Redis
	Keycloak *clients.Keycloak
	Socket   *clients.Socket
}
