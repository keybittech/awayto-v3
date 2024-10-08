package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"fmt"
	"net/http"
)

func (h *Handlers) AuthWebhook_REGISTER(req *http.Request, authEvent clients.AuthEvent) (string, error) {

	userProfile := &types.PostUserProfileRequest{}

	userProfile.FirstName = authEvent.FirstName
	userProfile.LastName = authEvent.LastName
	userProfile.Username = authEvent.Email
	userProfile.Email = authEvent.Email

	_, err := h.PostUserProfile(nil, req, userProfile)
	if err != nil {
		return "", util.ErrCheck(err)
	}

	if authEvent.GroupCode != "" {
		_, err := h.JoinGroup(nil, req, &types.JoinGroupRequest{Code: authEvent.GroupCode})
		if err != nil {
			return "", util.ErrCheck(err)
		}

		_, err = h.AttachUser(nil, req, &types.AttachUserRequest{Code: authEvent.GroupCode})
		if err != nil {
			return "", util.ErrCheck(err)
		}

		_, err = h.ActivateProfile(nil, req, &types.ActivateProfileRequest{})
		if err != nil {
			return "", util.ErrCheck(err)
		}

	}

	return `{ "success": true }`, nil
}

func (h *Handlers) AuthWebhook_REGISTER_VALIDATE(req *http.Request, authEvent clients.AuthEvent) (string, error) {

	group := &types.IGroup{}

	err := h.Database.Client().QueryRow(`
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

	var kcRoleSubgroupExternalId string

	err = h.Database.Client().QueryRow(`
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE group_id = $1 AND role_id = $2
	`, group.GetId(), group.GetDefaultRoleId()).Scan(&kcRoleSubgroupExternalId)
	if err != nil {
		return "", util.ErrCheck(err)
	}

	return fmt.Sprintf(`{ "success": true, "name": "%s", "allowedDomains": "%s", "id": "%s" }`, group.Name, group.AllowedDomains, kcRoleSubgroupExternalId), nil
}
