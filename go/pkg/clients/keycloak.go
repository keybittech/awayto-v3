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
	Params    KeycloakParams
	ReplyChan chan KeycloakResponse
}

var (
	KeycloakPublicKey *rsa.PublicKey
)

func InitKeycloak() interfaces.IKeycloak {
	cmds := make(chan KeycloakCommand)

	kc := &KeycloakClient{
		Server: os.Getenv("KC_INTERNAL"),
		Realm:  os.Getenv("KC_REALM"),
	}

	go func() {
		for cmd := range cmds {
			defer func() {
				if r := recover(); r != nil {
					err := errors.New(fmt.Sprintf("Did recover from %+v", r))
					cmd.ReplyChan <- KeycloakResponse{Error: err}
				}
			}()

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
				err := kc.UpdateUser(cmd.Params.UserId, cmd.Params.FirstName, cmd.Params.LastName)
				cmd.ReplyChan <- KeycloakResponse{Error: err}
			case CreateGroupKeycloakCommand:
				groupId, err := kc.CreateGroup(cmd.Params.GroupName)
				cmd.ReplyChan <- KeycloakResponse{Error: err, Group: &types.KeycloakGroup{Id: groupId}}
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
				var groups []*types.KeycloakGroup
				findBytes, err := kc.FindResource("groups", cmd.Params.GroupName)
				json.Unmarshal(findBytes, &groups)
				cmd.ReplyChan <- KeycloakResponse{Groups: groups, Error: err}
			case GetGroupSubgroupsKeycloakCommand:
				subgroups, err := kc.GetGroupSubgroups(cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Groups: subgroups, Error: err}
			case GetGroupRoleMappingsKeycloakCommand:
				mappings, err := kc.GetGroupRoleMappings(cmd.Params.GroupId)
				cmd.ReplyChan <- KeycloakResponse{Mappings: mappings, Error: err}
			default:
				cmd.ReplyChan <- KeycloakResponse{Error: errors.New("unknown command type")}
				util.ErrorLog.Println("unknown command type", cmd.Ty)
			}
			close(cmd.ReplyChan)
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

	kcc := &Keycloak{}
	kcc.Ch = cmds

	response, err := kcc.SendCommand(SetKeycloakTokenKeycloakCommand, KeycloakParams{})
	if err = ChannelError(err, response.Error); err != nil {
		log.Fatal(util.ErrCheck(response.Error))
	}

	response, err = kcc.SendCommand(SetKeycloakRealmClientsKeycloakCommand, KeycloakParams{})
	if err = ChannelError(err, response.Error); err != nil {
		log.Fatal(util.ErrCheck(response.Error))
	}

	response, err = kcc.SendCommand(SetKeycloakRolesKeycloakCommand, KeycloakParams{})
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

func (k *Keycloak) GetCommandChannel() chan<- KeycloakCommand {
	return k.Ch
}

func (k *Keycloak) SendCommand(cmdType KeycloakCommandType, params KeycloakParams) (KeycloakResponse, error) {
	createCmd := func(p KeycloakParams, replyChan chan KeycloakResponse) KeycloakCommand {
		return KeycloakCommand{
			Ty:        cmdType,
			Params:    p,
			ReplyChan: replyChan,
		}
	}

	return SendCommand(k, createCmd, params)
}

func (k *Keycloak) UpdateUser(id, firstName, lastName string) error {
	response, err := k.SendCommand(UpdateUserKeycloakCommand, KeycloakParams{
		UserId:    id,
		FirstName: firstName,
		LastName:  lastName,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) GetGroupAdminRoles() ([]*types.KeycloakRole, error) {
	response, err := k.SendCommand(GetGroupAdminRolesKeycloakCommand, KeycloakParams{})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Roles, nil
}

func (k *Keycloak) GetGroupSiteRoles(groupId string) ([]*types.ClientRoleMappingRole, error) {
	response, err := k.SendCommand(GetGroupRoleMappingsKeycloakCommand, KeycloakParams{
		GroupId: groupId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Mappings, nil
}

func (k *Keycloak) CreateGroup(name string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(CreateGroupKeycloakCommand, KeycloakParams{
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Group, nil
}

func (k *Keycloak) GetGroup(id string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupKeycloakCommand, KeycloakParams{
		GroupId: id,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Group, nil
}

func (k *Keycloak) GetGroupByName(name string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupByNameKeycloakCommand, KeycloakParams{
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Groups, nil
}

func (k *Keycloak) GetGroupSubgroups(groupId string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupSubgroupsKeycloakCommand, KeycloakParams{
		GroupId: groupId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Groups, nil
}

func (k *Keycloak) DeleteGroup(id string) error {
	response, err := k.SendCommand(DeleteGroupKeycloakCommand, KeycloakParams{
		GroupId: id,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) UpdateGroup(id, name string) error {
	response, err := k.SendCommand(UpdateGroupKeycloakCommand, KeycloakParams{
		GroupId:   id,
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) CreateOrGetSubGroup(groupExternalId, subGroupName string) (*types.KeycloakGroup, error) {
	kcCreateSubgroup, err := k.SendCommand(CreateSubgroupKeycloakCommand, KeycloakParams{
		GroupId:   groupExternalId,
		GroupName: subGroupName,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var kcSubGroup *types.KeycloakGroup

	if kcCreateSubgroup.Error != nil && strings.Contains(kcCreateSubgroup.Error.Error(), "exists") {
		groupSubgroupsReply, err := k.SendCommand(GetGroupSubgroupsKeycloakCommand, KeycloakParams{
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

func (k *Keycloak) AddRolesToGroup(id string, roles []*types.KeycloakRole) error {
	response, err := k.SendCommand(AddRolesToGroupKeycloakCommand, KeycloakParams{
		GroupId: id,
		Roles:   roles,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) DeleteRolesFromGroup(id string, roles []*types.KeycloakRole) error {
	response, err := k.SendCommand(DeleteRolesFromGroupKeycloakCommand, KeycloakParams{
		GroupId: id,
		Roles:   roles,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) AddUserToGroup(userId, groupId string) error {
	response, err := k.SendCommand(AddUserToGroupKeycloakCommand, KeycloakParams{
		GroupId: groupId,
		UserId:  userId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) DeleteUserFromGroup(userId, groupId string) error {
	response, err := k.SendCommand(DeleteUserFromGroupKeycloakCommand, KeycloakParams{
		GroupId: groupId,
		UserId:  userId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}
