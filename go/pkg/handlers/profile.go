package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"errors"
	"net/http"
	"time"
)

func (h *Handlers) PostUserProfile(w http.ResponseWriter, req *http.Request, data *types.PostUserProfileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostUserProfileResponse, error) {
	_, err := tx.Exec(`
		INSERT INTO dbtable_schema.users (sub, username, first_name, last_name, email, image, created_on, created_sub, ip_address, timezone)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::uuid, $9, $10)
	`, session.UserSub, data.GetUsername(), data.GetFirstName(), data.GetLastName(), data.GetEmail(), data.GetImage(), time.Now().Local().UTC(), session.UserSub, session.AnonIp, session.Timezone)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostUserProfileResponse{Success: true}, nil
}

func (h *Handlers) PatchUserProfile(w http.ResponseWriter, req *http.Request, data *types.PatchUserProfileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchUserProfileResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.users
		SET first_name = $2, last_name = $3, email = $4, image = $5, updated_sub = $1, updated_on = $6
		WHERE sub = $1
	`, session.UserSub, data.GetFirstName(), data.GetLastName(), data.GetEmail(), data.GetImage(), time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.UpdateUser(session.UserSub, data.GetFirstName(), data.GetLastName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"profile/details")

	return &types.PatchUserProfileResponse{Success: true}, nil
}

func (h *Handlers) GetUserProfileDetails(w http.ResponseWriter, req *http.Request, data *types.GetUserProfileDetailsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetUserProfileDetailsResponse, error) {
	var userProfiles []*types.IUserProfile

	err := tx.QueryRows(&userProfiles, `
		SELECT * 
		FROM dbview_schema.enabled_users_ext
		WHERE sub = $1
	`, session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(userProfiles) == 0 {
		return nil, util.ErrCheck(errors.New("user not found"))
	}

	userProfile := userProfiles[0]
	extendedGroups := make(map[string]*types.IGroup)

	for _, group := range userProfile.Groups {
		var groups []*types.IGroup

		err := tx.QueryRows(&groups, `
			SELECT *
			FROM dbview_schema.enabled_groups_ext
			WHERE id = $1
		`, group.GetId())

		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if len(groups) == 0 {
			continue
		}

		groups[0].Ldr = group.Ldr
		groups[0].Active = groups[0].Id == session.GroupId
		extendedGroups[group.GetId()] = groups[0]
	}

	if len(extendedGroups) > 0 {
		userProfile.Groups = extendedGroups
	}

	userProfile.AvailableUserGroupRoles = session.AvailableUserGroupRoles

	if err := h.Socket.RoleCall(session.UserSub); err != nil {
		return nil, util.ErrCheck(err)
	}

	userProfile.RoleName = session.RoleName

	return &types.GetUserProfileDetailsResponse{UserProfile: userProfile}, nil
}

func (h *Handlers) GetUserProfileDetailsBySub(w http.ResponseWriter, req *http.Request, data *types.GetUserProfileDetailsBySubRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetUserProfileDetailsBySubResponse, error) {

	var userProfiles []*types.IUserProfile

	err := tx.QueryRows(&userProfiles, `
    SELECT * FROM dbview_schema.enabled_users
    WHERE sub = $1 
	`, data.GetSub())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(userProfiles) == 0 {
		return nil, util.ErrCheck(errors.New("user not found"))
	}

	userProfile := userProfiles[0]

	return &types.GetUserProfileDetailsBySubResponse{UserProfile: userProfile}, nil
}

func (h *Handlers) GetUserProfileDetailsById(w http.ResponseWriter, req *http.Request, data *types.GetUserProfileDetailsByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetUserProfileDetailsByIdResponse, error) {
	var userProfiles []*types.IUserProfile

	err := tx.QueryRows(&userProfiles, `
    SELECT * FROM dbview_schema.enabled_users
    WHERE id = $1 
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(userProfiles) == 0 {
		return nil, util.ErrCheck(errors.New("user not found"))
	}

	userProfile := userProfiles[0]

	return &types.GetUserProfileDetailsByIdResponse{UserProfile: userProfile}, nil
}

func (h *Handlers) DisableUserProfile(w http.ResponseWriter, req *http.Request, data *types.DisableUserProfileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableUserProfileResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.users
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableUserProfileResponse{Success: true}, nil
}

func (h *Handlers) ActivateProfile(w http.ResponseWriter, req *http.Request, data *types.ActivateProfileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.ActivateProfileResponse, error) {
	err := tx.SetDbVar("user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		UPDATE dbtable_schema.users
		SET active = true, updated_on = $2, updated_sub = $1
		WHERE sub = $1
	`, session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.DeleteSession(req.Context(), session.UserSub)

	return &types.ActivateProfileResponse{Success: true}, nil
}

func (h *Handlers) DeactivateProfile(w http.ResponseWriter, req *http.Request, data *types.DeactivateProfileRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeactivateProfileResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.users
		SET active = false, updated_on = $2, updated_sub = $1
		WHERE sub = $1
	`, session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeactivateProfileResponse{Success: true}, nil
}
