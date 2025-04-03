package handlers

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) CheckGroupName(w http.ResponseWriter, req *http.Request, data *types.CheckGroupNameRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.CheckGroupNameResponse, error) {
	var count int

	time.Sleep(time.Second)

	err := tx.QueryRow(`
		SELECT COUNT(*) FROM dbtable_schema.groups WHERE name = $1
	`, data.GetName()).Scan(&count)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.CheckGroupNameResponse{IsValid: count == 0}, nil
}

func (h *Handlers) JoinGroup(w http.ResponseWriter, req *http.Request, data *types.JoinGroupRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.JoinGroupResponse, error) {
	var userId, groupId, allowedDomains, defaultRoleId string

	err := tx.QueryRow(`
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Group not found."))
	}

	err = tx.QueryRow(`SELECT id FROM dbtable_schema.users WHERE sub = $1`, session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if allowedDomains != "" {
		allowedDomainsSlice := strings.Split(allowedDomains, ",")
		if !slices.Contains(allowedDomainsSlice, strings.Split(session.UserEmail, "@")[1]) {
			return nil, util.ErrCheck(util.UserError("Group access is restricted."))
		}
	}

	err = tx.SetDbVar("group_id", groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var kcSubgroupExternalId string
	err = tx.QueryRow(`
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE role_id = $1
	`, defaultRoleId).Scan(&kcSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		INSERT INTO dbtable_schema.group_users (user_id, group_id, external_id, created_sub)
		VALUES ($1, $2, $3, $4::uuid)
	`, userId, groupId, kcSubgroupExternalId, session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.SetGroupSessionVersion(req.Context(), groupId)
	h.Redis.DeleteSession(req.Context(), session.UserSub)

	return &types.JoinGroupResponse{Success: true}, nil
}

func (h *Handlers) LeaveGroup(w http.ResponseWriter, req *http.Request, data *types.LeaveGroupRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.LeaveGroupResponse, error) {
	var userId, groupId, allowedDomains, defaultRoleId string

	err := tx.QueryRow(`
		SELECT id, allowed_domains, default_role_id FROM dbtable_schema.groups WHERE code = $1
	`, data.GetCode()).Scan(&groupId, &allowedDomains, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(util.UserError("Group not found."))
	}

	err = tx.QueryRow(`SELECT id FROM dbtable_schema.users WHERE sub = $1`, session.UserSub).Scan(&userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		DELETE FROM dbtable_schema.group_users WHERE user_id = $1 AND group_id = $2
	`, userId, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.DeleteUserFromGroup(session.UserSub, groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.LeaveGroupResponse{Success: true}, nil
}

// AttachUser
func (h *Handlers) AttachUser(w http.ResponseWriter, req *http.Request, data *types.AttachUserRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.AttachUserResponse, error) {
	var groupId, kcGroupExternalId, kcRoleSubgroupExternalId, defaultRoleId, createdSub string

	err := tx.QueryRow(`SELECT g.id FROM dbtable_schema.groups g WHERE g.code = $1`, data.GetCode()).Scan(&groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.SetDbVar("group_id", groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = tx.QueryRow(`
		SELECT g.external_id, g.default_role_id, g.created_sub, gr.external_id FROM dbtable_schema.groups g
		JOIN dbtable_schema.group_roles gr ON gr.role_id = g.default_role_id
		WHERE g.id = $1
	`, groupId).Scan(&kcGroupExternalId, &defaultRoleId, &createdSub, &kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddUserToGroup(session.UserSub, kcRoleSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if err := h.Socket.RoleCall(session.UserSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.Client().Del(req.Context(), createdSub+"profile/details")
	h.Redis.DeleteSession(req.Context(), session.UserSub)
	h.Redis.SetGroupSessionVersion(req.Context(), groupId)

	return &types.AttachUserResponse{Success: true}, nil
}

func (h *Handlers) CompleteOnboarding(w http.ResponseWriter, req *http.Request, data *types.CompleteOnboardingRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.CompleteOnboardingResponse, error) {
	service := data.GetService()
	schedule := data.GetSchedule()

	postServiceRes, err := h.PostService(w, req, &types.PostServiceRequest{Service: service}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	service.Id = postServiceRes.GetId()

	_, err = h.PostGroupSchedule(w, req, &types.PostGroupScheduleRequest{GroupSchedule: &types.IGroupSchedule{Schedule: schedule}}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.ActivateProfile(w, req, &types.ActivateProfileRequest{}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.CompleteOnboardingResponse{Success: true}, nil
}
