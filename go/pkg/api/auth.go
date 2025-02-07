package api

import (
	"av3api/pkg/clients"
	"av3api/pkg/util"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
)

var authTransport = http.DefaultTransport

func (a *API) InitAuthProxy(mux *http.ServeMux) {

	kcInternal := os.Getenv("KC_INTERNAL")

	mux.Handle("/auth/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		proxyPath := fmt.Sprintf("%s%s", kcInternal, strings.TrimPrefix(r.URL.Path, "/auth"))
		proxyURL, err := url.Parse(proxyPath)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			http.Error(w, "bad request", http.StatusInternalServerError)
			return
		}
		proxyURL.RawQuery = r.URL.RawQuery

		proxyReq, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
			return
		}

		// Copy the headers from the original request to the proxy request
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		// Use remoteAddr for keycloak rate limiting
		proxyReq.Header.Add("X-Forwarded-For", r.RemoteAddr)

		// Send the proxy request using the custom transport
		resp, err := authTransport.RoundTrip(proxyReq)
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
			http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy the headers from the proxy response to the original response
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Set the status code of the original response to the status code of the proxy response
		w.WriteHeader(resp.StatusCode)

		// Copy the body of the proxy response to the original response
		io.Copy(w, resp.Body)

	}))

}

func (a *API) GetAuthorizedSession(req *http.Request) (*clients.UserSession, error) {
	ctx := req.Context()

	token, ok := req.Header["Authorization"]
	if !ok {
		return nil, errors.New("no auth token")
	}

	_, err := a.Handlers.Keycloak.GetUserInfoByToken(token[0])
	if err != nil {
		return nil, err
	}

	userToken, err := clients.ParseJWT(token[0])
	if err != nil {
		return nil, err
	}

	buildSession := true

	gidSelect := req.Header.Get("X-Gid-Select")

	session, err := a.Handlers.Redis.ReqSession(req)
	if err != nil {
		return nil, err
	}

	if gidSelect == "" {
		groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(ctx, session.GroupId)
		if err != nil {
			return nil, err
		}

		if session.GroupSessionVersion == groupVersion && session.GroupId != "" {
			buildSession = false
		} else {
			session.GroupSessionVersion = groupVersion
		}
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

		err = a.Handlers.Database.Client().QueryRow(`
					SELECT g.id, g.ai, u.sub
					FROM dbtable_schema.groups g
					JOIN dbtable_schema.users u ON u.username = CONCAT('system_group_', g.id)
					WHERE g.external_id = $1
				`, session.GroupExternalId).Scan(&session.GroupId, &session.GroupAi, &session.GroupSub)
		if err != nil {
			return nil, err
		}

		groupVersion, err := a.Handlers.Redis.GetGroupSessionVersion(ctx, session.GroupId)
		if err != nil {
			return nil, err
		} else {
			session.GroupSessionVersion = groupVersion
		}

		a.Handlers.Redis.SetSession(ctx, userToken.Sub, session)
	}

	return session, nil
}
