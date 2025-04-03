package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PatchGroupUser(w http.ResponseWriter, req *http.Request, data *types.PatchGroupUserRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.PatchGroupUserResponse, error) {
	userId := data.GetUserId()
	roleId := data.GetRoleId()

	var groupUsers []*types.IGroupUser

	err := tx.QueryRows(&groupUsers, `
		SELECT g.external_id as "groupExternalId", gu.external_id as "externalId", u.sub as "userSub"
		FROM dbtable_schema.group_users gu
		JOIN dbtable_schema.users u ON u.id = gu.user_id
		JOIN dbtable_schema.groups g ON g.id = gu.group_id
		WHERE g.id = $1 AND gu.user_id = $2
	`, session.GroupId, userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupUsers) == 0 {
		return nil, util.ErrCheck(errors.New("group user not found"))
	}
	groupUser := groupUsers[0]

	var groupRoles []*types.IGroupRole

	err = tx.QueryRows(&groupRoles, `
		SELECT external_id as "externalId"
		FROM dbtable_schema.group_roles
		WHERE group_id = $1 AND role_id = $2
	`, session.GroupId, roleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupRoles) == 0 {
		return nil, util.ErrCheck(errors.New("group role not found"))
	}

	newSubgroupExternalId := groupRoles[0].GetExternalId()

	kcOldSubgroupExternalId := groupUser.GetExternalId()

	groupUserSub := groupUser.GetUserSub()

	err = h.Keycloak.DeleteUserFromGroup(groupUserSub, kcOldSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddUserToGroup(groupUserSub, newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		UPDATE dbtable_schema.group_users
		SET external_id = $3
		WHERE group_id = $1 AND user_id = $2
	`, session.GroupId, userId, newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/users")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/users/"+userId)

	// Target user will see their roles persisted through cache with this
	h.Redis.Client().Del(req.Context(), groupUserSub+"profile/details")
	h.Redis.DeleteSession(req.Context(), session.UserSub)
	// response := make([]*types.UserRolePair, 1)
	// response[0] = &types.UserRolePair{
	// 	Id:       userId,
	// 	RoleId:   roleId,
	// 	RoleName: roleName,
	// }

	return &types.PatchGroupUserResponse{Success: true}, nil
}

func (h *Handlers) GetGroupUsers(w http.ResponseWriter, req *http.Request, data *types.GetGroupUsersRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.GetGroupUsersResponse, error) {
	var groupUsers []*types.IGroupUser

	err := tx.QueryRows(&groupUsers, `
		SELECT TO_JSONB(eu) as "userProfile", egu.id, r.id as "roleId", r.name as "roleName"
		FROM dbview_schema.enabled_group_users egu
		LEFT JOIN dbview_schema.enabled_users eu ON eu.id = egu."userId"
		JOIN dbtable_schema.group_users gu ON gu.id = egu.id
		JOIN dbtable_schema.group_roles gr ON gr.external_id = gu.external_id
		JOIN dbtable_schema.roles r ON gr.role_id = r.id
		WHERE egu."groupId" = $1
	`, session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUsersResponse{GroupUsers: groupUsers}, nil
}

func (h *Handlers) GetGroupUserById(w http.ResponseWriter, req *http.Request, data *types.GetGroupUserByIdRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.GetGroupUserByIdResponse, error) {
	var groupUsers []*types.IGroupUser

	err := tx.QueryRows(&groupUsers, `
		SELECT
			eu.id,
			egu.groupId,
			egu.userId,
			eu.firstName,
			eu.lastName,
			eu.locked,
			eu.image,
			eu.email,
			er.id as "roleId",
			er.name as "roleName"
		FROM dbview_schema.enabled_group_users egu
		JOIN dbview_schema.enabled_users eu ON eu.id = egu.userId
		JOIN dbtable_schema.group_users gu ON gu.id = egu.id
		JOIN dbtable_schema.group_roles gr ON gr.external_id = gu.external_id
		JOIN dbview_schema.enabled_roles er ON er.id = gr.role_id
		WHERE egu.groupId = $1 and egu.userId = $2
	`, session.GroupId, data.GetUserId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupUsers) == 0 {
		return nil, util.ErrCheck(errors.New("group user not found"))
	}

	return &types.GetGroupUserByIdResponse{GroupUser: groupUsers[0]}, nil
}

func (h *Handlers) DeleteGroupUser(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupUserRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.DeleteGroupUserResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := tx.Exec(`
			DELETE FROM dbtable_schema.group_users
			WHERE group_id = $1 AND user_id = $2
		`, session.GroupId, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.DeleteGroupUserResponse{Success: true}, nil
}

func (h *Handlers) LockGroupUser(w http.ResponseWriter, req *http.Request, data *types.LockGroupUserRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.LockGroupUserResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := tx.Exec(`
			UPDATE dbtable_schema.group_users
			SET locked = true
			WHERE group_id = $1 AND user_id = $2
		`, session.GroupId, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/users")

	return &types.LockGroupUserResponse{Success: true}, nil
}

func (h *Handlers) UnlockGroupUser(w http.ResponseWriter, req *http.Request, data *types.UnlockGroupUserRequest, session *types.UserSession, tx interfaces.IDatabaseTx) (*types.UnlockGroupUserResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := tx.Exec(`
			UPDATE dbtable_schema.group_users
			SET locked = false
			WHERE group_id = $1 AND user_id = $2
		`, session.GroupId, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/users")

	return &types.UnlockGroupUserResponse{Success: true}, nil
}
