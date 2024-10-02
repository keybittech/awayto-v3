package clients

// https://github.com/PhilippHeuer/go-keycloak with mods

import (
	// Error Handling
	"av3api/pkg/util"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	// Encoding

	"encoding/json"
)

const (
	APP_ROLE_CALL       = "APP_ROLE_CALL"
	APP_GROUP_ADMIN     = "APP_GROUP_ADMIN"
	APP_GROUP_ROLES     = "APP_GROUP_ROLES"
	APP_GROUP_USERS     = "APP_GROUP_USERS"
	APP_GROUP_MATRIX    = "APP_GROUP_MATRIX"
	APP_GROUP_SERVICES  = "APP_GROUP_SERVICES"
	APP_GROUP_BOOKINGS  = "APP_GROUP_BOOKINGS"
	APP_GROUP_FEATURES  = "APP_GROUP_FEATURES"
	APP_GROUP_SCHEDULES = "APP_GROUP_SCHEDULES"
)

type OIDCToken struct {
	AccessToken      string  `json:"access_token"`
	ExpiresIn        float64 `json:"expires_in"`
	RefreshExpiresIn float64 `json:"refresh_expires_in"`
	RefreshToken     string  `json:"refresh_token"`
	TokenType        string  `json:"token_type"`
}

type KeycloakClient struct {
	Server          string
	Realm           string
	AppClient       *KeycloakRealmClient
	ApiClient       *KeycloakRealmClient
	RoleCall        KeycloakRole
	GroupAdminRoles []KeycloakRole
	Token           *OIDCToken
}

type KeycloakUser struct {
	Id                string `json:"id,omitempty"`
	Sub               string `json:"sub,omitempty"`
	CreatedTimestamp  int64  `json:"createdTimestamp,omitempty"`
	Username          string `json:"username,omitempty"`
	Enabled           bool   `json:"enabled,omitempty"`
	Totp              bool   `json:"totp,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	EmailVerified     bool   `json:"emailVerified,omitempty"`
	FirstName         string `json:"firstName,omitempty"`
	LastName          string `json:"lastName,omitempty"`
	Email             string `json:"email,omitempty"`
	FederationLink    string `json:"federationLink,omitempty"`
	Attributes        *struct {
		LDAPENTRYDN []string `json:"LDAP_ENTRY_DN,omitempty"`
		LDAPID      []string `json:"LDAP_ID,omitempty"`
	} `json:"attributes,omitempty"`
	DisableableCredentialTypes []interface{} `json:"disableableCredentialTypes,omitempty"`
	RequiredActions            []interface{} `json:"requiredActions,omitempty"`
	Access                     *struct {
		ManageGroupMembership bool `json:"manageGroupMembership,omitempty"`
		View                  bool `json:"view,omitempty"`
		MapRoles              bool `json:"mapRoles,omitempty"`
		Impersonate           bool `json:"impersonate,omitempty"`
		Manage                bool `json:"manage,omitempty"`
	} `json:"access,omitempty"`
}

type KeycloakUserGroup struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type KeycloakGroup struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	ParentId  string          `json:"parentId"`
	SubGroups []KeycloakGroup `json:"subGroups"`
}

type KeycloakRole struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	ScopeParamRequired bool   `json:"scopeParamRequired"`
	Composite          bool   `json:"composite"`
	ClientRole         bool   `json:"clientRole"`
	ContainerID        string `json:"containerId"`
	Description        string `json:"description,omitempty"`
}

type ClientRoleMapping struct {
	ID       string                  `json:"id"`
	Client   string                  `json:"client"`
	Mappings []ClientRoleMappingRole `json:"mappings"`
}

type ClientRoleMappingRole struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	ScopeParamRequired bool   `json:"scopeParamRequired"`
	Composite          bool   `json:"composite"`
	ClientRole         bool   `json:"clientRole"`
	ContainerID        string `json:"containerId"`
}

type GroupRoleMappingsResponse struct {
	ClientMappings map[string]ClientRoleMapping `json:"clientMappings"`
}

type KeycloakRealmClient struct {
	Id       string `json:"id"`
	ClientID string `json:"clientId"`
}

type KeycloakUserSession struct {
	Id            string            `json:"id"`
	Clients       map[string]string `json:""`
	IpAddress     string            `json:"ipAddress"`
	LastAccess    int               `json:"lastAccess"`
	Start         int               `json:"start"`
	UserId        string            `json:"userId"`
	Username      string            `json:"username"`
	TransientUser bool              `json:"transientUser"`
}

func (keycloakClient KeycloakClient) BasicHeaders() http.Header {
	headers := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + keycloakClient.Token.AccessToken},
	}
	return headers
}

func (keycloakClient KeycloakClient) DirectGrantAuthentication() (*OIDCToken, error) {

	headers := http.Header{}

	body := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {os.Getenv("KC_API_CLIENT")},
		"client_secret": {os.Getenv("KC_API_CLIENT_SECRET")},
	}

	resp, err := util.PostFormData(
		keycloakClient.Server+"/realms/"+keycloakClient.Realm+"/protocol/openid-connect/token",
		headers,
		body,
	)

	if err != nil {
		return nil, err
	}

	var result OIDCToken
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ExpiresIn > 0 {
		return &result, nil
	}

	return nil, util.ErrCheck(errors.New("Authentication failed"))
}

func (keycloakClient KeycloakClient) FindResource(resource, search string) ([]byte, error) {

	queryParams := url.Values{}
	queryParams.Add("first", "0")
	queryParams.Add("max", "11")
	queryParams.Add("search", search)

	resp, err := util.GetWithParams(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/"+resource,
		keycloakClient.BasicHeaders(),
		queryParams,
	)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (keycloakClient KeycloakClient) GetUserListInRealm() (*[]KeycloakUser, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, err
	}

	var result []KeycloakUser
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) GetUserInfoByAuthorization(token string) (*KeycloakUser, error) {

	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		keycloakClient.Server+"/realms/"+keycloakClient.Realm+"/protocol/openid-connect/userinfo",
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + token},
	}

	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer do.Body.Close()

	if do.StatusCode != 200 {
		return nil, util.ErrCheck(errors.New(fmt.Sprintf("kc user info status %d", do.StatusCode)))
	}

	resp, err := io.ReadAll(do.Body)
	if err != nil {
		return nil, err
	}

	var result KeycloakUser
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Printf("user info unmarshal body: %s", string(resp))
		return nil, err
	}

	result.Id = result.Sub

	return &result, nil
}

func (keycloakClient KeycloakClient) GetUserInfoById(userId string) (*KeycloakUser, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users/"+userId,
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, err
	}

	var result KeycloakUser
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	result.Sub = result.Id

	return &result, nil
}

func (keycloakClient KeycloakClient) GetUserGroups(userId string) (*[]KeycloakUserGroup, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users/"+userId+"/groups",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, err
	}

	var result []KeycloakUserGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) GetGroupRoleMappings(groupId string) ([]ClientRoleMappingRole, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/role-mappings",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, err
	}

	var f map[string]map[string]ClientRoleMapping

	if err := json.Unmarshal(resp, &f); err != nil {
		return nil, err
	}

	mappings := []ClientRoleMappingRole{}

	clientMappings := f["clientMappings"]
	for clientId, client := range clientMappings {
		if clientId == keycloakClient.AppClient.ClientID {
			mappings = client.Mappings
			break
		}
	}

	return mappings, nil
}

func (keycloakClient KeycloakClient) GetAppClientRoles() (*[]KeycloakRole, error) {
	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/clients/"+keycloakClient.AppClient.Id+"/roles",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, err
	}

	var result []KeycloakRole
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) GetRealmClients() (*[]KeycloakRealmClient, error) {
	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/clients",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, err
	}

	var result []KeycloakRealmClient
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) MutateRoleCall(method, userId string) error {

	_, err := util.Mutate(
		method,
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users/"+userId+"/role-mappings/clients/"+keycloakClient.AppClient.Id,
		keycloakClient.BasicHeaders(),
		[]byte(`[{ "id": "`+keycloakClient.RoleCall.Id+`", "name": "`+keycloakClient.RoleCall.Name+`" }]`),
	)

	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient KeycloakClient) UpdateUser(userId, firstName, lastName string) error {

	kcUserJson, err := json.Marshal(&KeycloakUser{
		Id:         userId,
		FirstName:  firstName,
		LastName:   lastName,
		Attributes: nil,
		Access:     nil,
	})

	if err != nil {
		return err
	}

	_, err = util.Mutate(
		"PUT",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users/"+userId,
		keycloakClient.BasicHeaders(),
		kcUserJson,
	)

	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient KeycloakClient) CreateGroup(name string) (string, error) {

	println("CREATE GROUP REQ", keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups", `{ "name": "`+name+`" }`)

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups",
		bytes.NewBuffer([]byte(`{ "name": "`+name+`" }`)),
	)

	if err != nil {
		return "", err
	}

	req.Header = keycloakClient.BasicHeaders()

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 201 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("create group error code %d %s", resp.StatusCode, string(body)))
	}

	locationHeader := resp.Header.Get("Location")

	parts := strings.SplitAfterN(locationHeader, "/", -1)
	if locationHeader == "" || len(parts) < 2 {
		return "", errors.New("no group id found")
	}

	return parts[len(parts)-1], nil
}

func (keycloakClient KeycloakClient) DeleteGroup(groupId string) error {
	_, err := util.Mutate(
		"DELETE",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId,
		keycloakClient.BasicHeaders(),
		[]byte(""),
	)

	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient KeycloakClient) UpdateGroup(groupId, groupName string) error {
	_, err := util.Mutate(
		"PUT",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId,
		keycloakClient.BasicHeaders(),
		[]byte(`{ "name": "`+groupName+`" }`),
	)

	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient KeycloakClient) GetGroup(groupId string) (*KeycloakGroup, error) {
	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId,
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, err
	}

	var result KeycloakGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) CreateSubgroup(groupId, subgroupName string) (*KeycloakGroup, error) {
	resp, err := util.Mutate(
		"POST",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/children",
		keycloakClient.BasicHeaders(),
		[]byte(`{ "name": "`+subgroupName+`" }`),
	)

	if err != nil {
		return nil, err
	}

	var result KeycloakGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) GetGroupSubgroups(groupId string) (*[]KeycloakGroup, error) {
	queryParams := url.Values{}
	queryParams.Add("first", "0")
	queryParams.Add("max", "11")

	resp, err := util.GetWithParams(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/children",
		keycloakClient.BasicHeaders(),
		queryParams,
	)

	if err != nil {
		return nil, err
	}

	var result []KeycloakGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) MutateGroupRoles(method, groupId string, roles []KeycloakRole) error {

	rolesBytes, err := json.Marshal(roles)
	if err != nil {
		return err
	}

	_, err = util.Mutate(
		method,
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/role-mappings/clients/"+keycloakClient.AppClient.Id,
		keycloakClient.BasicHeaders(),
		rolesBytes,
	)

	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient KeycloakClient) MutateUserGroupMembership(method, userId, groupId string) error {
	_, err := util.Mutate(
		method,
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users/"+userId+"/groups/"+groupId,
		keycloakClient.BasicHeaders(),
		[]byte(""),
	)

	if err != nil {
		return err
	}

	return nil
}
