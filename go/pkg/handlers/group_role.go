package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (h *Handlers) PostGroupRole(w http.ResponseWriter, req *http.Request, data *types.PostGroupRoleRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostGroupRoleResponse, error) {
	kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(session.GroupExternalId, data.GetName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var groupRoleId string
	err = tx.QueryRow(`
		INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (group_id, role_id) DO NOTHING
		RETURNING id
	`, session.GroupId, data.GetRoleId(), kcSubGroup.Id, time.Now().Local().UTC(), session.UserSub).Scan(&groupRoleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// update the default role
	if data.GetDefaultRole() {
		_, err = tx.Exec(`
			UPDATE dbtable_schema.groups
			SET default_role_id = $2
			WHERE id = $1
		`, session.GroupId, data.GetRoleId())
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/roles")
	h.Redis.Client().Del(req.Context(), session.UserSub+"group/assignments")
	h.Redis.SetGroupSessionVersion(req.Context(), session.GroupId)

	return &types.PostGroupRoleResponse{GroupRoleId: groupRoleId}, nil
}

func (h *Handlers) PatchGroupRole(w http.ResponseWriter, req *http.Request, data *types.PatchGroupRoleRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchGroupRoleResponse, error) {
	var existingRoleExternalId, defaultRoleId string
	err := tx.QueryRow(`
		SELECT gr.external_id, g.default_role_id FROM dbtable_schema.group_roles gr
		JOIN dbtable_schema.groups g ON g.id = gr.group_id
		WHERE gr.group_id = $1 AND gr.role_id = $2
	`, session.GroupId, data.GetRoleId()).Scan(&existingRoleExternalId, &defaultRoleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// create new general role or get existing id
	postRoleResponse, err := h.PostRole(w, req, &types.PostRoleRequest{Name: data.GetName()}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// if attempting to change the default role while unsetting it as the default, error out
	if postRoleResponse.GetId() == data.GetRoleId() && defaultRoleId == data.GetRoleId() && data.GetDefaultRole() == false {
		return nil, util.ErrCheck(util.UserError("Update a different role to be the default role, then modify this role."))
	}

	// update the default role
	if data.GetDefaultRole() {
		_, err = tx.Exec(`
			UPDATE dbtable_schema.groups
			SET default_role_id = $2
			WHERE id = $1
		`, session.GroupId, postRoleResponse.GetId())
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	if postRoleResponse.GetId() != data.GetRoleId() {
		// remove the old group role entry and
		_, err := tx.Exec(`
			DELETE FROM dbtable_schema.group_roles
			WHERE group_id = $1 AND role_id = $2
		`, session.GroupId, data.GetRoleId())
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		// create a new group_role attachment pointing to the new role id, retaining the keycloak external id
		_, err = tx.Exec(`
			INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (group_id, role_id) DO NOTHING
		`, session.GroupId, postRoleResponse.GetId(), existingRoleExternalId, time.Now().Local().UTC(), session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		// update the name of the keycloak group which controls this role
		err = h.Keycloak.UpdateGroup(existingRoleExternalId, data.GetName())
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")
		h.Redis.Client().Del(req.Context(), session.UserSub+"group/roles")
		h.Redis.Client().Del(req.Context(), session.UserSub+"group/assignments")
		h.Redis.SetGroupSessionVersion(req.Context(), session.GroupId)
	}

	return &types.PatchGroupRoleResponse{Success: true}, nil
}

func (h *Handlers) PatchGroupRoles(w http.ResponseWriter, req *http.Request, data *types.PatchGroupRolesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchGroupRolesResponse, error) {
	_, err := tx.Exec(`
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

	// Remove all unused roles from keycloak and the group records
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
			return nil, util.ErrCheck(err)
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
						return nil, util.ErrCheck(err)
					}
				}
			}
		}

		_, err = tx.Exec(`
			DELETE FROM dbtable_schema.group_roles
			WHERE id = ANY($1::uuid[])
		`, pq.Array(diffIds))

		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	// Add new roles to keycloak and group records
	for _, roleId := range roleIds {
		role := data.GetRoles()[roleId]

		if strings.ToLower(role.Name) == "admin" {
			continue
		}

		kcSubGroup, err := h.Keycloak.CreateOrGetSubGroup(session.GroupExternalId, role.Name)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		_, err = tx.Exec(`
			INSERT INTO dbtable_schema.group_roles (group_id, role_id, external_id, created_on, created_sub)
			VALUES ($1, $2, $3, $4, $5::uuid)
			ON CONFLICT (group_id, role_id) DO NOTHING
		`, session.GroupId, roleId, kcSubGroup.Id, time.Now().Local().UTC(), session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	return &types.PatchGroupRolesResponse{Success: true}, nil
}

func (h *Handlers) GetGroupRoles(w http.ResponseWriter, req *http.Request, data *types.GetGroupRolesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetGroupRolesResponse, error) {
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

func (h *Handlers) DeleteGroupRole(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupRoleRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteGroupRoleResponse, error) {
	groupRoleIds := strings.Split(data.GetIds(), ",")

	// handle default role id update
	var defaultGroupRoleId string
	err := tx.QueryRow(`
		SELECT gr.id FROM dbtable_schema.group_roles gr
		JOIN dbtable_schema.groups g ON g.default_role_id = gr.role_id
		WHERE g.id = $1
	`, session.GroupId).Scan(&defaultGroupRoleId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	println(fmt.Sprintf("%s %s", defaultGroupRoleId, strings.Join(groupRoleIds, ",")))

	if slices.Contains(groupRoleIds, defaultGroupRoleId) {
		return nil, util.ErrCheck(util.UserError("The default role may not be deleted. Update a different role to be default, then try again."))
	}

	for _, id := range groupRoleIds {
		if id != "" {
			var subGroupExternalId string
			err := tx.QueryRow(`
				DELETE FROM dbtable_schema.group_roles
				WHERE id = $1 AND group_id = $2
				RETURNING external_id
			`, id, session.GroupId).Scan(&subGroupExternalId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}

			err = h.Keycloak.DeleteGroup(subGroupExternalId)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/roles")

	return &types.DeleteGroupRoleResponse{Success: true}, nil
}
