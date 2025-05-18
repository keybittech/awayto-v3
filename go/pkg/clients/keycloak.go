package clients

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
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
	emptyAuthResponse      = AuthResponse{}
	authCommandMustHaveSub = errors.New("auth command request must contain a user sub")
)

const keycloakHandlerId = "keycloak"

type AuthRequest struct {
	*types.AuthRequestParams
}

type AuthResponse struct {
	Error error
	*types.AuthResponseParams
}

type AuthCommand struct {
	Request AuthRequest
	*types.WorkerCommandParams
	ReplyChan chan AuthResponse
}

func (cmd AuthCommand) GetClientId() string {
	return cmd.ClientId
}

func (cmd AuthCommand) GetReplyChannel() any {
	return cmd.ReplyChan
}

type Keycloak struct {
	handlerId string
	Client    *KeycloakClient
	ticker    *time.Ticker
	stopChan  chan struct{}
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
				replyChan <- AuthResponse{
					Error: err,
				}
			}
			close(replyChan)
		}(cmd.ReplyChan)

		switch cmd.Ty {
		case SetKeycloakTokenKeycloakCommand:
			oidcToken, err := kc.DirectGrantAuthentication()
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			kc.Token = oidcToken
			cmd.ReplyChan <- emptyAuthResponse
		case SetKeycloakRealmClientsKeycloakCommand:
			realmClients, err := kc.GetRealmClients()
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			for _, realmClient := range realmClients {
				if realmClient.ClientId == os.Getenv("KC_CLIENT") {
					kc.AppClient = realmClient
				}
				if realmClient.ClientId == os.Getenv("KC_API_CLIENT") {
					kc.ApiClient = realmClient
				}
			}
			cmd.ReplyChan <- emptyAuthResponse
		case SetKeycloakRolesKeycloakCommand:
			var groupAdminRoles []*types.KeycloakRole
			roles, err := kc.GetAppClientRoles()
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			for _, role := range roles {
				if role.Name == "APP_ROLE_CALL" {
					continue
				} else {
					groupAdminRoles = append(groupAdminRoles, role)
				}
			}
			kc.GroupAdminRoles = groupAdminRoles
			cmd.ReplyChan <- emptyAuthResponse

		case GetGroupAdminRolesKeycloakCommand:
			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Roles: kc.GroupAdminRoles,
				},
			}
		case GetUserListKeycloakCommand:
			users, err := kc.GetUserListInRealm()
			if err != nil {
				cmd.ReplyChan <- AuthResponse{Error: err}
				break
			}

			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Users: users,
				},
			}
		case UpdateUserKeycloakCommand:
			err := kc.UpdateUser(cmd.Request.UserId, cmd.Request.FirstName, cmd.Request.LastName)
			cmd.ReplyChan <- AuthResponse{
				Error: err,
			}
		case CreateGroupKeycloakCommand:
			groupId, err := kc.CreateGroup(cmd.Request.GroupName)
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Group: &types.KeycloakGroup{Id: groupId},
				},
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
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Group: group,
				},
				Error: err,
			}
		case CreateSubgroupKeycloakCommand:
			subGroup, err := kc.CreateSubgroup(cmd.Request.GroupId, cmd.Request.GroupName)
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Group: subGroup,
				},
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
			findBytes, err := kc.FindResource("groups", cmd.Request.GroupName, 0, 10)
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			err = json.Unmarshal(findBytes, &groups)
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Groups: groups,
				},
			}
		case GetGroupSubgroupsKeycloakCommand:
			subgroups, err := kc.GetGroupSubgroups(cmd.Request.GroupId)
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Groups: subgroups,
				},
			}
		case GetGroupRoleMappingsKeycloakCommand:
			mappings, err := kc.GetGroupRoleMappings(cmd.Request.GroupId)
			if err != nil {
				cmd.ReplyChan <- AuthResponse{
					Error: err,
				}
				break
			}

			cmd.ReplyChan <- AuthResponse{
				AuthResponseParams: &types.AuthResponseParams{
					Mappings: mappings,
				},
			}
		default:
			cmd.ReplyChan <- AuthResponse{
				Error: errors.New("unknown command type"),
			}
			util.ErrorLog.Println("unknown command type", cmd.Ty)
		}

		return true
	})

	k := &Keycloak{
		handlerId: keycloakHandlerId,
		stopChan:  make(chan struct{}, 1),
	}

	ticker := time.NewTicker(30 * time.Second)
	k.ticker = ticker
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				_, err := k.SendCommand(context.Background(), SetKeycloakTokenKeycloakCommand, &types.AuthRequestParams{
					UserSub: "worker",
				})
				if err != nil {
					util.ErrorLog.Println(util.ErrCheck(err))
				}
			case <-k.stopChan:
				return
			}
		}
	}()

	connected := false
	setupTicker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-setupTicker.C:
			ctx := context.Background()
			_, err := k.SendCommand(ctx, SetKeycloakTokenKeycloakCommand, &types.AuthRequestParams{
				UserSub: "worker",
			})
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				continue
			}

			_, err = k.SendCommand(ctx, SetKeycloakRealmClientsKeycloakCommand, &types.AuthRequestParams{
				UserSub: "worker",
			})
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				continue
			}

			_, err = k.SendCommand(ctx, SetKeycloakRolesKeycloakCommand, &types.AuthRequestParams{
				UserSub: "worker",
			})
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				continue
			}

			pk, err := kc.FetchPublicKey()
			if err != nil {
				util.ErrorLog.Println(util.ErrCheck(err))
				continue
			}

			kc.PublicKey = pk

			k.Client = kc

			connected = true

		default:
		}

		if connected {
			util.DebugLog.Println("Keycloak Init")
			break
		}
	}

	setupTicker.Stop()

	return k
}

func (s *Keycloak) RouteCommand(ctx context.Context, cmd AuthCommand) error {
	return GetGlobalWorkerPool().RouteCommand(ctx, cmd)
}

func (s *Keycloak) Close() {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	if s.stopChan != nil {
		close(s.stopChan)
	}

	GetGlobalWorkerPool().UnregisterProcessFunction(s.handlerId)
}

func (k *Keycloak) SendCommand(ctx context.Context, cmdType int32, request *types.AuthRequestParams) (AuthResponse, error) {
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

	res, err := SendCommand(ctx, k, createCmd)
	err = ChannelError(err, res.Error)
	if err != nil {
		return res, util.ErrCheck(err)
	}

	return res, nil
}

func (k *Keycloak) UpdateUser(ctx context.Context, userSub, id, firstName, lastName string) error {
	_, err := k.SendCommand(ctx, UpdateUserKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		UserId:    id,
		FirstName: firstName,
		LastName:  lastName,
	})

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) GetGroupAdminRoles(ctx context.Context, userSub string) ([]*types.KeycloakRole, error) {
	response, err := k.SendCommand(ctx, GetGroupAdminRolesKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.AuthResponseParams.Roles, nil
}

func (k *Keycloak) GetGroupSiteRoles(ctx context.Context, userSub, groupId string) ([]*types.ClientRoleMappingRole, error) {
	response, err := k.SendCommand(ctx, GetGroupRoleMappingsKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: groupId,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.AuthResponseParams.Mappings, nil
}

func (k *Keycloak) CreateGroup(ctx context.Context, userSub, name string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(ctx, CreateGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		GroupName: name,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.AuthResponseParams.Group, nil
}

func (k *Keycloak) GetGroup(ctx context.Context, userSub, id string) (*types.KeycloakGroup, error) {
	response, err := k.SendCommand(ctx, GetGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: id,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.AuthResponseParams.Group, nil
}

func (k *Keycloak) GetGroupByName(ctx context.Context, userSub, name string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(ctx, GetGroupByNameKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		GroupName: name,
	})

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.AuthResponseParams.Groups, nil
}

func (k *Keycloak) GetGroupSubGroups(ctx context.Context, userSub, groupId string) ([]*types.KeycloakGroup, error) {
	response, err := k.SendCommand(ctx, GetGroupSubgroupsKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: groupId,
	})
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return response.AuthResponseParams.Groups, nil
}

func (k *Keycloak) DeleteGroup(ctx context.Context, userSub, id string) error {
	_, err := k.SendCommand(ctx, DeleteGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: id,
	})

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) UpdateGroup(ctx context.Context, userSub, id, name string) error {
	_, err := k.SendCommand(ctx, UpdateGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		GroupId:   id,
		GroupName: name,
	})

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) CreateOrGetSubGroup(ctx context.Context, userSub, groupExternalId, subGroupName string) (*types.KeycloakGroup, error) {
	kcCreateSubgroup, err := k.SendCommand(ctx, CreateSubgroupKeycloakCommand, &types.AuthRequestParams{
		UserSub:   userSub,
		GroupId:   groupExternalId,
		GroupName: subGroupName,
	})
	if err != nil && !strings.Contains(err.Error(), "Conflict") && !strings.Contains(err.Error(), "exists") {
		return nil, util.ErrCheck(err)
	}

	if kcCreateSubgroup.AuthResponseParams.Group != nil {
		return kcCreateSubgroup.AuthResponseParams.Group, nil
	} else {
		groupSubgroupsReply, err := k.SendCommand(ctx, GetGroupSubgroupsKeycloakCommand, &types.AuthRequestParams{
			UserSub:   userSub,
			GroupId:   groupExternalId,
			GroupName: subGroupName,
		})
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, sg := range groupSubgroupsReply.Groups {
			if sg.Name == subGroupName {
				return sg, nil
			}
		}
	}

	return nil, util.ErrCheck(errors.New("no subgroup found in get or create"))
}

func (k *Keycloak) AddRolesToGroup(ctx context.Context, userSub, id string, roles []*types.KeycloakRole) error {
	_, err := k.SendCommand(ctx, AddRolesToGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: id,
		Roles:   roles,
	})

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) AddUserToGroup(ctx context.Context, userSub, joiningUserId, groupId string) error {
	_, err := k.SendCommand(ctx, AddUserToGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: groupId,
		UserId:  joiningUserId,
	})

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (k *Keycloak) DeleteUserFromGroup(ctx context.Context, userSub, deletingUserId, groupId string) error {
	_, err := k.SendCommand(ctx, DeleteUserFromGroupKeycloakCommand, &types.AuthRequestParams{
		UserSub: userSub,
		GroupId: groupId,
		UserId:  deletingUserId,
	})

	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}
