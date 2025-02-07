package clients

import (
	"av3api/pkg/util"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const sessionDuration = 24 * 30 * time.Hour

type UserSession struct {
	UserSub                 string   `json:"userSub"`
	UserEmail               string   `json:"userEmail"`
	GroupSessionVersion     int64    `json:"groupSessionVersion"`
	GroupName               string   `json:"groupName"`
	GroupId                 string   `json:"groupId"`
	GroupSub                string   `json:"groupSub"`
	GroupExternalId         string   `json:"groupExternalId"`
	GroupAi                 bool     `json:"ai"`
	SubGroups               []string `json:"subGroups"`
	SubGroupName            string   `json:"subGroupName"`
	SubGroupExternalId      string   `json:"subGroupExternalId"`
	RoleName                string   `json:"roleName"`
	AnonIp                  string   `json:"anonIp"`
	AvailableUserGroupRoles []string `json:"availableUserGroupRoles"`
}

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

func (r *Redis) GetSession(ctx context.Context, userSub string) (*UserSession, error) {
	sessionBytes, err := r.Client().Get(ctx, "user_session:"+userSub).Result()
	if err != nil {
		return nil, err
	}

	var userSession UserSession
	err = json.Unmarshal([]byte(sessionBytes), &userSession)
	if err != nil {
		return nil, err
	}

	r.Client().Expire(ctx, "user_session:"+userSub, sessionDuration)

	return &userSession, nil
}

func (r *Redis) SetSession(ctx context.Context, userSub string, session *UserSession) error {
	sessionJson, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.Client().Set(ctx, "user_session:"+userSub, sessionJson, sessionDuration).Err()
}

func (r *Redis) DeleteSession(ctx context.Context, userSub string) error {
	return r.Client().Del(ctx, "user_session:"+userSub).Err()
}

func (r *Redis) ReqSession(req *http.Request) (*UserSession, error) {

	token, ok := req.Header["Authorization"]
	if !ok {
		return nil, errors.New("no auth token")
	}

	userToken, err := ParseJWT(token[0])
	if err != nil {
		return nil, err
	}

	session, err := r.GetSession(req.Context(), userToken.Sub)
	if session == nil && err != nil && err.Error() != "redis: nil" {
		return nil, err
	}

	if session == nil {
		session = &UserSession{
			UserSub:   userToken.Sub,
			UserEmail: userToken.Email,
			SubGroups: userToken.Groups,
			AnonIp:    util.AnonIp(req.RemoteAddr),
		}

		err = r.SetSession(req.Context(), userToken.Sub, session)
		if err != nil {
			return nil, err
		}
	}

	return session, nil
}
