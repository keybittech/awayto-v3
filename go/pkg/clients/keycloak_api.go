package clients

// https://github.com/PhilippHeuer/go-keycloak with mods

import (
	// Error Handling
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	// Encoding

	"encoding/json"
	"encoding/pem"

	"github.com/golang-jwt/jwt"
)

const (
	APP_ROLE_CALL           = "APP_ROLE_CALL"
	APP_GROUP_ADMIN         = "APP_GROUP_ADMIN"
	APP_GROUP_BOOKINGS      = "APP_GROUP_BOOKINGS"
	APP_GROUP_SCHEDULES     = "APP_GROUP_SCHEDULES"
	APP_GROUP_SERVICES      = "APP_GROUP_SERVICES"
	APP_GROUP_SCHEDULE_KEYS = "APP_GROUP_SCHEDULE_KEYS"
	APP_GROUP_ROLES         = "APP_GROUP_ROLES"
	APP_GROUP_USERS         = "APP_GROUP_USERS"
	APP_GROUP_PERMISSIONS   = "APP_GROUP_PERMISSIONS"
)

// APP_GROUP_FEATURES      = "APP_GROUP_FEATURES"

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
	GroupAdminRoles []KeycloakRole
	Token           *OIDCToken
	PublicKey       *rsa.PublicKey
}

type KeycloakUser struct {
	jwt.StandardClaims
	Id                  string   `json:"id,omitempty"`
	Sub                 string   `json:"sub,omitempty"`
	CreatedTimestamp    int64    `json:"createdTimestamp,omitempty"`
	Username            string   `json:"username,omitempty"`
	Enabled             bool     `json:"enabled,omitempty"`
	Totp                bool     `json:"totp,omitempty"`
	Name                string   `json:"name,omitempty"`
	PreferredUsername   string   `json:"preferred_username,omitempty"`
	GivenName           string   `json:"given_name,omitempty"`
	FamilyName          string   `json:"family_name,omitempty"`
	EmailVerified       bool     `json:"emailVerified,omitempty"`
	FirstName           string   `json:"firstName,omitempty"`
	LastName            string   `json:"lastName,omitempty"`
	Email               string   `json:"email,omitempty"`
	FederationLink      string   `json:"federationLink,omitempty"`
	Groups              []string `json:"groups,omitempty"`
	AvailableGroupRoles []string `json:"availableGroupRoles,omitempty"`
	Azp                 string   `json:"azp,omitempty"`
	ResourceAccess      map[string]struct {
		Roles []string `json:"roles,omitempty"`
	} `json:"resource_access,omitempty"`
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
	Mappings []ClientRoleMappingRole `json:"mappings"`
}

type ClientRoleMappingRole struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	ScopeParamRequired bool   `json:"scopeParamRequired"`
	Composite          bool   `json:"composite"`
	ClientRole         bool   `json:"clientRole"`
}

type KeycloakRealmClient struct {
	Id       string `json:"id"`
	ClientId string `json:"clientId"`
}

type KeycloakRealmInfo struct {
	Realm           string `json:"realm"`
	PublicKey       string `json:"public_key"`
	TokenService    string `json:"token-service"`
	AccountService  string `json:"account-service"`
	TokensNotBefore int    `json:"tokens-not-before"`
}

func (keycloakClient KeycloakClient) BasicHeaders() http.Header {
	headers := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + keycloakClient.Token.AccessToken},
	}
	return headers
}

func (keycloakClient KeycloakClient) FetchPublicKey() (*rsa.PublicKey, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/realms/"+keycloakClient.Realm,
		keycloakClient.BasicHeaders(),
	)
	if err != nil {
		return nil, err
	}

	var result KeycloakRealmInfo
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", result.PublicKey)))
	if block == nil {
		log.Fatal("empty pem block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	if parsed, ok := pubKey.(*rsa.PublicKey); ok {
		return parsed, nil
	}

	return nil, errors.New("key could not be parsed")
}

func (keycloakClient KeycloakClient) ValidateToken(token string) (*KeycloakUser, error) {
	if strings.Contains(token, "Bearer") {
		token = strings.Split(token, " ")[1]
	}

	parsedToken, err := jwt.ParseWithClaims(token, &KeycloakUser{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("bad signing method")
		}
		return keycloakClient.PublicKey, nil
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if !parsedToken.Valid {
		return nil, util.ErrCheck(errors.New("invalid token during parse"))
	}

	if claims, ok := parsedToken.Claims.(*KeycloakUser); ok {
		return claims, nil
	}

	return nil, nil
}

func (keycloakClient KeycloakClient) DirectGrantAuthentication() (*OIDCToken, error) {

	headers := http.Header{}

	clientSecret, err := util.EnvFile(os.Getenv("KC_API_CLIENT_SECRET_FILE"))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	body := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {os.Getenv("KC_API_CLIENT")},
		"client_secret": {string(clientSecret)},
	}

	resp, err := util.PostFormData(
		keycloakClient.Server+"/realms/"+keycloakClient.Realm+"/protocol/openid-connect/token",
		headers,
		body,
	)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result OIDCToken
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
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
		if clientId == keycloakClient.AppClient.ClientId {
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

func (keycloakClient KeycloakClient) UpdateUser(userId, firstName, lastName string) error {

	kcUserJson, err := json.Marshal(&KeycloakUser{
		Id:        userId,
		FirstName: firstName,
		LastName:  lastName,
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
