package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

func (h *Handlers) CheckGroupInfo(ctx context.Context, subGroupPath string) (ccg *types.ConcurrentCachedGroup, ccsg *types.ConcurrentCachedSubGroup, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = util.ErrCheck(fmt.Errorf("failed to check group info, subGroupPath %s, err: %v", subGroupPath, r))
		}
	}()

	subGroupPathIdx := strings.LastIndex(subGroupPath, "/")
	subGroupName := subGroupPath[subGroupPathIdx+1:]

	groupPath := subGroupPath[:subGroupPathIdx]
	groupName := groupPath[1:]

	concurrentCachedGroup, err := h.Cache.Groups.LoadOrSet(groupPath, func() (*types.CachedGroup, error) {
		batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
		groupReq := util.BatchQueryRow[types.IGroup](batch, `
			SELECT id, code, external_id as "externalId", sub, ai
			FROM dbtable_schema.groups
			WHERE name = $1
		`, groupName)
		batch.Send(ctx)

		group := *groupReq

		kcSubGroups, err := h.Keycloak.GetGroupSubGroups(ctx, "worker", group.ExternalId)
		if err != nil {
			return nil, err
		}

		cachedGroup := &types.CachedGroup{
			SubGroupPaths: make([]string, 0, len(kcSubGroups)),
			Id:            group.Id,
			Code:          group.Code,
			Name:          groupName,
			ExternalId:    group.ExternalId,
			Sub:           group.Sub,
			Path:          groupPath,
			Ai:            group.Ai,
		}

		for _, subGroup := range kcSubGroups {
			cachedGroup.SubGroupPaths = append(cachedGroup.SubGroupPaths, subGroup.Path)
		}

		if _, found := h.Cache.GroupSessionVersions.Load(group.Id); !found {
			h.Cache.GroupSessionVersions.Store(group.Id, time.Now().UnixNano())
		}

		return cachedGroup, nil
	})
	if err != nil {
		return nil, nil, util.ErrCheck(err)
	}

	concurrentCachedSubGroup, err := h.Cache.SubGroups.LoadOrSet(subGroupPath, func() (*types.CachedSubGroup, error) {
		batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
		subGroupReq := util.BatchQueryRow[types.IGroupRole](batch, `
			SELECT egr."externalId"
			FROM dbview_schema.enabled_roles er
			JOIN dbview_schema.enabled_group_roles egr ON egr."roleId" = er.id
			JOIN dbview_schema.enabled_groups eg ON eg.id = egr."groupId"
			WHERE er.name = $1 AND eg.name = $2
		`, subGroupName, groupName)

		batch.Send(ctx)

		subGroup := *subGroupReq

		cachedSubGroup := &types.CachedSubGroup{
			ExternalId: subGroup.ExternalId,
			Name:       subGroupName,
			GroupPath:  groupPath,
		}

		return cachedSubGroup, nil
	})
	if err != nil {
		return nil, nil, util.ErrCheck(err)
	}
	return concurrentCachedGroup, concurrentCachedSubGroup, nil
}

func (h *Handlers) StoreSession(ctx context.Context, session *types.UserSession, sessionId ...string) (concurrentSession *types.ConcurrentUserSession, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = util.ErrCheck(fmt.Errorf("failed to store session, err: %v", r))
		}
	}()

	var gid string
	if len(session.GetSubGroupPaths()) > 0 {
		subGroupPath := session.GetSubGroupPaths()[0]

		var concurrentCachedGroup *types.ConcurrentCachedGroup
		concurrentCachedGroup, _, err = h.CheckGroupInfo(ctx, subGroupPath)
		if err != nil {
			err = util.ErrCheck(err)
			return
		}

		gid = concurrentCachedGroup.GetId()
	}

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
	var sid string
	if len(sessionId) > 0 {
		sid = sessionId[0]
		util.BatchExec(batch, `
			UPDATE dbtable_schema.user_sessions
			SET id_token = $2, access_token = $3, access_expires_at = $4, refresh_token = $5, refresh_expires_at = $6, ip_address = $7, timezone = $8, user_agent = $9, group_id = $10
			WHERE id = $1
		`, sid, session.GetIdToken(), session.GetAccessToken(), time.Unix(0, session.GetAccessExpiresAt()), session.GetRefreshToken(), time.Unix(0, session.GetRefreshExpiresAt()), session.GetAnonIp(), session.GetTimezone(), session.GetUserAgent(), gid)
		batch.Send(ctx)
	} else {
		dbSessionInsert := util.BatchQueryRow[types.ILookup](batch, `
			INSERT INTO dbtable_schema.user_sessions (sub, id_token, access_token, access_expires_at, refresh_token, refresh_expires_at, ip_address, timezone, user_agent, group_id)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`, session.GetUserSub(), session.GetIdToken(), session.GetAccessToken(), time.Unix(0, session.GetAccessExpiresAt()), session.GetRefreshToken(), time.Unix(0, session.GetRefreshExpiresAt()), session.GetAnonIp(), session.GetTimezone(), session.GetUserAgent(), gid)
		batch.Send(ctx)
		dbSession := *dbSessionInsert
		sid = dbSession.GetId()
	}

	session.Id = sid

	h.Cache.UserSessionIds.Store(session.GetUserSub(), session.GetId())
	concurrentSession = types.NewConcurrentUserSession(session)
	h.Cache.UserSessions.Store(session.GetId(), concurrentSession)

	return
}

func (h *Handlers) DeleteSession(ctx context.Context, sessionId string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to delete session, err: %v", r)
		}
	}()

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
	util.BatchExec(batch, `
		DELETE FROM dbtable_schema.user_sessions
		WHERE id = $1
	`, sessionId)
	batch.Send(ctx)

	session, ok := h.Cache.UserSessions.Get(sessionId)
	if ok {
		h.Cache.UserSessionIds.Delete(session.GetUserSub())
	}
	h.Cache.UserSessions.Delete(sessionId)

	return nil
}

func (h *Handlers) ResetGroupSession(req *http.Request, groupId string) {
	if groupId == "" {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			util.ErrorLog.Printf("failed to reset group session, err %v", r)
		}
	}()

	h.Cache.GroupSessionVersions.Store(groupId, time.Now().UnixNano())

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
	sessionSubsReq := util.BatchQuery[types.ILookup](batch, `
		SELECT sub as id
		FROM dbtable_schema.user_sessions
		WHERE group_id = $1
	`, groupId)
	batch.Send(req.Context())

	for _, sub := range *sessionSubsReq {
		_, err := h.RefreshSession(req, sub.GetId())
		if err != nil {
			panic(util.ErrCheck(fmt.Errorf("failed to refresh the session for %s", sub.GetId())))
		}
	}

	_ = h.Socket.GroupRoleCall("worker", groupId)
}

// TODO can keycloak batch session info or obviate token completely?

func (h *Handlers) RefreshSession(req *http.Request, userSub ...string) (*types.ConcurrentUserSession, error) {
	var sessionId string
	var exists bool
	if len(userSub) > 0 {
		sessionId, exists = h.Cache.UserSessionIds.Load(userSub[0])
		if !exists {
			return nil, util.ErrCheck(fmt.Errorf("no user session for %s", userSub[0]))
		}
	} else {
		sessionId = util.GetSessionIdFromCookie(req)
		if sessionId == "" {
			return nil, util.ErrCheck(errors.New("no session id to refresh token"))
		}
	}

	session, ok := h.Cache.UserSessions.Get(sessionId)
	if !ok {
		return nil, util.ErrCheck(errors.New("failed to get user session to refresh token"))
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {util.E_KC_USER_CLIENT},
		"client_secret": {util.E_KC_USER_CLIENT_SECRET},
		"refresh_token": {session.GetRefreshToken()},
	}

	util.SetForwardingHeaders(req)

	resp, err := util.PostFormData(util.E_KC_OPENID_TOKEN_URL, req.Header, data)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var tokens types.OIDCToken
	if err := protojson.Unmarshal(resp, &tokens); err != nil {
		return nil, util.ErrCheck(err)
	}

	// If userSub provided, likely admin is refreshing group users, so use existing session info
	var ua, tz, ip string
	if len(userSub) > 0 {
		ua = session.GetUserAgent()
		tz = session.GetTimezone()
		ip = session.GetAnonIp()
	} else {
		ua = util.GetUA(req.Header.Get("User-Agent"))
		tz = req.Header.Get("X-Tz")
		ip = util.AnonIp(req.RemoteAddr)
	}

	refreshedSession, err := util.ValidateToken(&tokens, ua, tz, ip)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	concurrentSession, err := h.StoreSession(req.Context(), refreshedSession, sessionId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return concurrentSession, nil
}
