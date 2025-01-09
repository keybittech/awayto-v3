package clients

import (
	"av3api/pkg/util"
	"context"
	"encoding/json"
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
	Nonce                   string   `json:"nonce"`
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
		return nil, util.ErrCheck(err)
	}

	var userSession UserSession
	err = json.Unmarshal([]byte(sessionBytes), &userSession)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	r.Client().Expire(ctx, "user_session:"+userSub, sessionDuration)

	return &userSession, nil
}

func (r *Redis) SetSession(ctx context.Context, userSub string, session *UserSession) error {
	sessionJson, err := json.Marshal(session)
	if err != nil {
		return util.ErrCheck(err)
	}
	return r.Client().Set(ctx, "user_session:"+userSub, sessionJson, sessionDuration).Err()
}

func (r *Redis) DeleteSession(ctx context.Context, userSub string) error {
	return r.Client().Del(ctx, "user_session:"+userSub).Err()
}

func (r *Redis) ReqSession(req *http.Request) *UserSession {
	if userSession := req.Context().Value("UserSession"); userSession != nil {
		return userSession.(*UserSession)
	} else {
		return &UserSession{}
	}
}
