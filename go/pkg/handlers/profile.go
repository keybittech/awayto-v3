package handlers

import (
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostUserProfile(info ReqInfo, data *types.PostUserProfileRequest) (*types.PostUserProfileResponse, error) {
	userSub := info.Session.GetUserSub()
	_, err := info.Tx.Exec(info.Ctx, `
		INSERT INTO dbtable_schema.users (sub, username, first_name, last_name, email, image, created_sub, ip_address, timezone)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::uuid, $8, $9)
	`, userSub, data.Username, data.FirstName, data.LastName, data.Email, data.Image, userSub, info.Session.GetAnonIp(), info.Session.GetTimezone())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PostUserProfileResponse{Success: true}, nil
}

func (h *Handlers) PatchUserProfile(info ReqInfo, data *types.PatchUserProfileRequest) (*types.PatchUserProfileResponse, error) {
	userSub := info.Session.GetUserSub()
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.users
		SET first_name = $2, last_name = $3, email = $4, image = $5, updated_sub = $1, updated_on = $6
		WHERE sub = $1
	`, userSub, data.FirstName, data.LastName, data.Email, data.Image, time.Now())

	info.Batch.Send(info.Ctx)

	err := h.Keycloak.UpdateUser(info.Ctx, userSub, userSub, data.FirstName, data.LastName)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, userSub+"profile/details")

	return &types.PatchUserProfileResponse{Success: true}, nil
}

func (h *Handlers) GetUserProfileDetails(info ReqInfo, data *types.GetUserProfileDetailsRequest) (*types.GetUserProfileDetailsResponse, error) {
	userSub := info.Session.GetUserSub()
	groupId := info.Session.GetGroupId()
	groupCode := info.Session.GetGroupCode()

	upReq := util.BatchQueryRow[types.IUserProfile](info.Batch, `
		SELECT "firstName", "lastName",	image, email, locked,	active
		FROM dbview_schema.enabled_users
		WHERE sub = $1
	`, userSub)

	var groupsReq *map[string]*types.IGroup
	var groupRolesReq *map[string]*types.IGroupRole
	var quotesReq *map[string]*types.IQuote
	var bookingsReq *map[string]*types.IBooking

	if groupId != "" {
		groupsReq = util.BatchQueryMap[types.IGroup](info.Batch, "code", `
			SELECT code, name, "defaultRoleId", "displayName", "createdOn", purpose, ai, true as active
			FROM dbview_schema.enabled_groups
			WHERE code = $1
		`, groupCode)

		groupRolesReq = util.BatchQueryMap[types.IGroupRole](info.Batch, "id", `
			SELECT egr.id, egr."roleId", er.name
			FROM dbview_schema.enabled_group_roles egr
			JOIN dbview_schema.enabled_roles er ON er.id = egr."roleId"
			WHERE egr."groupId" = $1 AND er.name != 'Admin'
		`, groupId)

		quotesReq = util.BatchQueryMap[types.IQuote](info.Batch, "id", `
			SELECT id, "slotDate", "startTime", "scheduleBracketSlotId", "serviceTierName", "serviceName", "createdOn"
			FROM dbview_schema.enabled_quotes
			WHERE "slotCreatedSub" = $1
		`, userSub)

		bookingsReq = util.BatchQueryMap[types.IBooking](info.Batch, "id", `
			SELECT id, "slotDate", "scheduleBracketSlot", service, "serviceTier", "createdOn"
			FROM dbview_schema.enabled_bookings eb
			WHERE "createdSub" = $1 OR "quoteCreatedSub" = $1
		`, userSub)
	}

	info.Batch.Send(info.Ctx)

	up := *upReq

	if groupId != "" {
		up.Groups = *groupsReq
		up.Groups[groupCode].Roles = *groupRolesReq
		up.Quotes = *quotesReq
		up.Bookings = *bookingsReq
	}

	up.RoleBits = info.Session.GetRoleBits()
	up.RoleName = info.Session.GetSubGroupName()

	// Try to send a request if the user has an active socket connection
	// but no need to catch errors as they may not yet have a connection
	go h.Socket.RoleCall(userSub)

	return &types.GetUserProfileDetailsResponse{UserProfile: up}, nil
}

func (h *Handlers) DisableUserProfile(info ReqInfo, data *types.DisableUserProfileRequest) (*types.DisableUserProfileResponse, error) {
	userSub := info.Session.GetUserSub()
	util.BatchExec(info.Batch, `
		UPDATE dbtable_schema.users
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE sub = $1
	`, userSub, time.Now(), userSub)

	info.Batch.Send(info.Ctx)

	return &types.DisableUserProfileResponse{Success: true}, nil
}

func (h *Handlers) ActivateProfile(info ReqInfo, data *types.ActivateProfileRequest) (*types.ActivateProfileResponse, error) {
	_, err := info.Tx.Exec(info.Ctx, `
		UPDATE dbtable_schema.users
		SET active = true, updated_on = $2, updated_sub = $1
		WHERE sub = $1
	`, info.Session.GetUserSub(), time.Now())
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
	`, info.Session.GetUserSub(), time.Now())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeactivateProfileResponse{Success: true}, nil
}
