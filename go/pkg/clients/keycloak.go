package clients

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	_ "github.com/lib/pq"
)

type KeycloakCommandType int

const (
	SetKeycloakTokenKeycloakCommand = iota
	SetKeycloakRealmClientsKeycloakCommand
	SetKeycloakRolesKeycloakCommand
	GetGroupAdminRolesKeycloakCommand
	GetUserListKeycloakCommand
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

type KeycloakRequest struct {
	Method    string
	Resource  string
	UserSub   string
	UserId    string
	GroupId   string
	FirstName string
	LastName  string
	GroupName string
	Token     string
	Roles     []*types.KeycloakRole
}

type KeycloakResponse struct {
	User        *types.KeycloakUser
	UserSession *types.UserSession
	Users       []*types.KeycloakUser
	Group       *types.KeycloakGroup
	Groups      []*types.KeycloakGroup
	Roles       []*types.KeycloakRole
	Mappings    []*types.ClientRoleMappingRole
	Error       error
	Valid       bool
}

type KeycloakCommand struct {
	Ty        KeycloakCommandType
	ClientId  string
	Request   KeycloakRequest
	ReplyChan chan KeycloakResponse
}

func (cmd KeycloakCommand) GetClientId() string {
	return cmd.ClientId
}

func (cmd KeycloakCommand) GetReplyChannel() interface{} {
	return cmd.ReplyChan
}

var (
	KeycloakPublicKey *rsa.PublicKey
)

type Keycloak struct {
	interfaces.IKeycloak
	C         *KeycloakClient
	handlerId string
}

const keycloakHandlerId = "keycloak"

func InitKeycloak() *Keycloak {
	cmds := make(chan KeycloakCommand)

	kc := &KeycloakClient{
		Server: os.Getenv("KC_INTERNAL"),
		Realm:  os.Getenv("KC_REALM"),
	}

	InitGlobalWorkerPool(4, 8)

	GetGlobalWorkerPool().RegisterProcessFunction(keycloakHandlerId, func(keycloakCmd CombinedCommand) bool {

		cmd, ok := keycloakCmd.(KeycloakCommand)
		if !ok {
			return false
		}

		defer func(replyChan chan KeycloakResponse) {
			if r := recover(); r != nil {
				err := errors.New(fmt.Sprintf("Did recover from %+v", r))
				replyChan <- KeycloakResponse{Error: err}
			}
			close(replyChan)
		}(cmd.ReplyChan)

		switch cmd.Ty {
		case SetKeycloakTokenKeycloakCommand:
			oidcToken, err := kc.DirectGrantAuthentication()
			if err != nil {
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			} else {
				kc.Token = oidcToken
				cmd.ReplyChan <- KeycloakResponse{}
			}
		case SetKeycloakRealmClientsKeycloakCommand:
			realmClients, err := kc.GetRealmClients()
			if err != nil {
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			} else {
				for _, realmClient := range realmClients {
					if realmClient.ClientId == string(os.Getenv("KC_CLIENT")) {
						kc.AppClient = realmClient
					}
					if realmClient.ClientId == string(os.Getenv("KC_API_CLIENT")) {
						kc.ApiClient = realmClient
					}
				}
				cmd.ReplyChan <- KeycloakResponse{}
			}
		case SetKeycloakRolesKeycloakCommand:
			var groupAdminRoles []*types.KeycloakRole
			roles, err := kc.GetAppClientRoles()
			if err != nil {
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			} else {
				for _, role := range roles {
					if role.Name == "APP_ROLE_CALL" {
						continue
					} else {
						groupAdminRoles = append(groupAdminRoles, role)
					}
				}
				kc.GroupAdminRoles = groupAdminRoles
				cmd.ReplyChan <- KeycloakResponse{}
			}
		case GetGroupAdminRolesKeycloakCommand:
			cmd.ReplyChan <- KeycloakResponse{Roles: kc.GroupAdminRoles}
		case GetUserListKeycloakCommand:
			users, err := kc.GetUserListInRealm()
			cmd.ReplyChan <- KeycloakResponse{Users: users, Error: err}
		case UpdateUserKeycloakCommand:
			err := kc.UpdateUser(cmd.Request.UserId, cmd.Request.FirstName, cmd.Request.LastName)
			cmd.ReplyChan <- KeycloakResponse{Error: err}
		case CreateGroupKeycloakCommand:
			groupId, err := kc.CreateGroup(cmd.Request.GroupName)
			cmd.ReplyChan <- KeycloakResponse{Error: err, Group: &types.KeycloakGroup{Id: groupId}}
		case DeleteGroupKeycloakCommand:
			err := kc.DeleteGroup(cmd.Request.GroupId)
			cmd.ReplyChan <- KeycloakResponse{Error: err}
		case UpdateGroupKeycloakCommand:
			err := kc.UpdateGroup(cmd.Request.GroupId, cmd.Request.GroupName)
			cmd.ReplyChan <- KeycloakResponse{Error: err}
		case GetGroupKeycloakCommand:
			group, err := kc.GetGroup(cmd.Request.GroupId)
			cmd.ReplyChan <- KeycloakResponse{Group: group, Error: err}
		case CreateSubgroupKeycloakCommand:
			subgroup, err := kc.CreateSubgroup(cmd.Request.GroupId, cmd.Request.GroupName)
			cmd.ReplyChan <- KeycloakResponse{Error: err, Group: subgroup}
		case AddUserToGroupKeycloakCommand:
			err := kc.MutateUserGroupMembership(http.MethodPut, cmd.Request.UserId, cmd.Request.GroupId)
			cmd.ReplyChan <- KeycloakResponse{Error: err}
		case DeleteUserFromGroupKeycloakCommand:
			err := kc.MutateUserGroupMembership(http.MethodDelete, cmd.Request.UserId, cmd.Request.GroupId)
			cmd.ReplyChan <- KeycloakResponse{Error: err}
		case AddRolesToGroupKeycloakCommand:
			err := kc.MutateGroupRoles(http.MethodPost, cmd.Request.GroupId, cmd.Request.Roles)
			cmd.ReplyChan <- KeycloakResponse{Error: err}
		case DeleteRolesFromGroupKeycloakCommand:
			err := kc.MutateGroupRoles(http.MethodDelete, cmd.Request.GroupId, cmd.Request.Roles)
			cmd.ReplyChan <- KeycloakResponse{Error: err}
		case GetGroupByNameKeycloakCommand:
			var groups []*types.KeycloakGroup
			findBytes, err := kc.FindResource("groups", cmd.Request.GroupName)
			json.Unmarshal(findBytes, &groups)
			cmd.ReplyChan <- KeycloakResponse{Groups: groups, Error: err}
		case GetGroupSubgroupsKeycloakCommand:
			subgroups, err := kc.GetGroupSubgroups(cmd.Request.GroupId)
			cmd.ReplyChan <- KeycloakResponse{Groups: subgroups, Error: err}
		case GetGroupRoleMappingsKeycloakCommand:
			mappings, err := kc.GetGroupRoleMappings(cmd.Request.GroupId)
			cmd.ReplyChan <- KeycloakResponse{Mappings: mappings, Error: err}
		default:
			cmd.ReplyChan <- KeycloakResponse{Error: errors.New("unknown command type")}
			util.ErrorLog.Println("unknown command type", cmd.Ty)
		}

		return true
	})

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				cmds <- KeycloakCommand{Ty: SetKeycloakTokenKeycloakCommand}
			}
		}
	}()

	kcc := &Keycloak{handlerId: keycloakHandlerId}

	response, err := kcc.SendCommand(SetKeycloakTokenKeycloakCommand, KeycloakRequest{
		UserSub: "worker",
	})
	if err = ChannelError(err, response.Error); err != nil {
		log.Fatal(util.ErrCheck(response.Error))
	}

	response, err = kcc.SendCommand(SetKeycloakRealmClientsKeycloakCommand, KeycloakRequest{
		UserSub: "worker",
	})
	if err = ChannelError(err, response.Error); err != nil {
		log.Fatal(util.ErrCheck(response.Error))
	}

	response, err = kcc.SendCommand(SetKeycloakRolesKeycloakCommand, KeycloakRequest{
		UserSub: "worker",
	})
	if err = ChannelError(err, response.Error); err != nil {
		log.Fatal(util.ErrCheck(response.Error))
	}

	pk, err := kc.FetchPublicKey()
	if err != nil {
		log.Fatal(err)
	}
	kc.PublicKey = pk
	KeycloakPublicKey = pk

	kcc.C = kc

	println("Keycloak Init")
	return kcc
}

func (s *Keycloak) RouteCommand(cmd KeycloakCommand) error {
	return GetGlobalWorkerPool().RouteCommand(cmd)
}

func (s *Keycloak) Close() {
	GetGlobalWorkerPool().UnregisterProcessFunction(s.handlerId)
}

func (k *Keycloak) SendCommand(cmdType KeycloakCommandType, params KeycloakRequest) (KeycloakResponse, error) {
	if params.UserSub == "" {
		return KeycloakResponse{}, errors.New("keycloak command request must contain user sub")
	}
	createCmd := func(replyChan chan KeycloakResponse) KeycloakCommand {
		return KeycloakCommand{
			Ty:        cmdType,
			Request:   params,
			ReplyChan: replyChan,
			ClientId:  params.UserSub,
		}
	}

	return SendCommand(k, createCmd)
}

func (k *Keycloak) UpdateUser(userSub, id, firstName, lastName string) error {
	response, err := k.SendCommand(UpdateUserKeycloakCommand, KeycloakRequest{
		UserSub:   userSub,
		UserId:    id,
		FirstName: firstName,
		LastName:  lastName,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) GetGroupAdminRoles(userSub string) ([]*types.KeycloakRole, error) {
	response, err := k.SendCommand(GetGroupAdminRolesKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Roles, nil
}

func (k *Keycloak) GetGroupSiteRoles(userSub, groupId string) ([]*types.ClientRoleMappingRole, error) {
	response, err := k.SendCommand(GetGroupRoleMappingsKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
		GroupId: groupId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Mappings, nil
}

func (k *Keycloak) CreateGroup(userSub, name string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(CreateGroupKeycloakCommand, KeycloakRequest{
		UserSub:   userSub,
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Group, nil
}

func (k *Keycloak) GetGroup(userSub, id string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
		GroupId: id,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Group, nil
}

func (k *Keycloak) GetGroupByName(userSub, name string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupByNameKeycloakCommand, KeycloakRequest{
		UserSub:   userSub,
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Groups, nil
}

func (k *Keycloak) GetGroupSubgroups(userSub, groupId string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupSubgroupsKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
		GroupId: groupId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Groups, nil
}

func (k *Keycloak) DeleteGroup(userSub, id string) error {
	response, err := k.SendCommand(DeleteGroupKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
		GroupId: id,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) UpdateGroup(userSub, id, name string) error {
	response, err := k.SendCommand(UpdateGroupKeycloakCommand, KeycloakRequest{
		UserSub:   userSub,
		GroupId:   id,
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) CreateOrGetSubGroup(userSub, groupExternalId, subGroupName string) (*types.KeycloakGroup, error) {
	kcCreateSubgroup, err := k.SendCommand(CreateSubgroupKeycloakCommand, KeycloakRequest{
		UserSub:   userSub,
		GroupId:   groupExternalId,
		GroupName: subGroupName,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var kcSubGroup *types.KeycloakGroup

	if kcCreateSubgroup.Error != nil && strings.Contains(kcCreateSubgroup.Error.Error(), "exists") {
		groupSubgroupsReply, err := k.SendCommand(GetGroupSubgroupsKeycloakCommand, KeycloakRequest{
			UserSub:   userSub,
			GroupId:   groupExternalId,
			GroupName: subGroupName,
		})

		if err = ChannelError(err, groupSubgroupsReply.Error); err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, sg := range groupSubgroupsReply.Groups {
			if sg.Name == subGroupName {
				kcSubGroup = sg
			}
		}
	} else if kcCreateSubgroup.Error != nil || kcCreateSubgroup.Group.Id == "" {
		return nil, util.ErrCheck(kcCreateSubgroup.Error)
	} else {
		kcSubGroup = kcCreateSubgroup.Group
	}

	return kcSubGroup, nil
}

func (k *Keycloak) AddRolesToGroup(userSub, id string, roles []*types.KeycloakRole) error {
	response, err := k.SendCommand(AddRolesToGroupKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
		GroupId: id,
		Roles:   roles,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) AddUserToGroup(userSub, joiningUserId, groupId string) error {
	response, err := k.SendCommand(AddUserToGroupKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
		GroupId: groupId,
		UserId:  joiningUserId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) DeleteUserFromGroup(userSub, deletingUserId, groupId string) error {
	response, err := k.SendCommand(DeleteUserFromGroupKeycloakCommand, KeycloakRequest{
		UserSub: userSub,
		GroupId: groupId,
		UserId:  deletingUserId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}
