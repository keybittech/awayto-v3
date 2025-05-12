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

	info.Batch.Queue(`
		SELECT "firstName", "lastName",	image, email, locked,	active
		FROM dbview_schema.enabled_users
		WHERE sub = $1
	`, info.Session.UserSub)

	// info.Batch.Queue(`
	// 	SELECT name, "displayName", "createdOn", purpose, ai
	// 	FROM dbview_schema.enabled_groups
	// 	WHERE id = $1
	// `, info.Session.GroupId)
	//
	results, err := h.Database.DatabaseClient.SendBatch(info.Ctx, info.Batch)
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	defer results.Close()

	up := &types.IUserProfile{}
	profileRow := results.QueryRow()
	err = profileRow.Scan(&up.FirstName, &up.LastName, &up.Image, &up.Email, &up.Locked, &up.Active)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// g := &types.IGroup{}
	// groupRow := results.QueryRow()
	// err = groupRow.Scan(&g.Name, &g.DisplayName, &g.CreatedOn, &g.Purpose, &g.Ai)
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	//
	// up.Groups = make(map[string]*types.IGroup, 1)
	// up.Groups[info.Session.GroupId] = g

	// println(fmt.Sprint(up))

	// up := &types.IUserProfile{}
	// row := info.Tx.QueryRow(info.Ctx, `
	// 	SELECT "firstName", "lastName",	image, email, locked,	active
	// 	FROM dbview_schema.enabled_users
	// 	WHERE sub = $1
	// `, info.Session.UserSub)
	// err := row.Scan(&up.FirstName, &up.LastName, &up.Image, &up.Email, &up.Locked, &up.Active)
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }

	// println(fmt.Sprint(up))

	//
	// row := info.Tx.QueryRow(info.Ctx, )
	//
	// // Variables to hold JSON data from the database for map fields
	// var groupsJSON, rolesJSON, quotesJSON, bookingsJSON []byte
	//
	// row := info.Tx.QueryRow(info.Ctx, `
	// 	SELECT
	// 		"firstName",
	// 		"lastName",
	// 		sub,
	// 		image,
	// 		email,
	// 		locked,
	// 		active,
	// 		COALESCE(groups::TEXT, '{}') as groups,
	// 		COALESCE(roles::TEXT, '{}') as roles,
	// 		COALESCE(quotes::TEXT, '{}') as quotes,
	// 		COALESCE(bookings::TEXT, '{}') as bookings
	// 	FROM dbview_schema.enabled_users_ext
	// 	WHERE sub = $1
	// `, info.Session.UserSub)
	// err := row.Scan(
	// 	&userProfile.FirstName,
	// 	&userProfile.LastName,
	// 	&userProfile.Sub,
	// 	&userProfile.Image,
	// 	&userProfile.Email,
	// 	&userProfile.Locked,
	// 	&userProfile.Active,
	// 	&groupsJSON,
	// 	&rolesJSON,
	// 	&quotesJSON,
	// 	&bookingsJSON,
	// )
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	//
	// if err := json.Unmarshal(groupsJSON, &userProfile.Groups); err != nil {
	// 	return nil, util.ErrCheck(fmt.Errorf("failed to unmarshal user groupsJSON: %w", err))
	// }
	//
	// if err := json.Unmarshal(rolesJSON, &userProfile.Roles); err != nil {
	// 	return nil, util.ErrCheck(fmt.Errorf("failed to unmarshal user rolesJSON: %w", err))
	// }
	//
	// if err := json.Unmarshal(quotesJSON, &userProfile.Quotes); err != nil {
	// 	return nil, util.ErrCheck(fmt.Errorf("failed to unmarshal user quotesJSON: %w", err))
	// }
	//
	// if err := json.Unmarshal(bookingsJSON, &userProfile.Bookings); err != nil {
	// 	return nil, util.ErrCheck(fmt.Errorf("failed to unmarshal user bookingsJSON: %w", err))
	// }
	//
	// if _, ok := userProfile.Groups[info.Session.GroupId]; ok {
	// 	group := &types.IGroup{}
	// 	var groupRolesJSON []byte
	//
	// 	groupRow := info.Tx.QueryRow(info.Ctx, `
	// 		SELECT
	// 			name,
	// 			"displayName",
	// 			purpose,
	// 			ai,
	// 			code,
	// 			COALESCE("defaultRoleId"::TEXT, '') as "defaultRoleId",
	// 			"allowedDomains",
	// 			true as active,
	// 			COALESCE(roles::TEXT, '{}') as roles
	// 		FROM dbview_schema.enabled_groups_ext
	// 		WHERE id = $1
	// 	`, info.Session.GroupId)
	// 	err := groupRow.Scan(
	// 		&group.Name,
	// 		&group.DisplayName,
	// 		&group.Purpose,
	// 		&group.Ai,
	// 		&group.Code,
	// 		&group.DefaultRoleId,
	// 		&group.AllowedDomains,
	// 		&group.Active,
	// 		&groupRolesJSON,
	// 	)
	// 	if err != nil {
	// 		return nil, util.ErrCheck(err)
	// 	}
	//
	// 	if err := json.Unmarshal(groupRolesJSON, &group.Roles); err != nil {
	// 		return nil, util.ErrCheck(fmt.Errorf("failed to unmarshal user groupRolesJSON: %w", err))
	// 	}
	// }

	// userProfile, err := clients.QueryProto[types.IUserProfile](info.Ctx, info.Tx, `
	// 	SELECT
	// 		"firstName",
	// 		"lastName",
	// 		sub,
	// 		image,
	// 		email,
	// 		locked,
	// 		active,
	// 		groups,
	// 		roles,
	// 		quotes,
	// 		bookings
	// 	FROM dbview_schema.enabled_users_ext
	// 	WHERE sub = $1
	// `, info.Session.UserSub)
	// if err != nil {
	// 	return nil, util.ErrCheck(err)
	// }
	//
	// if _, ok := userProfile.Groups[info.Session.GroupId]; ok {
	// 	userProfile.Groups[info.Session.GroupId], err = clients.QueryProto[types.IGroup](info.Ctx, info.Tx, `
	// 		SELECT
	// 			name,
	// 			"displayName",
	// 			purpose,
	// 			ai,
	// 			code,
	// 			COALESCE("defaultRoleId"::TEXT, '') as "defaultRoleId",
	// 			"allowedDomains",
	// 			roles,
	// 			true as active
	// 		FROM dbview_schema.enabled_groups_ext
	// 		WHERE id = $1
	// 	`, info.Session.GroupId)
	// 	if err != nil {
	// 		return nil, util.ErrCheck(err)
	// 	}
	// }

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
