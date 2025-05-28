package handlers

import (
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"
)

func (h *Handlers) PatchGroupUser(info ReqInfo, data *types.PatchGroupUserRequest) (*types.PatchGroupUserResponse, error) {
	var userId, oldSubgroupExternalId string
	err := info.Tx.QueryRow(info.Ctx, `
		SELECT gu.external_id, gu.user_id
		FROM dbtable_schema.group_users gu
		JOIN dbtable_schema.users u ON u.id = gu.user_id
		WHERE gu.group_id = $1 AND u.sub = $2
	`, info.Session.GetGroupId(), data.UserSub).Scan(&oldSubgroupExternalId, &userId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var newSubgroupExternalId string
	err = info.Tx.QueryRow(info.Ctx, `
		SELECT external_id
		FROM dbtable_schema.group_roles
		WHERE group_id = $1 AND role_id = $2
	`, info.Session.GetGroupId(), data.RoleId).Scan(&newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.DeleteUserFromGroup(info.Ctx, info.Session.GetUserSub(), data.UserSub, oldSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.AddUserToGroup(info.Ctx, info.Session.GetUserSub(), data.UserSub, newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.group_users
		SET external_id = $3
		WHERE group_id = $1 AND user_id = $2
	`, info.Session.GetGroupId(), userId, newSubgroupExternalId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.RefreshSession(info.Req, data.UserSub)

	_ = h.Socket.RoleCall(data.UserSub)

	return &types.PatchGroupUserResponse{Success: true}, nil
}

func (h *Handlers) GetGroupUsers(info ReqInfo, data *types.GetGroupUsersRequest) (*types.GetGroupUsersResponse, error) {
	groupUsers := util.BatchQuery[types.IGroupUser](info.Batch, `
		SELECT TO_JSONB(eu) as "userProfile", egu.id, r.id as "roleId", r.name as "roleName"
		FROM dbview_schema.enabled_group_users egu
		LEFT JOIN dbview_schema.enabled_users eu ON eu.id = egu."userId"
		JOIN dbtable_schema.group_users gu ON gu.id = egu.id
		JOIN dbtable_schema.group_roles gr ON gr.external_id = gu.external_id
		JOIN dbtable_schema.roles r ON gr.role_id = r.id
		WHERE egu."groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupUsersResponse{GroupUsers: *groupUsers}, nil
}

func (h *Handlers) GetGroupUserById(info ReqInfo, data *types.GetGroupUserByIdRequest) (*types.GetGroupUserByIdResponse, error) {
	groupUser := util.BatchQueryRow[types.IGroupUser](info.Batch, `
		SELECT gu.id, egu.groupId, egu.userId, er.id as "roleId", er.name as "roleName"
		FROM dbview_schema.enabled_group_users egu
		JOIN dbtable_schema.group_users gu ON gu.id = egu.id
		JOIN dbtable_schema.group_roles gr ON gr.external_id = gu.external_id
		JOIN dbview_schema.enabled_roles er ON er.id = gr.role_id
		WHERE egu.groupId = $1 and egu.userId = $2
	`, info.Session.GetGroupId(), data.UserId)

	info.Batch.Send(info.Ctx)

	return &types.GetGroupUserByIdResponse{GroupUser: *groupUser}, nil
}

func (h *Handlers) DeleteGroupUser(info ReqInfo, data *types.DeleteGroupUserRequest) (*types.DeleteGroupUserResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	rows, err := info.Tx.Query(info.Ctx, `
		SELECT u.sub as "userSub", gu.external_id as "externalId"
		FROM dbtable_schema.users u
		JOIN dbtable_schema.group_users gu ON gu.user_id = u.id
		WHERE gu.user_id = ANY($1)
	`, pq.Array(ids))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	usersInfo, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[types.IGroupUser])
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = info.Tx.Exec(info.Ctx, `
		DELETE FROM dbtable_schema.group_users
		WHERE group_id = $1 AND user_id = ANY($2)
	`, info.Session.GetGroupId(), pq.Array(ids))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for _, u := range usersInfo {
		err = h.Keycloak.DeleteUserFromGroup(info.Ctx, info.Session.GetUserSub(), u.UserSub, u.ExternalId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	return &types.DeleteGroupUserResponse{Success: true}, nil
}

func (h *Handlers) LockGroupUser(info ReqInfo, data *types.LockGroupUserRequest) (*types.LockGroupUserResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.group_users
		SET locked = true
		WHERE group_id = $1 AND user_id = $2
	`, info.Session.GetGroupId(), pq.Array(strings.Split(data.Ids, ",")))

	info.Batch.Send(info.Ctx)

	return &types.LockGroupUserResponse{Success: true}, nil
}

func (h *Handlers) UnlockGroupUser(info ReqInfo, data *types.UnlockGroupUserRequest) (*types.UnlockGroupUserResponse, error) {
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.group_users
		SET locked = false
		WHERE group_id = $1 AND user_id = $2
	`, info.Session.GetGroupId(), pq.Array(strings.Split(data.Ids, ",")))

	info.Batch.Send(info.Ctx)

	return &types.UnlockGroupUserResponse{Success: true}, nil
}
