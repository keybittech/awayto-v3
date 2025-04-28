package handlers

import (
	"errors"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PatchGroupUser(info ReqInfo, data *types.PatchGroupUserRequest) (*types.PatchGroupUserResponse, error) {
	userSub := data.GetUserSub()
	roleId := data.GetRoleId()

	var userId, oldSubgroupExternalId string
	err := info.Tx.QueryRow(info.Req.Context(), `
		SELECT gu.external_id as "externalId", gu.user_id as "userId"
		FROM dbtable_schema.group_users gu
		JOIN dbtable_schema.users u ON u.id = gu.user_id
		JOIN dbtable_schema.groups g ON g.id = gu.group_id
		WHERE g.id = $1 AND u.sub = $2
	`, info.Session.GroupId, userSub).Scan(&oldSubgroupExternalId, &userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var newSubgroupExternalId string
	err = info.Tx.QueryRow(info.Req.Context(), `
		SELECT external_id as "externalId"
		FROM dbtable_schema.group_roles
		WHERE group_id = $1 AND role_id = $2
	`, info.Session.GroupId, roleId).Scan(&newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.DeleteUserFromGroup(info.Session.UserSub, userSub, oldSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddUserToGroup(info.Session.UserSub, userSub, newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.group_users
		SET external_id = $3
		WHERE group_id = $1 AND user_id = $2
	`, info.Session.GroupId, userId, newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"profile/details")
	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/users")
	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/users/"+userId)

	// Target user will see their roles persisted through cache with this
	h.Redis.Client().Del(info.Req.Context(), userSub+"profile/details")
	h.Redis.DeleteSession(info.Req.Context(), info.Session.UserSub)
	// response := make([]*types.UserRolePair, 1)
	// response[0] = &types.UserRolePair{
	// 	Id:       userId,
	// 	RoleId:   roleId,
	// 	RoleName: roleName,
	// }

	return &types.PatchGroupUserResponse{Success: true}, nil
}

func (h *Handlers) GetGroupUsers(info ReqInfo, data *types.GetGroupUsersRequest) (*types.GetGroupUsersResponse, error) {
	var groupUsers []*types.IGroupUser

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &groupUsers, `
		SELECT TO_JSONB(eu) as "userProfile", egu.id, r.id as "roleId", r.name as "roleName"
		FROM dbview_schema.enabled_group_users egu
		LEFT JOIN dbview_schema.enabled_users eu ON eu.id = egu."userId"
		JOIN dbtable_schema.group_users gu ON gu.id = egu.id
		JOIN dbtable_schema.group_roles gr ON gr.external_id = gu.external_id
		JOIN dbtable_schema.roles r ON gr.role_id = r.id
		WHERE egu."groupId" = $1
	`, info.Session.GroupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupUsersResponse{GroupUsers: groupUsers}, nil
}

func (h *Handlers) GetGroupUserById(info ReqInfo, data *types.GetGroupUserByIdRequest) (*types.GetGroupUserByIdResponse, error) {
	var groupUsers []*types.IGroupUser

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &groupUsers, `
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
	`, info.Session.GroupId, data.GetUserId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(groupUsers) == 0 {
		return nil, util.ErrCheck(errors.New("group user not found"))
	}

	return &types.GetGroupUserByIdResponse{GroupUser: groupUsers[0]}, nil
}

func (h *Handlers) DeleteGroupUser(info ReqInfo, data *types.DeleteGroupUserRequest) (*types.DeleteGroupUserResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := info.Tx.Exec(info.Req.Context(), `
			DELETE FROM dbtable_schema.group_users
			WHERE group_id = $1 AND user_id = $2
		`, info.Session.GroupId, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.DeleteGroupUserResponse{Success: true}, nil
}

func (h *Handlers) LockGroupUser(info ReqInfo, data *types.LockGroupUserRequest) (*types.LockGroupUserResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := info.Tx.Exec(info.Req.Context(), `
			UPDATE dbtable_schema.group_users
			SET locked = true
			WHERE group_id = $1 AND user_id = $2
		`, info.Session.GroupId, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/users")

	return &types.LockGroupUserResponse{Success: true}, nil
}

func (h *Handlers) UnlockGroupUser(info ReqInfo, data *types.UnlockGroupUserRequest) (*types.UnlockGroupUserResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := info.Tx.Exec(info.Req.Context(), `
			UPDATE dbtable_schema.group_users
			SET locked = false
			WHERE group_id = $1 AND user_id = $2
		`, info.Session.GroupId, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"group/users")

	return &types.UnlockGroupUserResponse{Success: true}, nil
}
