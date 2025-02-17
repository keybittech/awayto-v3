package clients

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Keycloak struct {
	C  *KeycloakClient
	Ch chan<- KeycloakCommand
}

type KeycloakCommandType int

const (
	SetKeycloakTokenKeycloakCommand = iota
	SetKeycloakRealmClientsKeycloakCommand
	SetKeycloakRolesKeycloakCommand
	GetGroupAdminRolesKeycloakCommand
	GetUserListKeycloakCommand
	GetUserInfoByIdKeycloakCommand
	GetUserTokenValidKeycloakCommand
	UpdateUserKeycloakCommand
	CreateGroupKeycloakCommand
	DeleteGroupKeycloakCommand
	UpdateGroupKeycloakCommand
	GetGroupKeycloakCommand
	GetGroupByNameKeycloakCommand
	GetGroupSubgroupsKeycloakCommand
	GetGroupRoleMappingsKeycloakCommand
	CreateSubgroupKeycloakCommand
	DeleteUserFromGroupKeycloakCommand
	AddUserToGroupKeycloakCommand
	AddRolesToGroupKeycloakCommand
	DeleteRolesFromGroupKeycloakCommand
)

type KeycloakParams struct {
	Method    string
	Resource  string
	UserId    string
	GroupId   string
	FirstName string
	LastName  string
	GroupName string
	Token     string
	Roles     []KeycloakRole
}

type KeycloakResponse struct {
	User     *KeycloakUser
	Users    *[]KeycloakUser
	Group    *KeycloakGroup
	Groups   *[]KeycloakGroup
	Roles    []KeycloakRole
	Mappings []ClientRoleMappingRole
	Error    error
	Valid    bool
}

type KeycloakCommand struct {
	Ty        KeycloakCommandType
	Params    KeycloakParams
	ReplyChan chan KeycloakResponse
}

type AuthEventDetails struct {
	AuthMethod     string `json:"auth_method"`
	AuthType       string `json:"auth_type"`
	Email          string `json:"email"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Username       string `json:"username"`
	RegisterMethod string `json:"register_method"`
	RedirectUri    string `json:"redirect_uri"`
}

type AuthEvent struct {
	ClientId    string           `json:"clientId"`
	EventId     string           `json:"id"`
	IpAddress   string           `json:"ipAddress"`
	RealmId     string           `json:"realmId"`
	Timestamp   string           `json:"time"`
	UserId      string           `json:"userId"`
	WebhookName string           `json:"webhookName"`
	GroupCode   string           `json:"groupCode"`
	Email       string           `json:"email"`
	FirstName   string           `json:"firstName"`
	LastName    string           `json:"lastName"`
	Username    string           `json:"username"`
	Details     AuthEventDetails `json:"details"`
	Timezone    string           `json:"timezone"`
}

func InitKeycloak() IKeycloak {
	cmds := make(chan KeycloakCommand)

	kc := &KeycloakClient{
		Server: os.Getenv("KC_INTERNAL"),
		Realm:  os.Getenv("KC_REALM"),
	}

	go func() {
		for cmd := range cmds {
			// defer func() {
			// 	if r := recover(); r != nil {
			// 		err := errors.New(fmt.Sprintf("Did recover from %+v", r))
			// 		cmd.ReplyChan <- KeycloakResponse{Error: err}
			// 	}
			// }()

			switch cmd.Ty {
			case SetKeycloakTokenKeycloakCommand:
				oidcToken, err := kc.DirectGrantAuthentication()
				if err != nil {
					cmd.ReplyChan <- KeycloakResponse{Error: err}
				} else {
					kc.Token = oidcToken
				}
			case SetKeycloakRealmClientsKeycloakCommand:
				realmClients, err := kc.GetRealmClients()
				if err != nil {
					cmd.ReplyChan <- KeycloakResponse{Error: err}
				} else {
					for _, realmClient := range *realmClients {
						if realmClient.ClientID == string(os.Getenv("KC_CLIENT")) {
							println("setting app client")
							kc.AppClient = &realmClient
						}
						if realmClient.ClientID == string(os.Getenv("KC_API_CLIENT")) {
							println("setting api client")
							kc.ApiClient = &realmClient
						}
					}
				}
			case SetKeycloakRolesKeycloakCommand:
				var groupAdminRoles []KeycloakRole
				roles, err := kc.GetAppClientRoles()
				if err != nil {
					cmd.ReplyChan <- KeycloakResponse{Error: err}
				} else {
					for _, role := range *roles {
						if role.Name == "APP_ROLE_CALL" {
							continue
						} else {
							groupAdminRoles = append(groupAdminRoles, role)
						}
					}
					kc.GroupAdminRoles = groupAdminRoles
				}
			case GetGroupAdminRolesKeycloakCommand:
				cmd.ReplyChan <- KeycloakResponse{Roles: kc.GroupAdminRoles}
			case GetUserListKeycloakCommand:
				users, err := kc.GetUserListInRealm()
				cmd.ReplyChan <- KeycloakResponse{Users: users, Error: err}
			case GetUserInfoByIdKeycloakCommand:
				user, err := kc.GetUserInfoById(cmd.Params.UserId)
				cmd.ReplyChan <- KeycloakResponse{User: user, Error: err}
			case GetUserTokenValidKeycloakCommand:
				valid, err := kc.ValidateToken(cmd.Params.Token)
				cmd.ReplyChan <- KeycloakResponse{Valid: valid, Error: err}
			case UpdateUserKeycloakCommand:
				err := kc.UpdateUser(cmd.Params.UserId, cmd.Params.FirstName, cmd.Params.LastName)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case CreateGroupKeycloakCommand:
				groupId, err := kc.CreateGroup(cmd.Params.GroupName)
				cmd.ReplyChan <- KeycloakResponse{Error: err, Group: &KeycloakGroup{Id: groupId}}
			case DeleteGroupKeycloakCommand:
				err := kc.DeleteGroup(cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case UpdateGroupKeycloakCommand:
				err := kc.UpdateGroup(cmd.Params.GroupId, cmd.Params.GroupName)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case GetGroupKeycloakCommand:
				group, err := kc.GetGroup(cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Group: group, Error: err}
			case CreateSubgroupKeycloakCommand:
				subgroup, err := kc.CreateSubgroup(cmd.Params.GroupId, cmd.Params.GroupName)
				cmd.ReplyChan <- KeycloakResponse{Error: err, Group: subgroup}
			case AddUserToGroupKeycloakCommand:
				err := kc.MutateUserGroupMembership(http.MethodPut, cmd.Params.UserId, cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case DeleteUserFromGroupKeycloakCommand:
				err := kc.MutateUserGroupMembership(http.MethodDelete, cmd.Params.UserId, cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case AddRolesToGroupKeycloakCommand:
				err := kc.MutateGroupRoles(http.MethodPost, cmd.Params.GroupId, cmd.Params.Roles)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case DeleteRolesFromGroupKeycloakCommand:
				err := kc.MutateGroupRoles(http.MethodDelete, cmd.Params.GroupId, cmd.Params.Roles)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case GetGroupByNameKeycloakCommand:
				var groups []KeycloakGroup
				findBytes, err := kc.FindResource("groups", cmd.Params.GroupName)
				json.Unmarshal(findBytes, &groups)
				cmd.ReplyChan <- KeycloakResponse{Groups: &groups, Error: err}
			case GetGroupSubgroupsKeycloakCommand:
				subgroups, err := kc.GetGroupSubgroups(cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Groups: subgroups, Error: err}
			case GetGroupRoleMappingsKeycloakCommand:
				mappings, err := kc.GetGroupRoleMappings(cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Mappings: mappings, Error: err}
			default:
				log.Fatal("unknown command type", cmd.Ty)
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				cmds <- KeycloakCommand{Ty: SetKeycloakTokenKeycloakCommand}
			}
		}
	}()

	cmds <- KeycloakCommand{Ty: SetKeycloakTokenKeycloakCommand}
	cmds <- KeycloakCommand{Ty: SetKeycloakRealmClientsKeycloakCommand}
	cmds <- KeycloakCommand{Ty: SetKeycloakRolesKeycloakCommand}

	pk, err := kc.FetchPublicKey()
	if err != nil {
		log.Fatal(err)
	}
	kc.PublicKey = pk

	kcc := &Keycloak{}
	kcc.C = kc
	kcc.Ch = cmds
	return kcc
}

func (k *Keycloak) Chan() chan<- KeycloakCommand {
	return k.Ch
}

func (k *Keycloak) Client() *KeycloakClient {
	return k.C
}

func (k *Keycloak) UpdateUser(id, firstName, lastName string) error {
	kcReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty: UpdateUserKeycloakCommand,
		Params: KeycloakParams{
			UserId:    id,
			FirstName: firstName,
			LastName:  lastName,
		},
		ReplyChan: kcReplyChan,
	}

	kcReply := <-kcReplyChan

	if kcReply.Error != nil {
		return kcReply.Error
	}

	return nil
}

func (k *Keycloak) GetUserTokenValid(token string) (bool, error) {
	kcUserInfoTokenReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        GetUserTokenValidKeycloakCommand,
		Params:    KeycloakParams{Token: token},
		ReplyChan: kcUserInfoTokenReplyChan,
	}
	kcUserInfoTokenReply := <-kcUserInfoTokenReplyChan
	close(kcUserInfoTokenReplyChan)

	if kcUserInfoTokenReply.Error != nil {
		return false, kcUserInfoTokenReply.Error
	}

	return kcUserInfoTokenReply.Valid, nil
}

func (k *Keycloak) GetUserInfoById(id string) (*KeycloakUser, error) {
	kcUserInfoReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        GetUserInfoByIdKeycloakCommand,
		Params:    KeycloakParams{UserId: id},
		ReplyChan: kcUserInfoReplyChan,
	}
	kcUserInfoReply := <-kcUserInfoReplyChan
	close(kcUserInfoReplyChan)

	if kcUserInfoReply.Error != nil {
		return nil, kcUserInfoReply.Error
	}

	return kcUserInfoReply.User, nil
}

func (k *Keycloak) GetGroupAdminRoles() []KeycloakRole {
	kcGroupAdminRolesChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        GetGroupAdminRolesKeycloakCommand,
		ReplyChan: kcGroupAdminRolesChan,
	}
	kcGroupAdminRoles := <-kcGroupAdminRolesChan
	close(kcGroupAdminRolesChan)
	return kcGroupAdminRoles.Roles
}

func (k *Keycloak) GetGroupSiteRoles(groupId string) []ClientRoleMappingRole {
	groupRoleMappingsReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        GetGroupRoleMappingsKeycloakCommand,
		Params:    KeycloakParams{GroupId: groupId},
		ReplyChan: groupRoleMappingsReplyChan,
	}
	groupRoleMappingReply := <-groupRoleMappingsReplyChan
	close(groupRoleMappingsReplyChan)

	return groupRoleMappingReply.Mappings
}

func (k *Keycloak) CreateGroup(name string) (*KeycloakGroup, error) {
	kcCreateGroupReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        CreateGroupKeycloakCommand,
		Params:    KeycloakParams{GroupName: name},
		ReplyChan: kcCreateGroupReplyChan,
	}
	kcCreateGroupReply := <-kcCreateGroupReplyChan
	close(kcCreateGroupReplyChan)

	if kcCreateGroupReply.Error != nil {
		return nil, kcCreateGroupReply.Error
	}

	return kcCreateGroupReply.Group, nil
}

func (k *Keycloak) GetGroup(id string) (*KeycloakGroup, error) {
	kcGetGroupChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        GetGroupKeycloakCommand,
		Params:    KeycloakParams{GroupId: id},
		ReplyChan: kcGetGroupChan,
	}
	kcGetGroup := <-kcGetGroupChan
	close(kcGetGroupChan)

	if kcGetGroup.Error != nil {
		return nil, kcGetGroup.Error
	}

	return kcGetGroup.Group, nil
}

func (k *Keycloak) GetGroupByName(name string) (*[]KeycloakGroup, error) {
	groupByNameReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        GetGroupByNameKeycloakCommand,
		Params:    KeycloakParams{GroupName: name},
		ReplyChan: groupByNameReplyChan,
	}
	groupByNameReply := <-groupByNameReplyChan
	close(groupByNameReplyChan)

	return groupByNameReply.Groups, nil
}

func (k *Keycloak) GetGroupSubgroups(groupId string) (*[]KeycloakGroup, error) {
	subgroupsReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        GetGroupSubgroupsKeycloakCommand,
		Params:    KeycloakParams{GroupId: groupId},
		ReplyChan: subgroupsReplyChan,
	}
	subgroupsReply := <-subgroupsReplyChan
	close(subgroupsReplyChan)

	return subgroupsReply.Groups, nil
}

func (k *Keycloak) DeleteGroup(id string) error {
	deleteReply := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        DeleteGroupKeycloakCommand,
		Params:    KeycloakParams{GroupId: id},
		ReplyChan: deleteReply,
	}
	deleteGroup := <-deleteReply
	close(deleteReply)
	if deleteGroup.Error != nil {
		return deleteGroup.Error
	}

	return nil
}

func (k *Keycloak) UpdateGroup(id, name string) error {
	kcUpdateSubgroupChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        UpdateGroupKeycloakCommand,
		Params:    KeycloakParams{GroupId: id, GroupName: name},
		ReplyChan: kcUpdateSubgroupChan,
	}
	kcUpdateSubgroup := <-kcUpdateSubgroupChan
	close(kcUpdateSubgroupChan)

	if kcUpdateSubgroup.Error != nil {
		return kcUpdateSubgroup.Error
	}

	return nil
}

func (k *Keycloak) CreateOrGetSubGroup(groupExternalId, subGroupName string) (*KeycloakGroup, error) {
	kcCreateSubgroupChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        CreateSubgroupKeycloakCommand,
		Params:    KeycloakParams{GroupId: groupExternalId, GroupName: subGroupName},
		ReplyChan: kcCreateSubgroupChan,
	}
	kcCreateSubgroup := <-kcCreateSubgroupChan
	close(kcCreateSubgroupChan)

	var kcSubGroup *KeycloakGroup

	if kcCreateSubgroup.Error != nil && strings.Contains(kcCreateSubgroup.Error.Error(), "exists") {
		groupSubgroupsReplyChan := make(chan KeycloakResponse)
		k.Chan() <- KeycloakCommand{
			Ty:        GetGroupSubgroupsKeycloakCommand,
			Params:    KeycloakParams{GroupId: groupExternalId},
			ReplyChan: groupSubgroupsReplyChan,
		}
		groupSubgroupsReply := <-groupSubgroupsReplyChan
		close(groupSubgroupsReplyChan)

		if groupSubgroupsReply.Error != nil {
			return nil, groupSubgroupsReply.Error
		}

		for _, sg := range *groupSubgroupsReply.Groups {
			if sg.Name == subGroupName {
				kcSubGroup = &sg
			}
		}
	} else if kcCreateSubgroup.Error != nil || kcCreateSubgroup.Group.Id == "" {
		return nil, kcCreateSubgroup.Error
	} else {
		kcSubGroup = kcCreateSubgroup.Group
	}

	return kcSubGroup, nil
}

func (k *Keycloak) AddRolesToGroup(id string, roles []KeycloakRole) error {
	kcAddGroupRolesReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        AddRolesToGroupKeycloakCommand,
		Params:    KeycloakParams{GroupId: id, Roles: roles},
		ReplyChan: kcAddGroupRolesReplyChan,
	}
	kcAddGroupRolesReply := <-kcAddGroupRolesReplyChan
	close(kcAddGroupRolesReplyChan)

	if kcAddGroupRolesReply.Error != nil {
		return kcAddGroupRolesReply.Error
	}

	return nil
}

func (k *Keycloak) DeleteRolesFromGroup(id string, roles []KeycloakRole) error {
	kcDelGroupRoleReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        DeleteRolesFromGroupKeycloakCommand,
		Params:    KeycloakParams{GroupId: id, Roles: roles},
		ReplyChan: kcDelGroupRoleReplyChan,
	}
	deleteGroupRoleReply := <-kcDelGroupRoleReplyChan
	close(kcDelGroupRoleReplyChan)
	if deleteGroupRoleReply.Error != nil {
		return deleteGroupRoleReply.Error
	}

	return nil
}

func (k *Keycloak) AddUserToGroup(userId, groupId string) error {
	kcAddUserToGroupReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        AddUserToGroupKeycloakCommand,
		Params:    KeycloakParams{UserId: userId, GroupId: groupId},
		ReplyChan: kcAddUserToGroupReplyChan,
	}
	kcAddUserToGroupReply := <-kcAddUserToGroupReplyChan
	close(kcAddUserToGroupReplyChan)
	if kcAddUserToGroupReply.Error != nil {
		return kcAddUserToGroupReply.Error
	}

	return nil
}

func (k *Keycloak) DeleteUserFromGroup(userId, groupId string) error {
	deleteUserReplyChan := make(chan KeycloakResponse)
	k.Chan() <- KeycloakCommand{
		Ty:        DeleteUserFromGroupKeycloakCommand,
		Params:    KeycloakParams{GroupId: groupId, UserId: userId},
		ReplyChan: deleteUserReplyChan,
	}
	deleteUserReply := <-deleteUserReplyChan
	close(deleteUserReplyChan)

	if deleteUserReply.Error != nil {
		return deleteUserReply.Error
	}

	return nil
}
