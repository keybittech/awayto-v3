package api

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

const (
	maxCacheBuffer        = 1 << 12
	cacheKeySuffixModTime = ":mod"
)

var (
	apiPathLen     = len(os.Getenv("API_PATH"))
	duration180, _ = time.ParseDuration("180s")
)

type CacheItem[V any] struct {
	Value      V            // must be thread-safe
	LastAccess atomic.Value // time.Time
}

func newCacheItem[V any](value V) *CacheItem[V] {
	item := &CacheItem[V]{
		Value: value,
	}
	item.LastAccess.Store(time.Now())
	return item
}

type InternalCache[K comparable, V any] struct {
	internal sync.Map
}

func NewInternalCache[K comparable, V any]() InternalCache[K, V] {
	return InternalCache[K, V]{}
}

func (c *InternalCache[K, V]) Load(key K) (V, bool) {
	var zero V
	value, ok := c.internal.Load(key)
	if !ok {
		return zero, false
	}

	item := value.(*CacheItem[V])

	item.LastAccess.Store(time.Now())

	return item.Value, true
}

func (c *InternalCache[K, V]) Store(key K, value V) {
	c.internal.Store(key, newCacheItem(value))
}

func (c *InternalCache[K, V]) CleanupOlderThan(olderThan time.Time) {
	c.internal.Range(func(k, v interface{}) bool {
		item := v.(*CacheItem[V])
		lastAccess := item.LastAccess.Load().(time.Time)

		if olderThan.After(lastAccess) {
			c.internal.Delete(k)
		}

		return true
	})
}

type SessionTokensCache struct {
	InternalCache[string, *types.ConcurrentUserSession]
}

func NewSessionTokensCache() *SessionTokensCache {
	return &SessionTokensCache{
		InternalCache: NewInternalCache[string, *types.ConcurrentUserSession](),
	}
}

type GroupsCache struct {
	InternalCache[string, *types.ConcurrentCachedGroup]
}

func NewGroupsCache() *GroupsCache {
	return &GroupsCache{
		InternalCache: NewInternalCache[string, *types.ConcurrentCachedGroup](),
	}
}

type SubGroupsCache struct {
	InternalCache[string, *types.ConcurrentCachedSubGroup]
}

func NewSubGroupsCache() *SubGroupsCache {
	return &SubGroupsCache{
		InternalCache: NewInternalCache[string, *types.ConcurrentCachedSubGroup](),
	}
}

type GroupSessionVersionsCache struct {
	InternalCache[string, int64]
}

func NewGroupSessionVersionsCache() *GroupSessionVersionsCache {
	return &GroupSessionVersionsCache{
		InternalCache: NewInternalCache[string, int64](),
	}
}

func (c *GroupSessionVersionsCache) Load(groupId string) int64 {
	if v, ok := c.InternalCache.Load(groupId); ok {
		return v
	}
	return 0
}

func (c *GroupSessionVersionsCache) Store(groupId string) {
	if groupId != "" {
		c.InternalCache.Store(groupId, time.Now().UnixNano())
	}
}

type GroupSubGroupsCache struct {
	InternalCache[string, []string]
}

func NewGroupSubGroupsCache() *GroupSubGroupsCache {
	return &GroupSubGroupsCache{
		InternalCache: NewInternalCache[string, []string](),
	}
}

type Cache struct {
	SessionTokens        *SessionTokensCache
	Groups               *GroupsCache
	SubGroups            *SubGroupsCache
	GroupSessionVersions *GroupSessionVersionsCache
	GroupSubGroups       *GroupSubGroupsCache
}

func NewCache() *Cache {
	return &Cache{
		SessionTokens:        NewSessionTokensCache(),
		Groups:               NewGroupsCache(),
		SubGroups:            NewSubGroupsCache(),
		GroupSessionVersions: NewGroupSessionVersionsCache(),
		GroupSubGroups:       NewGroupSubGroupsCache(),
	}
}
