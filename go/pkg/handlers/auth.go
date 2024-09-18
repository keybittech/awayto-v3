package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
)

func (h *Handlers) AuthWebhook_REGISTER(req *http.Request, authEvent clients.AuthEvent) {

	userProfile := &types.PostUserProfileRequest{}

	userProfile.FirstName = authEvent.Details.FirstName
	userProfile.LastName = authEvent.Details.LastName
	userProfile.Username = authEvent.Details.Username
	userProfile.Email = authEvent.Details.Email

	_, err := h.PostUserProfile(nil, req, userProfile)
	if err != nil {
		util.ErrCheck(err)
		return
	}
}
