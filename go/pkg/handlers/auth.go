package handlers

import (
	"fmt"
	"net/http"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) AuthWebhook_REGISTER(req *http.Request, authEvent *types.AuthEvent, session *types.UserSession, tx *clients.PoolTx) (string, error) {

	_, err := h.PostUserProfile(nil, req, &types.PostUserProfileRequest{
		FirstName: authEvent.FirstName,
		LastName:  authEvent.LastName,
		Username:  authEvent.Email,
		Email:     authEvent.Email,
		Timezone:  authEvent.Timezone,
	}, session, tx)
	if err != nil {
		return "", util.ErrCheck(err)
	}

	if authEvent.GroupCode != "" {
		_, err := h.JoinGroup(nil, req, &types.JoinGroupRequest{Code: authEvent.GroupCode}, session, tx)
		if err != nil {
			return "", util.ErrCheck(err)
		}

		_, err = h.ActivateProfile(nil, req, &types.ActivateProfileRequest{}, session, tx)
		if err != nil {
			return "", util.ErrCheck(err)
		}
	}

	return `{ "success": true }`, nil
}

func (h *Handlers) AuthWebhook_REGISTER_VALIDATE(req *http.Request, authEvent *types.AuthEvent, session *types.UserSession, tx *clients.PoolTx) (string, error) {

	group := &types.IGroup{}
	err := tx.QueryRow(req.Context(), `
		SELECT id, default_role_id, name, allowed_domains
		FROM dbtable_schema.groups
		WHERE code = $1
	`, authEvent.GroupCode).Scan(&group.Id, &group.DefaultRoleId, &group.Name, &group.AllowedDomains)
	if err != nil {
		return "", util.ErrCheck(err)
	}

	if group.Name == "" {
		return `{ "success": false, "reason": "invalid group code" }`, nil
	}

	session.GroupId = group.GetId()
	ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, session)

	var kcRoleSubgroupExternalId string

	row, done, err := ds.SessionBatchQueryRow(req.Context(), `
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE group_id = $1 AND role_id = $2
	`, group.GetId(), group.GetDefaultRoleId())
	if err != nil {
		return "", util.ErrCheck(err)
	}
	defer done()

	err = row.Scan(&kcRoleSubgroupExternalId)
	if err != nil {
		return "", util.ErrCheck(err)
	}

	return fmt.Sprintf(`{ "success": true, "name": "%s", "allowedDomains": "%s", "id": "%s" }`, group.Name, group.AllowedDomains, kcRoleSubgroupExternalId), nil
}
