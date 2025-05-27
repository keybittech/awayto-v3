package util

import (
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

type Cache struct {
	TempAuthSessions     *types.ConcurrentTempAuthSessionCache
	UserSessions         *types.ConcurrentUserSessionCache
	Groups               *types.ConcurrentCachedGroupCache
	SubGroups            *types.ConcurrentCachedSubGroupCache
	GroupSessionVersions *types.InternalCache[string, int64]
}

func NewCache() *Cache {
	return &Cache{
		TempAuthSessions:     types.NewConcurrentTempAuthSessionCache(),
		UserSessions:         types.NewConcurrentUserSessionCache(),
		Groups:               types.NewConcurrentCachedGroupCache(),
		SubGroups:            types.NewConcurrentCachedSubGroupCache(),
		GroupSessionVersions: types.NewInternalCache[string, int64](),
	}
}
