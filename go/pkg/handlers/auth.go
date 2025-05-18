package handlers

import (
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) AuthWebhook_REGISTER(info ReqInfo, authEvent *types.AuthEvent) (*types.AuthWebhookResponse, error) {
	_, err := h.PostUserProfile(info, &types.PostUserProfileRequest{
		FirstName: authEvent.FirstName,
		LastName:  authEvent.LastName,
		Username:  authEvent.Email,
		Email:     authEvent.Email,
		Timezone:  authEvent.Timezone,
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if authEvent.GroupCode != "" {
		_, err := h.JoinGroup(info, &types.JoinGroupRequest{Code: authEvent.GroupCode})
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = h.ActivateProfile(info, &types.ActivateProfileRequest{})
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.AuthWebhookResponse{Value: `{ "success": true }`}, nil
}

func (h *Handlers) AuthWebhook_REGISTER_VALIDATE(info ReqInfo, authEvent *types.AuthEvent) (*types.AuthWebhookResponse, error) {

	group := &types.IGroup{}
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT id, default_role_id, name, allowed_domains
		FROM dbtable_schema.groups
		WHERE code = $1
	`, authEvent.GroupCode).Scan(&group.Id, &group.DefaultRoleId, &group.Name, &group.AllowedDomains)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if group.Name == "" {
		return &types.AuthWebhookResponse{Value: `{ "success": false, "reason": "invalid group code" }`}, nil
	}

	info.Session.SetGroupId(group.GetId())
	ds := clients.NewGroupDbSession(h.Database.DatabaseClient.Pool, info.Session)

	var kcRoleSubgroupExternalId string

	row, done, err := ds.SessionBatchQueryRow(info.Ctx, `
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE group_id = $1 AND role_id = $2
	`, group.GetId(), group.GetDefaultRoleId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer done()

	err = row.Scan(&kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var authRes strings.Builder
	authRes.WriteString(`{ "success": true, "name": "`)
	authRes.WriteString(group.Name)
	authRes.WriteString(`", "allowedDomains": "`)
	authRes.WriteString(group.AllowedDomains)
	authRes.WriteString(`", "id": "`)
	authRes.WriteString(kcRoleSubgroupExternalId)
	authRes.WriteString(`" }`)

	return &types.AuthWebhookResponse{Value: authRes.String()}, nil
}
