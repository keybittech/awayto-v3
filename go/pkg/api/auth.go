package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/util"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"slices"
	"strings"
)

func SetForwardingHeadersAndServe(prox *httputil.ReverseProxy, w http.ResponseWriter, r *http.Request) {
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	r.Header.Add("X-Forwarded-Proto", "https")
	r.Header.Add("X-Forwarded-Host", r.Host)
	prox.ServeHTTP(w, r)
}

func (a *API) InitAuthProxy(mux *http.ServeMux) {

	kcRealm := os.Getenv("KC_REALM")
	kcInternal, err := url.Parse(os.Getenv("KC_INTERNAL"))
	if err != nil {
		log.Fatal("invalid keycloak url")
	}

	authProxy := httputil.NewSingleHostReverseProxy(kcInternal)

	userRoutes := []string{
		"login-actions/registration",
		"login-actions/authenticate",
		"login-actions/reset-credentials",
		"protocol/openid-connect/3p-cookies/step1.html",
		"protocol/openid-connect/3p-cookies/step2.html",
		"protocol/openid-connect/login-status-iframe.html",
		"protocol/openid-connect/login-status-iframe.html/init",
		"protocol/openid-connect/registrations",
		"protocol/openid-connect/auth",
		"protocol/openid-connect/token",
		"protocol/openid-connect/logout",
	}

	for _, ur := range userRoutes {
		authRoute := fmt.Sprintf("/auth/realms/%s/%s", kcRealm, ur)
		mux.Handle(authRoute, http.StripPrefix("/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SetForwardingHeadersAndServe(authProxy, w, r)
		})))
	}

	mux.Handle("/auth/resources/", http.StripPrefix("/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetForwardingHeadersAndServe(authProxy, w, r)
	})))
}

func (a *API) GetAuthorizedSession(req *http.Request, tx clients.IDatabaseTx) (*clients.UserSession, error) {
	ctx := req.Context()

	token, ok := req.Header["Authorization"]
	if !ok {
		return nil, errors.New("no auth token")
	}

	valid, err := a.Handlers.Keycloak.GetUserTokenValid(token[0])
	if !valid || err != nil {
		return nil, util.ErrCheck(errors.New(err.Error() + fmt.Sprintf(" Validity check: %t", valid)))
	}

	userToken, _, err := clients.ParseJWT(token[0])
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	buildSession := true

	gidSelect := req.Header.Get("X-Gid-Select")

	session, err := a.Handlers.Redis.ReqSession(req)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if gidSelect == "" {
		groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(ctx, session.GroupId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		if session.GroupSessionVersion == groupVersion && session.GroupId != "" {
			buildSession = false
		} else {
			session.GroupSessionVersion = groupVersion
		}
	}

	err = tx.SetDbVar("user_sub", session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	groupId := session.GroupExternalId
	if session.GroupId != "" {
		groupId = session.GroupId
	}

	err = tx.SetDbVar("group_id", groupId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if buildSession && len(session.SubGroups) > 0 {

		if gidSelect != "" && slices.Contains(session.SubGroups, gidSelect) {
			session.SubGroupName = gidSelect
		} else {
			session.SubGroupName = fmt.Sprint(session.SubGroups[0])
		}

		names := strings.Split(session.SubGroupName, "/")
		session.GroupName = names[1]
		session.RoleName = names[2]
		kcGroupName := "/" + session.GroupName

		kcGroups, err := a.Handlers.Keycloak.GetGroupByName(kcGroupName)

		for _, gr := range *kcGroups {
			if gr.Path == kcGroupName {
				session.GroupExternalId = gr.Id
				break
			}
		}

		kcSubgroups, err := a.Handlers.Keycloak.GetGroupSubgroups(session.GroupExternalId)

		for _, gr := range *kcSubgroups {
			if gr.Path == session.SubGroupName {
				session.SubGroupExternalId = gr.Id
				break
			}
		}

		if session.GroupExternalId == "" || session.SubGroupExternalId == "" {
			return nil, errors.New("could not describe group or subgroup external ids")
		}

		sgRoles := a.Handlers.Keycloak.GetGroupSiteRoles(session.SubGroupExternalId)
		for _, mapping := range sgRoles {
			session.AvailableUserGroupRoles = append(session.AvailableUserGroupRoles, mapping.Name)
		}

		err = tx.QueryRow(`
			SELECT id, ai, sub
			FROM dbtable_schema.groups
			WHERE external_id = $1
		`, session.GroupExternalId).Scan(&session.GroupId, &session.GroupAi, &session.GroupSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		err = tx.SetDbVar("group_id", session.GroupId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(ctx, session.GroupId)
		if err != nil {
			return nil, util.ErrCheck(err)
		} else {
			session.GroupSessionVersion = groupVersion
		}

		err = tx.QueryRow(`SELECT timezone FROM dbtable_schema.users WHERE sub = $1`, session.UserSub).Scan(&session.Timezone)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		a.Handlers.Redis.SetSession(ctx, userToken.Sub, session)
	}

	return session, nil
}
