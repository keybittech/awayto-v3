package clients

// https://github.com/PhilippHeuer/go-keycloak with mods

import (
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
	"strconv"
	"strings"

	json "encoding/json"
	"encoding/pem"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/golang-jwt/jwt"
)

type KeycloakUserWithClaims struct {
	types.KeycloakUser
	jwt.StandardClaims
	ResourceAccess map[string]struct {
		Roles []string `json:"roles,omitempty"`
	} `json:"resource_access,omitempty"`
}

type KeycloakClient struct {
	GroupAdminRoles []*types.KeycloakRole
	Server          string
	Realm           string
	AppClient       *types.KeycloakRealmClient
	ApiClient       *types.KeycloakRealmClient
	Token           *types.OIDCToken
	PublicKey       *rsa.PublicKey
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
		log.Fatal(util.ErrCheck(err))
	}

	var result types.KeycloakRealmInfo
	if err := protojson.Unmarshal(resp, &result); err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	block, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\n" + result.PublicKey + "\n-----END PUBLIC KEY-----"))
	if block == nil {
		log.Fatal(util.ErrCheck(errors.New("empty pem block")))
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	if parsed, ok := pubKey.(*rsa.PublicKey); ok {
		return parsed, nil
	}

	log.Fatal(util.ErrCheck(errors.New("key could not be parsed")))
	return nil, nil
}

func (keycloakClient KeycloakClient) DirectGrantAuthentication() (*types.OIDCToken, error) {

	headers := http.Header{}

	clientSecret, err := util.GetEnvFile("KC_API_CLIENT_SECRET_FILE", 128)
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

	var result types.OIDCToken
	if err := protojson.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	if result.ExpiresIn > 0 {
		return &result, nil
	}

	return nil, util.ErrCheck(errors.New("Authentication failed"))
}

func (keycloakClient KeycloakClient) FindResource(resource, search string, first, last int) ([]byte, error) {

	queryParams := url.Values{}
	queryParams.Add("first", strconv.Itoa(first))
	queryParams.Add("max", strconv.Itoa(last))
	queryParams.Add("search", search)

	resp, err := util.GetWithParams(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/"+resource,
		keycloakClient.BasicHeaders(),
		queryParams,
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return resp, nil
}

func (keycloakClient KeycloakClient) GetUserListInRealm() ([]*types.KeycloakUser, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result []*types.KeycloakUser
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	return result, nil
}

func (keycloakClient KeycloakClient) GetUserGroups(userId string) ([]*types.KeycloakUserGroup, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users/"+userId+"/groups",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result []*types.KeycloakUserGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	return result, nil
}

func (keycloakClient KeycloakClient) GetGroupRoleMappings(groupId string) ([]*types.ClientRoleMappingRole, error) {

	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/role-mappings",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var f map[string]map[string]*types.ClientRoleMapping

	if err := json.Unmarshal(resp, &f); err != nil {
		return nil, util.ErrCheck(err)
	}

	mappings := []*types.ClientRoleMappingRole{}

	clientMappings := f["clientMappings"]
	for clientId, client := range clientMappings {
		if clientId == keycloakClient.AppClient.ClientId {
			mappings = client.Mappings
			break
		}
	}

	return mappings, nil
}

func (keycloakClient KeycloakClient) GetAppClientRoles() ([]*types.KeycloakRole, error) {
	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/clients/"+keycloakClient.AppClient.Id+"/roles",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result []*types.KeycloakRole
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	return result, nil
}

func (keycloakClient KeycloakClient) GetRealmClients() ([]*types.KeycloakRealmClient, error) {
	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/clients",
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result []*types.KeycloakRealmClient
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	return result, nil
}

func (keycloakClient KeycloakClient) UpdateUser(userId, firstName, lastName string) error {

	kcUser := &KeycloakUserWithClaims{}
	kcUser.Id = userId
	kcUser.FirstName = firstName
	kcUser.LastName = lastName

	kcUserJson, err := json.Marshal(kcUser)

	if err != nil {
		return util.ErrCheck(err)
	}

	_, err = util.Mutate(
		"PUT",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/users/"+userId,
		keycloakClient.BasicHeaders(),
		kcUserJson,
	)

	if err != nil {
		return util.ErrCheck(err)
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
		return "", util.ErrCheck(errors.New(fmt.Sprintf("create group error code %d %s", resp.StatusCode, string(body))))
	}

	locationHeader := resp.Header.Get("Location")

	parts := strings.SplitAfterN(locationHeader, "/", -1)
	if locationHeader == "" || len(parts) < 2 {
		return "", util.ErrCheck(errors.New("no group id found"))
	}

	return parts[len(parts)-1], nil
}

func (keycloakClient KeycloakClient) DeleteGroup(groupId string) error {
	deleteGroupRes, err := util.Mutate(
		"DELETE",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId,
		keycloakClient.BasicHeaders(),
		[]byte(""),
	)
	if err != nil {
		return util.ErrCheck(err)
	}

	println("DELETED A GROUP/SUBGROUP RESPONSE", string(deleteGroupRes))

	return nil
}

func (keycloakClient KeycloakClient) UpdateGroup(groupId, groupName string) error {
	groupResponse, err := util.Mutate(
		"PUT",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId,
		keycloakClient.BasicHeaders(),
		[]byte(`{ "name": "`+groupName+`" }`),
	)
	if err != nil {
		return util.ErrCheck(err)
	}

	println("KEYCLOAK UPDATE GROUP RESPONSE", string(groupResponse))

	return nil
}

func (keycloakClient KeycloakClient) GetGroup(groupId string) (*types.KeycloakGroup, error) {
	resp, err := util.Get(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId,
		keycloakClient.BasicHeaders(),
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result types.KeycloakGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) CreateSubgroup(groupId, subgroupName string) (*types.KeycloakGroup, error) {
	resp, err := util.Mutate(
		"POST",
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/children",
		keycloakClient.BasicHeaders(),
		[]byte(`{ "name": "`+subgroupName+`" }`),
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result types.KeycloakGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	return &result, nil
}

func (keycloakClient KeycloakClient) GetGroupSubgroups(groupId string) ([]*types.KeycloakGroup, error) {
	queryParams := url.Values{}
	queryParams.Add("first", "0")
	queryParams.Add("max", "11")

	resp, err := util.GetWithParams(
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/children",
		keycloakClient.BasicHeaders(),
		queryParams,
	)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var result []*types.KeycloakGroup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, util.ErrCheck(err)
	}

	return result, nil
}

func (keycloakClient KeycloakClient) MutateGroupRoles(method, groupId string, roles []*types.KeycloakRole) error {

	rolesBytes, err := json.Marshal(roles)
	if err != nil {
		return util.ErrCheck(err)
	}

	_, err = util.Mutate(
		method,
		keycloakClient.Server+"/admin/realms/"+keycloakClient.Realm+"/groups/"+groupId+"/role-mappings/clients/"+keycloakClient.AppClient.Id,
		keycloakClient.BasicHeaders(),
		rolesBytes,
	)

	if err != nil {
		return util.ErrCheck(err)
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
		return util.ErrCheck(err)
	}

	return nil
}
