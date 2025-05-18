package handlers

// type HandlerCache struct {
// }
//
// func (stc *SessionTokensCache) Store(token string, session *types.UserSession) {
// 	thing := types.NewConcurrentUserSession(session)
// 	thing.SetAnonIp("")
// 	stc.Map.Store(token, session)
// }
//
// func (stc *SessionTokensCache) Load(token string) *types.UserSession {
// 	s, ok := stc.Map.Load(token)
// 	if !ok {
// 		return nil
// 	}
// 	return s.(*types.UserSession)
// }
//
// func (gsvc *GroupSessionVersionsCache) Store(groupId string) {
// 	gsvc.Map.Store(groupId, time.Now().UnixNano())
// }
//
// func (gsvc *GroupSessionVersionsCache) Load(groupId string) int64 {
// 	v, ok := gsvc.Map.Load(groupId)
// 	if !ok {
// 		return 0
// 	}
// 	return v.(int64)
// }
//
// func (gc *GroupsCache) Store(groupPath string, group *types.CachedGroup) {
// 	gc.Map.Store(groupPath, group)
// }
//
// func (gc *GroupsCache) Load(groupPath string) *types.CachedGroup {
// 	v, ok := gc.Map.Load(groupPath)
// 	if !ok {
// 		return nil
// 	}
// 	proto.Clone(v.(proto.Message))
// 	return v.(*types.CachedGroup)
// }
//
// func (hc *HandlerCache) UpsertCachedGroup(groupPath, id, externalId, sub, name string, ai bool, subGroupPaths []string) {
// 	group := hc.GetCachedGroup(groupPath)
// 	if group == nil {
// 		hc.groups.Store(groupPath, &types.CachedGroup{
// 			Id:            id,
// 			ExternalId:    externalId,
// 			Sub:           sub,
// 			Name:          name,
// 			Ai:            ai,
// 			SubGroupPaths: subGroupPaths,
// 		})
// 		return
// 	}
//
// 	if name != group.Name {
// 		subGroupPaths := make([]string, len(subGroupPaths))
//
// 	}
// }
//
// func (hc *HandlerCache) GetCachedGroup(groupPath string) *types.CachedGroup {
// 	v, ok := hc.groups.Load(groupPath)
// 	if !ok {
// 		return nil
// 	}
// 	return v.(*types.CachedGroup)
// }
//
// func (hc *HandlerCache) UnsetCachedGroup(groupPath string) {
// 	hc.groups.Delete(groupPath)
// }
//
// func (hc *HandlerCache) SetCachedSubGroup(subGroupPath, externalId, name, groupPath string) {
// 	hc.subGroups.Store(subGroupPath, &types.CachedSubGroup{
// 		ExternalId: externalId,
// 		GroupPath:  groupPath,
// 		Name:       name,
// 	})
// }
//
// func (hc *HandlerCache) GetCachedSubGroup(subGroupPath string) *types.CachedSubGroup {
// 	v, ok := hc.subGroups.Load(subGroupPath)
// 	if !ok {
// 		return nil
// 	}
// 	return v.(*types.CachedSubGroup)
// }
//
// func (hc *HandlerCache) UnsetCachedSubGroup(subGroupPath string) {
// 	hc.subGroups.Delete(subGroupPath)
// }
//
// func (hc *HandlerCache) SetGroupSubGroups(groupPath string, subGroupPaths []string) {
// 	hc.groupSubGroups.Store(groupPath, subGroupPaths)
// }
//
// func (hc *HandlerCache) GetGroupSubGroups(groupPath string) []string {
// 	v, ok := hc.groupSubGroups.Load(groupPath)
// 	if !ok {
// 		return nil
// 	}
// 	return v.([]string)
// }
