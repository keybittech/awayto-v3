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
)

func (h *Handlers) GenerateLoginOrRegisterParams(req *http.Request) string {
	tz := req.URL.Query().Get("tz")
	ua := req.Header.Get("User-Agent")
	if tz == "" || ua == "" {
		return ""
	}

	codeVerifier := util.GenerateCodeVerifier()
	codeChallenge := util.GenerateCodeChallenge(codeVerifier)
	state := util.GenerateState()

	tempSession := types.NewConcurrentTempAuthSession(&types.TempAuthSession{
		CodeVerifier: codeVerifier,
		State:        state,
		CreatedAt:    time.Now().UnixNano(),
		Tz:           tz,
		Ua:           ua,
	})

	h.Cache.TempAuthSessions.Store(state, tempSession)

	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {util.E_KC_USER_CLIENT},
		"redirect_uri":          {util.E_APP_HOST_URL + "/auth/callback"},
		"scope":                 {"openid profile email groups"},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	return "?" + params.Encode()
}

func (h *Handlers) GetCachedGroups(ctx context.Context, subGroupPath string) (ccg *types.ConcurrentCachedGroup, ccsg *types.ConcurrentCachedSubGroup, err error) {
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
			Id:            group.GetId(),
			Code:          group.GetCode(),
			Name:          groupName,
			ExternalId:    group.GetExternalId(),
			Sub:           group.GetSub(),
			Path:          groupPath,
			Ai:            group.GetAi(),
		}

		for _, subGroup := range kcSubGroups {
			cachedGroup.SubGroupPaths = append(cachedGroup.SubGroupPaths, subGroup.GetPath())
		}

		if _, found := h.Cache.GroupSessionVersions.Load(group.Id); !found {
			h.Cache.GroupSessionVersions.Store(group.GetId(), time.Now().UnixNano())
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
			ExternalId: subGroup.GetExternalId(),
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

func (h *Handlers) GetSession(req *http.Request, userSub ...string) (concurrentSession *types.ConcurrentUserSession, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during get session, err: %v", r)
		}
	}()

	var sessionId string
	var found bool

	ctx := req.Context()

	// If userSub is passed, current user is refreshing another's token
	// so check whose session we need to get.
	if len(userSub) > 0 {
		sessionId, found = h.Cache.UserSessionIds.Load(userSub[0])
		if !found {
			batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
			sessionIdReq := util.BatchQueryRow[types.ILookup](batch, `
				SELECT id
				FROM dbtable_schema.user_sessions
				WHERE sub = $1
			`, userSub[0])
			batch.Send(ctx)

			if sessionIdRes := *sessionIdReq; sessionIdRes != nil {
				sessionId = sessionIdRes.GetId()
			}
		}
	} else {
		sessionId = util.GetSessionIdFromCookie(req)
	}
	if sessionId == "" {
		return nil, util.ErrCheck(errors.New("no session id to get user session"))
	}

	concurrentSession, found = h.Cache.UserSessions.Get(sessionId)
	if !found {
		batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
		sessionReq := util.BatchQueryRow[types.UserSession](batch, `
			SELECT id
			FROM dbtable_schema.user_sessions
			WHERE id = $1
		`, sessionId)
		batch.Send(ctx)

		session := *sessionReq

		if session == nil || session.GetId() == "" {
			err = util.ErrCheck(fmt.Errorf("could not find session in cache or db, sessionId: %s", sessionId))
			return
		}

		// Force a refresh if we had to go to the db to get session,
		// as the db only stores a partial record of the session and
		// we want to make sure the token claims are attached
		concurrentSession, err = h.RefreshSession(req, types.NewConcurrentUserSession(session))
		if err != nil {
			err = util.ErrCheck(fmt.Errorf("could not refresh session during db reconstruct, err: %v", err))
			return
		}
	}

	return
}

func (h *Handlers) RefreshOrDelete(req *http.Request, session *types.ConcurrentUserSession) *types.ConcurrentUserSession {
	refreshedSession, err := h.RefreshSession(req, session)
	if err != nil {
		util.ErrorLog.Printf("could not refresh token during access expired, err: %v", err)
		if err := h.DeleteSession(req.Context(), session.GetId()); err != nil {
			util.ErrorLog.Printf("could not delete session during token refresh fail, err: %v", err)
		}
	}
	return refreshedSession
}

func (h *Handlers) CheckSessionExpiry(req *http.Request, session *types.ConcurrentUserSession) *types.ConcurrentUserSession {
	if time.Now().After(time.Unix(0, session.GetAccessExpiresAt()).Add(-30 * time.Second)) {
		session = h.RefreshOrDelete(req, session)
	}
	if session == nil {
		return nil
	}

	if groupId := session.GetGroupId(); groupId != "" {
		currentGroupVersion, _ := h.Cache.GroupSessionVersions.Load(groupId)
		if currentGroupVersion > 0 && session.GetGroupSessionVersion() < currentGroupVersion {
			session = h.RefreshOrDelete(req, session)
		}
	}

	return session
}

func (h *Handlers) ResetGroupSession(req *http.Request, groupId string) {
	if groupId == "" {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("failed to reset group session, err %v", r)
			util.ErrorLog.Println(util.ErrCheck(err))
		}
	}()

	// Forces downstream updates to include group update
	newGroupVersion := time.Now().UnixNano()
	h.Cache.GroupSessionVersions.Store(groupId, newGroupVersion)

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
	sessionSubsReq := util.BatchQuery[types.ILookup](batch, `
		SELECT sub as id
		FROM dbtable_schema.user_sessions
		WHERE group_id = $1
	`, groupId)
	batch.Send(req.Context())

	for _, sub := range *sessionSubsReq {
		userSession, err := h.GetSession(req, sub.GetId())
		if err != nil {
			util.ErrorLog.Printf("failed to get session during reset group session, userSub: %s, groupId: %s, err: %v", sub.GetId(), groupId, err)
		}
		// Version check here saves us from refreshing again if we already did refresh in GetSession
		if userSession != nil && userSession.GetGroupSessionVersion() != newGroupVersion {
			_, err = h.RefreshSession(req, userSession)
			if err != nil {
				util.ErrorLog.Printf("failed to refresh session during reset group session, userSub: %s, groupId: %s, err: %v", sub.GetId(), groupId, err)
			}
		}
	}

	err := h.Socket.GroupRoleCall("worker", groupId)
	if err != nil {
		util.ErrorLog.Printf("failed to group role call during reset group session, groupId: %s, err: %v", groupId, err)
	}
}

// TODO can keycloak batch session info or obviate token completely?
func (h *Handlers) RefreshSession(req *http.Request, session *types.ConcurrentUserSession) (*types.ConcurrentUserSession, error) {
	refreshToken := session.GetRefreshToken()
	if refreshToken == "" {
		return nil, util.ErrCheck(errors.New("refresh token is empty"))
	}

	refreshedSession, err := util.GetValidTokenRefresh(req, refreshToken, session.GetUserAgent(), session.GetTimezone(), session.GetAnonIp())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	refreshedSession.Id = session.GetId()

	return h.StoreSession(req.Context(), refreshedSession)
}

// Existing sessions can provide sessionId to overwrite existing
func (h *Handlers) StoreSession(ctx context.Context, session *types.UserSession) (concurrentSession *types.ConcurrentUserSession, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = util.ErrCheck(fmt.Errorf("failed to store session, err: %v", r))
		}
	}()

	var groupId string
	if len(session.GetSubGroupPaths()) > 0 {
		subGroupPath := session.GetSubGroupPaths()[0]

		concurrentCachedGroup, concurrentCachedSubGroup, cacheErr := h.GetCachedGroups(ctx, subGroupPath)
		if cacheErr != nil {
			err = util.ErrCheck(cacheErr)
			return
		}

		groupId = concurrentCachedGroup.GetId()

		sessionGroupVersion := session.GetGroupSessionVersion()

		currentGroupVersion, _ := h.Cache.GroupSessionVersions.Load(groupId)

		if sessionGroupVersion == 0 || sessionGroupVersion < currentGroupVersion {
			session.GroupSessionVersion = currentGroupVersion

			session.GroupId = groupId
			session.GroupPath = concurrentCachedGroup.GetPath()
			session.GroupCode = concurrentCachedGroup.GetCode()
			session.GroupName = concurrentCachedGroup.GetName()
			session.GroupExternalId = concurrentCachedGroup.GetExternalId()
			session.GroupSub = concurrentCachedGroup.GetSub()
			session.GroupAi = concurrentCachedGroup.GetAi()

			session.SubGroupPath = subGroupPath
			session.SubGroupName = concurrentCachedSubGroup.GetName()
			session.SubGroupExternalId = concurrentCachedSubGroup.GetExternalId()
		}
	}

	params := []any{
		nil,                      // 1
		session.GetIdToken(),     // 2
		session.GetAccessToken(), // 3
		time.Unix(0, session.GetAccessExpiresAt()),  // 4
		session.GetRefreshToken(),                   // 5
		time.Unix(0, session.GetRefreshExpiresAt()), // 6
		session.GetAnonIp(),                         // 7
		session.GetTimezone(),                       // 8
		session.GetUserAgent(),                      // 9
		groupId,                                     // 10
	}

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)
	if sid := session.GetId(); sid != "" {
		params[0] = sid
		util.BatchExec(batch, `
			UPDATE dbtable_schema.user_sessions
			SET id_token = $2, access_token = $3, access_expires_at = $4, refresh_token = $5, refresh_expires_at = $6, ip_address = $7, timezone = $8, user_agent = $9, group_id = $10
			WHERE id = $1
		`, params...)
		batch.Send(ctx)
	} else {

		params[0] = session.GetUserSub()
		dbSessionInsert := util.BatchQueryRow[types.ILookup](batch, `
			INSERT INTO dbtable_schema.user_sessions (sub, id_token, access_token, access_expires_at, refresh_token, refresh_expires_at, ip_address, timezone, user_agent, group_id)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`, params...)
		batch.Send(ctx)
		dbSession := *dbSessionInsert
		session.Id = dbSession.GetId()
	}

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
