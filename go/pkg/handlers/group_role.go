package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (h *Handlers) PostGroupRole(w http.ResponseWriter, req *http.Request, data *types.PostGroupRoleRequest) (*types.PostGroupRoleResponse, error) {
	session := h.Redis.ReqSession(req)

	kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(session.GroupExternalId, data.GetRole().GetName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var groupRoleId string
	err = h.Database.Client().QueryRow(`
		INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (group_id, role_id) DO NOTHING
		RETURNING id
	`, session.GroupId, data.GetRole().GetId(), kcSubGroup.Id, time.Now().Local().UTC(), session.UserSub).Scan(&groupRoleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostGroupRoleResponse{Id: groupRoleId}, nil
}

func (h *Handlers) PatchGroupRoles(w http.ResponseWriter, req *http.Request, data *types.PatchGroupRolesRequest) (*types.PatchGroupRolesResponse, error) {
	session := h.Redis.ReqSession(req)

	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.groups
		SET default_role_id = $2, updated_sub = $3, updated_on = $4 
		WHERE id = $1
	`, session.GroupId, data.GetDefaultRoleId(), session.UserSub, time.Now().Local().UTC())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	roleIds := make([]string, 0, len(data.GetRoles()))
	for roleId := range data.GetRoles() {
		roleIds = append(roleIds, roleId)
	}

	var diffs []*types.IGroupRole

	rows, err := h.Database.Client().Query(`
		SELECT egr.id, egr."roleId", er.name
		FROM dbview_schema.enabled_group_roles egr
		JOIN dbview_schema.enabled_roles er ON er.id = egr."roleId"
		WHERE er.name != 'Admin' AND egr."groupId" = $1 AND egr."roleId" != ANY($2::uuid[])
	`, session.GroupId, pq.Array(roleIds))
	defer rows.Close()

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	for rows.Next() {
		var groupRoleId, roleId, roleName string
		rows.Scan(&groupRoleId, &roleId, &roleName)
		diffs = append(diffs, &types.IGroupRole{
			Id: groupRoleId,
			Role: &types.IRole{
				Id:   roleId,
				Name: roleName,
			},
		})
	}

	if len(diffs) > 0 {

		kcGroup, err := h.Keycloak.GetGroup(session.GroupExternalId)
		if err != nil {
			return nil, err
		}

		diffIds := make([]string, 0, len(diffs))
		diffRoleNames := make([]string, 0, len(diffs))
		for _, diff := range diffs {
			diffIds = append(diffIds, diff.GetId())
			diffRoleNames = append(diffRoleNames, diff.GetRole().GetName())
		}

		if len(kcGroup.SubGroups) > 0 {
			for _, subGroup := range kcGroup.SubGroups {
				if util.StringIn(subGroup.Name, diffRoleNames) {
					err = h.Keycloak.DeleteGroup(subGroup.Id)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		_, err = h.Database.Client().Exec(`
			DELETE FROM dbtable_schema.group_roles
			WHERE id = ANY($1::uuid[])
		`, pq.Array(diffIds))

		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	for _, roleId := range roleIds {
		role := data.GetRoles()[roleId]

		if strings.ToLower(role.Name) == "admin" {
			continue
		}

		kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(session.GroupExternalId, role.Name)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		h.Database.Client().Exec(`
			INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4, $5::uuid)
			ON CONFLICT (group_id, role_id) DO NOTHING
		`, session.GroupId, roleId, kcSubGroup.Id, time.Now().Local().UTC(), session.UserSub)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	return &types.PatchGroupRolesResponse{Success: true}, nil
}

func (h *Handlers) GetGroupRoles(w http.ResponseWriter, req *http.Request, data *types.GetGroupRolesRequest) (*types.GetGroupRolesResponse, error) {
	session := h.Redis.ReqSession(req)

	var roles []*types.IGroupRole
	err := h.Database.QueryRows(&roles, `
		SELECT 
			TO_JSONB(er) as role,
			egr.id,
			egr."roleId",
			egr."groupId",
			egr."createdOn"
		FROM dbview_schema.enabled_group_roles egr
		JOIN dbview_schema.enabled_roles er ON er.id = egr."roleId"
		WHERE egr."groupId" = $1 AND er.name != 'Admin'
	`, session.GroupId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupRolesResponse{GroupRoles: roles}, nil
}

func (h *Handlers) DeleteGroupRole(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupRoleRequest) (*types.DeleteGroupRoleResponse, error) {
	session := h.Redis.ReqSession(req)

	for _, id := range strings.Split(data.GetIds(), ",") {
		if id != "" {
			var subGroupExternalId string
			err := h.Database.Client().QueryRow(`
				DELETE FROM dbtable_schema.group_roles
				WHERE id = $1 AND group_id = $2
				RETURNING external_id
			`, id, session.GroupId).Scan(&subGroupExternalId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			h.Keycloak.DeleteGroup(subGroupExternalId)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/roles")

	return &types.DeleteGroupRoleResponse{Success: true}, nil
}
