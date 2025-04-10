package clients

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

const sessionDuration = 24 * 30 * time.Hour

func (r *Redis) SetGroupSessionVersion(ctx context.Context, groupId string) (int64, error) {
	newVersion := time.Now().UTC().UnixMilli()
	scmd := r.Client().Set(ctx, "group_session_version:"+groupId, newVersion, sessionDuration)
	if scmd.Err() != nil {
		return 0, scmd.Err()
	}
	return newVersion, nil
}

func (r *Redis) GetGroupSessionVersion(ctx context.Context, groupId string) (int64, error) {
	groupVersion, err := r.Client().Get(ctx, "group_session_version:"+groupId).Int64()
	if err != nil {
		return r.SetGroupSessionVersion(ctx, groupId)
	}
	return groupVersion, nil
}

func (r *Redis) GetSession(ctx context.Context, userSub string) (*types.UserSession, error) {
	sessionBytes, err := r.Client().Get(ctx, "user_session:"+userSub).Bytes()
	if err != nil {
		return nil, err
	}

	var userSession types.UserSession
	err = json.Unmarshal(sessionBytes, &userSession)
	if err != nil {
		return nil, err
	}

	// not sure if this is needed
	// r.Client().Expire(ctx, "user_session:"+userSub, sessionDuration)

	return &userSession, nil
}

func (r *Redis) SetSession(ctx context.Context, session *types.UserSession) error {
	sessionJson, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.Client().Set(ctx, "user_session:"+session.UserSub, sessionJson, sessionDuration).Err()
}

func (r *Redis) DeleteSession(ctx context.Context, userSub string) error {
	return r.Client().Del(ctx, "user_session:"+userSub).Err()
}
