package handlers

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostUserProfile(info ReqInfo, data *types.PostUserProfileRequest) (*types.PostUserProfileResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		INSERT INTO dbtable_schema.users (sub, username, first_name, last_name, email, image, created_on, created_sub, ip_address, timezone)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::uuid, $9, $10)
	`, info.Session.UserSub, data.GetUsername(), data.GetFirstName(), data.GetLastName(), data.GetEmail(), data.GetImage(), time.Now().Local().UTC(), info.Session.UserSub, info.Session.AnonIp, info.Session.Timezone)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostUserProfileResponse{Success: true}, nil
}

func (h *Handlers) PatchUserProfile(info ReqInfo, data *types.PatchUserProfileRequest) (*types.PatchUserProfileResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.users
		SET first_name = $2, last_name = $3, email = $4, image = $5, updated_sub = $1, updated_on = $6
		WHERE sub = $1
	`, info.Session.UserSub, data.GetFirstName(), data.GetLastName(), data.GetEmail(), data.GetImage(), time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.UpdateUser(info.Session.UserSub, info.Session.UserSub, data.GetFirstName(), data.GetLastName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Req.Context(), info.Session.UserSub+"profile/details")

	return &types.PatchUserProfileResponse{Success: true}, nil
}

func (h *Handlers) GetUserProfileDetails(info ReqInfo, data *types.GetUserProfileDetailsRequest) (*types.GetUserProfileDetailsResponse, error) {
	profileRows, err := info.Tx.Query(info.Req.Context(), `
		SELECT 
			"firstName",
			"lastName",
			sub,
			image,
			email,
			locked,
			active,
			groups,
			roles,
			quotes,
			bookings
		FROM dbview_schema.enabled_users_ext
		WHERE sub = $1
	`, info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer profileRows.Close()

	userProfile, err := pgx.CollectOneRow(profileRows, pgx.RowToAddrOfStructByNameLax[types.IUserProfile])
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if _, ok := userProfile.Groups[info.Session.GroupId]; ok {
		groupRows, err := info.Tx.Query(info.Req.Context(), `
			SELECT name, "displayName", purpose, ai, code, "defaultRoleId", "allowedDomains", roles, true as active
			FROM dbview_schema.enabled_groups_ext
			WHERE id = $1
		`, info.Session.GroupId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
		defer groupRows.Close()

		userProfile.Groups[info.Session.GroupId], err = pgx.CollectOneRow(groupRows, pgx.RowToAddrOfStructByNameLax[types.IGroup])
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	roleBits := info.Session.RoleBits
	userProfile.RoleBits = roleBits

	roleName := info.Session.RoleName
	userProfile.RoleName = roleName

	// Try to send a request if the user has an active socket connection
	// but no need to catch errors as they may not yet have a connection
	go h.Socket.RoleCall(info.Session.UserSub)

	return &types.GetUserProfileDetailsResponse{UserProfile: userProfile}, nil
}

func (h *Handlers) GetUserProfileDetailsBySub(info ReqInfo, data *types.GetUserProfileDetailsBySubRequest) (*types.GetUserProfileDetailsBySubResponse, error) {

	var userProfiles []*types.IUserProfile

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &userProfiles, `
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

func (h *Handlers) GetUserProfileDetailsById(info ReqInfo, data *types.GetUserProfileDetailsByIdRequest) (*types.GetUserProfileDetailsByIdResponse, error) {
	var userProfiles []*types.IUserProfile

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &userProfiles, `
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

func (h *Handlers) DisableUserProfile(info ReqInfo, data *types.DisableUserProfileRequest) (*types.DisableUserProfileResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.users
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableUserProfileResponse{Success: true}, nil
}

func (h *Handlers) ActivateProfile(info ReqInfo, data *types.ActivateProfileRequest) (*types.ActivateProfileResponse, error) {

	_, err := info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.users
		SET active = true, updated_on = $2, updated_sub = $1
		WHERE sub = $1
	`, info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.DeleteSession(info.Req.Context(), info.Session.UserSub)

	return &types.ActivateProfileResponse{Success: true}, nil
}

func (h *Handlers) DeactivateProfile(info ReqInfo, data *types.DeactivateProfileRequest) (*types.DeactivateProfileResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.users
		SET active = false, updated_on = $2, updated_sub = $1
		WHERE sub = $1
	`, info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeactivateProfileResponse{Success: true}, nil
}
