package handlers

import (
	"errors"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostUserProfile(info ReqInfo, data *types.PostUserProfileRequest) (*types.PostUserProfileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.users (sub, username, first_name, last_name, email, image, created_on, created_sub, ip_address, timezone)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::uuid, $9, $10)
	`, info.Session.UserSub, data.GetUsername(), data.GetFirstName(), data.GetLastName(), data.GetEmail(), data.GetImage(), time.Now().Local().UTC(), info.Session.UserSub, info.Session.AnonIp, info.Session.Timezone)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostUserProfileResponse{Success: true}, nil
}

func (h *Handlers) PatchUserProfile(info ReqInfo, data *types.PatchUserProfileRequest) (*types.PatchUserProfileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.users
		SET first_name = $2, last_name = $3, email = $4, image = $5, updated_sub = $1, updated_on = $6
		WHERE sub = $1
	`, info.Session.UserSub, data.GetFirstName(), data.GetLastName(), data.GetEmail(), data.GetImage(), time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	err = h.Keycloak.UpdateUser(info.Ctx, info.Session.UserSub, info.Session.UserSub, data.GetFirstName(), data.GetLastName())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.UserSub+"profile/details")

	return &types.PatchUserProfileResponse{Success: true}, nil
}

func (h *Handlers) GetUserProfileDetails(info ReqInfo, data *types.GetUserProfileDetailsRequest) (*types.GetUserProfileDetailsResponse, error) {
	upReq := util.BatchQueryRow[types.IUserProfile](info.Batch, `
		SELECT "firstName", "lastName",	image, email, locked,	active
		FROM dbview_schema.enabled_users
		WHERE sub = $1
	`, info.Session.UserSub)

	groupsReq := util.BatchQueryMap[types.IGroup](info.Batch, "id", `
		SELECT id, name, "displayName", "createdOn", purpose, ai
		FROM dbview_schema.enabled_groups
		WHERE id = $1
	`, info.Session.GroupId)

	rolesReq := util.BatchQueryMap[types.IRole](info.Batch, "id", `
		SELECT er.id, er.name, er."createdOn"
		FROM dbview_schema.enabled_user_roles eur
		JOIN dbview_schema.enabled_roles er ON er.id = eur."roleId"
		WHERE eur.sub = $1
	`, info.Session.UserSub)

	quotesReq := util.BatchQueryMap[types.IQuote](info.Batch, "id", `
		SELECT id, "slotDate", "startTime", "scheduleBracketSlotId", "serviceTierName", "serviceName", "createdOn"
		FROM dbview_schema.enabled_quotes
		WHERE "slotCreatedSub" = $1
	`, info.Session.UserSub)

	bookingsReq := util.BatchQueryMap[types.IBooking](info.Batch, "id", `
		SELECT id, "slotDate", "scheduleBracketSlot", service, "serviceTier", "createdOn"
		FROM dbview_schema.enabled_bookings eb
		WHERE "createdSub" = $1 OR "quoteCreatedSub" = $1
	`, info.Session.UserSub)

	err := info.Batch.Send(info.Ctx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	up := *upReq
	up.Groups = *groupsReq
	up.Roles = *rolesReq
	up.Quotes = *quotesReq
	up.Bookings = *bookingsReq

	roleBits := info.Session.RoleBits
	up.RoleBits = roleBits

	roleName := info.Session.RoleName
	up.RoleName = roleName

	// Try to send a request if the user has an active socket connection
	// but no need to catch errors as they may not yet have a connection
	go h.Socket.RoleCall(info.Session.UserSub)

	return &types.GetUserProfileDetailsResponse{UserProfile: up}, nil
}

func (h *Handlers) GetUserProfileDetailsBySub(info ReqInfo, data *types.GetUserProfileDetailsBySubRequest) (*types.GetUserProfileDetailsBySubResponse, error) {

	var userProfiles []*types.IUserProfile

	err := h.Database.QueryRows(info.Ctx, info.Tx, &userProfiles, `
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

	err := h.Database.QueryRows(info.Ctx, info.Tx, &userProfiles, `
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
	_, err := info.Tx.Exec(info.Ctx, `
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
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.users
		SET active = true, updated_on = $2, updated_sub = $1
		WHERE sub = $1
	`, info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.ActivateProfileResponse{Success: true}, nil
}

func (h *Handlers) DeactivateProfile(info ReqInfo, data *types.DeactivateProfileRequest) (*types.DeactivateProfileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.users
		SET active = false, updated_on = $2, updated_sub = $1
		WHERE sub = $1
	`, info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeactivateProfileResponse{Success: true}, nil
}
