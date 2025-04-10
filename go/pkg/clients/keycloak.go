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

var (
	KeycloakPublicKey      *rsa.PublicKey
	emptyAuthResponse      = AuthResponse{}
	authCommandMustHaveSub = errors.New("auth command request must contain a user sub")
)

const keycloakHandlerId = "keycloak"

type AuthRequest struct {
	*types.AuthRequestParams
}

type AuthResponse struct {
	*types.AuthResponseParams
	Error error
}

type AuthCommand struct {
	*types.WorkerCommandParams
	Request   AuthRequest
	ReplyChan chan AuthResponse
}

func (cmd AuthCommand) GetClientId() string {
	return cmd.ClientId
}

func (cmd AuthCommand) GetReplyChannel() interface{} {
	return cmd.ReplyChan
}

type Keycloak struct {
	Client    *KeycloakClient
	handlerId string
}

func InitKeycloak() *Keycloak {

	kc := &KeycloakClient{
		Server: os.Getenv("KC_INTERNAL"),
		Realm:  os.Getenv("KC_REALM"),
	}

	InitGlobalWorkerPool(4, 8)

	GetGlobalWorkerPool().RegisterProcessFunction(keycloakHandlerId, func(keycloakCmd CombinedCommand) bool {

		cmd, ok := keycloakCmd.(AuthCommand)
		if !ok {
			return false
		}

		defer func(replyChan chan AuthResponse) {
			if r := recover(); r != nil {
				err := errors.New(fmt.Sprintf("Did recover from %+v", r))
				replyChan <- AuthResponse{Error: err}
			}
			close(replyChan)
		}(cmd.ReplyChan)

		switch cmd.Ty {
		case SetKeycloakTokenKeycloakCommand:
			oidcToken, err := kc.DirectGrantAuthentication()
			if err != nil {
				cmd.ReplyChan <- AuthResponse{Error: err}
			} else {
				kc.Token = oidcToken
				cmd.ReplyChan <- AuthResponse{}
			}
		case SetKeycloakRealmClientsKeycloakCommand:
			realmClients, err := kc.GetRealmClients()
			if err != nil {
				cmd.ReplyChan <- AuthResponse{Error: err}
			} else {
				for _, realmClient := range realmClients {
					if realmClient.ClientId == string(os.Getenv("KC_CLIENT")) {
						kc.AppClient = realmClient
					}
					if realmClient.ClientId == string(os.Getenv("KC_API_CLIENT")) {
						kc.ApiClient = realmClient
					}
				}
				cmd.ReplyChan <- AuthResponse{}
			}
		case SetKeycloakRolesKeycloakCommand:
			var groupAdminRoles []*types.KeycloakRole
			roles, err := kc.GetAppClientRoles()
			if err != nil {
				cmd.ReplyChan <- AuthResponse{Error: err}
			} else {
				for _, role := range roles {
					if role.Name == "APP_ROLE_CALL" {
						continue
					} else {
						groupAdminRoles = append(groupAdminRoles, role)
					}
				}
				kc.GroupAdminRoles = groupAdminRoles
				cmd.ReplyChan <- AuthResponse{}
			}
		case GetGroupAdminRolesKeycloakCommand:
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Roles: kc.GroupAdminRoles,
				},
			}
		case GetUserListKeycloakCommand:
			users, err := kc.GetUserListInRealm()
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Users: users,
				},
				Error: err,
			}
		case UpdateUserKeycloakCommand:
			err := kc.UpdateUser(cmd.Request.UserId, cmd.Request.FirstName, cmd.Request.LastName)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case CreateGroupKeycloakCommand:
			groupId, err := kc.CreateGroup(cmd.Request.GroupName)
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Group: &types.KeycloakGroup{Id: groupId},
				},
				Error: err,
			}
		case DeleteGroupKeycloakCommand:
			err := kc.DeleteGroup(cmd.Request.GroupId)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case UpdateGroupKeycloakCommand:
			err := kc.UpdateGroup(cmd.Request.GroupId, cmd.Request.GroupName)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case GetGroupKeycloakCommand:
			group, err := kc.GetGroup(cmd.Request.GroupId)
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Group: group,
				},
				Error: err,
			}
		case CreateSubgroupKeycloakCommand:
			subgroup, err := kc.CreateSubgroup(cmd.Request.GroupId, cmd.Request.GroupName)
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Group: subgroup,
				},
				Error: err,
			}
		case AddUserToGroupKeycloakCommand:
			err := kc.MutateUserGroupMembership(http.MethodPut, cmd.Request.UserId, cmd.Request.GroupId)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case DeleteUserFromGroupKeycloakCommand:
			err := kc.MutateUserGroupMembership(http.MethodDelete, cmd.Request.UserId, cmd.Request.GroupId)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case AddRolesToGroupKeycloakCommand:
			err := kc.MutateGroupRoles(http.MethodPost, cmd.Request.GroupId, cmd.Request.Roles)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case DeleteRolesFromGroupKeycloakCommand:
			err := kc.MutateGroupRoles(http.MethodDelete, cmd.Request.GroupId, cmd.Request.Roles)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case GetGroupByNameKeycloakCommand:
			var groups []*types.KeycloakGroup
			findBytes, err := kc.FindResource("groups", cmd.Request.GroupName)
			json.Unmarshal(findBytes, &groups)
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Groups: groups,
				},
				Error: err,
			}
		case GetGroupSubgroupsKeycloakCommand:
			subgroups, err := kc.GetGroupSubgroups(cmd.Request.GroupId)
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Groups: subgroups,
				},
				Error: err,
			}
		case GetGroupRoleMappingsKeycloakCommand:
			mappings, err := kc.GetGroupRoleMappings(cmd.Request.GroupId)
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Mappings: mappings,
				},
				Error: err,
			}
		default:
			cmd.ReplyChan <- AuthResponse{
				Error: errors.New("unknown command type"),
			}
			util.ErrorLog.Println("unknown command type", cmd.Ty)
		}

		return true
	})

	kcc := &Keycloak{handlerId: keycloakHandlerId}

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				response, err := kcc.SendCommand(SetKeycloakTokenKeycloakCommand, &types.AuthRequestParams{
					UserSub: "worker",
				})
				if err = ChannelError(err, response.Error); err != nil {
					util.ErrorLog.Println(util.ErrCheck(response.Error))
				}
			}
		}
	}()

	response, err := kcc.SendCommand(SetKeycloakTokenKeycloakCommand, &types.AuthRequestParams{
		UserSub: "worker",
	})
	if err = ChannelError(err, response.Error); err != nil {
		log.Fatal(util.ErrCheck(response.Error))
	}

	response, err = kcc.SendCommand(SetKeycloakRealmClientsKeycloakCommand, &types.AuthRequestParams{
		UserSub: "worker",
	})
	if err = ChannelError(err, response.Error); err != nil {
		log.Fatal(util.ErrCheck(response.Error))
	}

	response, err = kcc.SendCommand(SetKeycloakRolesKeycloakCommand, &types.AuthRequestParams{
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

	kcc.Client = kc

	println("Keycloak Init")
	return kcc
}

func (s *Keycloak) RouteCommand(cmd AuthCommand) error {
	return GetGlobalWorkerPool().RouteCommand(cmd)
}

func (s *Keycloak) Close() {
	GetGlobalWorkerPool().UnregisterProcessFunction(s.handlerId)
}

func (k *Keycloak) SendCommand(cmdType int32, request *types.AuthRequestParams) (AuthResponse, error) {
	if request.UserSub == "" {
		return emptyAuthResponse, authCommandMustHaveSub
	}
	createCmd := func(replyChan chan AuthResponse) AuthCommand {
		return AuthCommand{
			WorkerCommandParams: &types.WorkerCommandParams{
				Ty:       cmdType,
				ClientId: request.UserSub,
			},
			Request: AuthRequest{
				AuthRequestParams: request,
			},
			ReplyChan: replyChan,
		}
	}

	return SendCommand(k, createCmd)
}

func (k *Keycloak) UpdateUser(userSub, id, firstName, lastName string) error {
	response, err := k.SendCommand(UpdateUserKeycloakCommand, &types.AuthRequestParams{
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
	response, err := k.SendCommand(GetGroupAdminRolesKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Roles, nil
}

func (k *Keycloak) GetGroupSiteRoles(userSub, groupId string) ([]*types.ClientRoleMappingRole, error) {
	response, err := k.SendCommand(GetGroupRoleMappingsKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: groupId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Mappings, nil
}

func (k *Keycloak) CreateGroup(userSub, name string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(CreateGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Group, nil
}

func (k *Keycloak) GetGroup(userSub, id string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: id,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Group, nil
}

func (k *Keycloak) GetGroupByName(userSub, name string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupByNameKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		GroupName: name,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Groups, nil
}

func (k *Keycloak) GetGroupSubgroups(userSub, groupId string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(GetGroupSubgroupsKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: groupId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.Groups, nil
}

func (k *Keycloak) DeleteGroup(userSub, id string) error {
	response, err := k.SendCommand(DeleteGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: id,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) UpdateGroup(userSub, id, name string) error {
	response, err := k.SendCommand(UpdateGroupKeycloakCommand, &types.AuthRequestParams{
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
	kcCreateSubgroup, err := k.SendCommand(CreateSubgroupKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		GroupId:   groupExternalId,
		GroupName: subGroupName,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	var kcSubGroup *types.KeycloakGroup

	if kcCreateSubgroup.Error != nil && strings.Contains(kcCreateSubgroup.Error.Error(), "exists") {
		groupSubgroupsReply, err := k.SendCommand(GetGroupSubgroupsKeycloakCommand, &types.AuthRequestParams{
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
	response, err := k.SendCommand(AddRolesToGroupKeycloakCommand, &types.AuthRequestParams{
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
	response, err := k.SendCommand(AddUserToGroupKeycloakCommand, &types.AuthRequestParams{
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
	response, err := k.SendCommand(DeleteUserFromGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: groupId,
		UserId:  deletingUserId,
	})

	if err = ChannelError(err, response.Error); err != nil {
		return util.ErrCheck(err)
	}

	return nil
}
