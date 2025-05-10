package handlers

import (
	"sync"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

type HandlerCache struct {
	sessionTokens        sync.Map // map[string]*types.UserSession
	groupSessionVersions sync.Map // map[string]int64
	groups               sync.Map // map[string]*types.CachedGroup
	subGroups            sync.Map // map[string]*types.CachedSubGroup
}

func (hc *HandlerCache) SetSessionToken(token string, session *types.UserSession) {
	hc.sessionTokens.Store(token, session)
}

func (hc *HandlerCache) GetSessionToken(token string) *types.UserSession {
	s, ok := hc.sessionTokens.Load(token)
	if !ok {
		return nil
	}
	return s.(*types.UserSession)
}

func (hc *HandlerCache) SetGroupSessionVersion(groupId string) {
	hc.groupSessionVersions.Store(groupId, time.Now().UnixNano())
}

func (hc *HandlerCache) GetGroupSessionVersion(groupId string) int64 {
	v, ok := hc.groupSessionVersions.Load(groupId)
	if !ok {
		return 0
	}
	return v.(int64)
}

func (hc *HandlerCache) SetCachedGroup(groupPath, id, externalId, sub, name string, ai bool) {
	hc.groups.Store(groupPath, &types.CachedGroup{
		Id:         id,
		ExternalId: externalId,
		Sub:        sub,
		Name:       name,
		Ai:         ai,
	})
}

func (hc *HandlerCache) GetCachedGroup(groupPath string) *types.CachedGroup {
	v, ok := hc.groups.Load(groupPath)
	if !ok {
		return nil
	}
	return v.(*types.CachedGroup)
}

func (hc *HandlerCache) UnsetCachedGroup(groupPath string) {
	hc.groups.Delete(groupPath)
}

func (hc *HandlerCache) SetCachedSubGroup(subGroupPath, externalId, name, groupPath string) {
	hc.subGroups.Store(subGroupPath, &types.CachedSubGroup{
		ExternalId: externalId,
		GroupPath:  groupPath,
		Name:       name,
	})
}

func (hc *HandlerCache) GetCachedSubGroup(subGroupPath string) *types.CachedSubGroup {
	v, ok := hc.subGroups.Load(subGroupPath)
	if !ok {
		return nil
	}
	return v.(*types.CachedSubGroup)
}

func (hc *HandlerCache) UnsetCachedSubGroup(subGroupPath string) {
	hc.subGroups.Delete(subGroupPath)
}
