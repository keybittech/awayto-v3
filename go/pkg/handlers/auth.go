package handlers

import (
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) AuthWebhook_REGISTER(info ReqInfo, authEvent *types.AuthEvent) (*types.AuthWebhookResponse, error) {

	tx, err := h.Database.DatabaseClient.Pool.Begin(info.Ctx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer tx.Rollback(info.Ctx)

	poolTx := &clients.PoolTx{
		Tx: tx,
	}

	info.Tx = poolTx

	err = poolTx.SetSession(info.Ctx, info.Session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.PostUserProfile(info, &types.PostUserProfileRequest{
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
		_, err := h.JoinGroup(info, &types.JoinGroupRequest{
			Code:        authEvent.GroupCode,
			Registering: true,
		})
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	err = poolTx.UnsetSession(info.Ctx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = poolTx.Commit(info.Ctx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.AuthWebhookResponse{Value: `{ "success": true }`}, nil
}

// Validating registration is done by taking the group code on the registration screen
// then using it to fetch the basic details about the group which are used during
// registration in keycloak (limiting email domains, joining group). The implication is
// a user is joining an existing group, so we need to get the *external_id of the default_role_id*
// of the group, as roles are "roles" in our app, but also "groups" in keycloak.
func (h *Handlers) AuthWebhook_REGISTER_VALIDATE(info ReqInfo, authEvent *types.AuthEvent) (msg *types.AuthWebhookResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				util.ErrorLog.Println(util.ErrCheck(err))
			}
			msg = &types.AuthWebhookResponse{Value: `{ "success": false, "reason": "BAD_GROUP" }`}
		}
	}()

	batch := util.NewBatchable(h.Database.DatabaseClient.Pool, "worker", "", 0)

	// Cast the check for roleId against defaultRoleId as defaultRoleId could be empty
	groupLookup := util.BatchQueryRow[types.IGroup](batch, `
		SELECT eg."displayName", eg."allowedDomains", egr."externalId"
		FROM dbview_schema.enabled_groups eg
		JOIN dbview_schema.enabled_group_roles egr ON egr."groupId" = eg.id AND egr."roleId"::TEXT = eg."defaultRoleId"
		WHERE code = $1
	`, authEvent.GroupCode)

	batch.Send(info.Ctx)

	group := *groupLookup

	var authRes strings.Builder
	authRes.WriteString(`{ "success": true, "groupName": "`)
	authRes.WriteString(group.GetDisplayName())
	authRes.WriteString(`", "allowedDomains": "`)
	authRes.WriteString(group.GetAllowedDomains())
	authRes.WriteString(`", "roleGroupId": "`)
	// using group for convenience but this is the *external id of the group's default role*
	authRes.WriteString(group.GetExternalId())
	authRes.WriteString(`" }`)

	return &types.AuthWebhookResponse{Value: authRes.String()}, nil
}
