package handlers

import (
	"github.com/keybittech/awayto-v3/go/pkg/crypto"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func (h *Handlers) GetVaultKey(info ReqInfo, data *types.GetVaultKeyRequest) (*types.GetVaultKeyResponse, error) {
	return &types.GetVaultKeyResponse{Key: crypto.EncodedVaultKey}, nil
}
