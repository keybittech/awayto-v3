package clients

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	_ "github.com/lib/pq"
)

func TestKeycloak_UpdateUser(t *testing.T) {
	type args struct {
		id        string
		firstName string
		lastName  string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.k.UpdateUser("test", tt.args.id, tt.args.firstName, tt.args.lastName); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.UpdateUser(%v, %v, %v) error = %v, wantErr %v", tt.args.id, tt.args.firstName, tt.args.lastName, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloak_GetGroupAdminRoles(t *testing.T) {
	tests := []struct {
		name string
		k    *Keycloak
		want []*types.KeycloakRole
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.k.GetGroupAdminRoles("test"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.GetGroupAdminRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeycloak_GetGroupSiteRoles(t *testing.T) {
	type args struct {
		groupId string
	}
	tests := []struct {
		name string
		k    *Keycloak
		args args
		want []*types.ClientRoleMappingRole
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.k.GetGroupSiteRoles("test", tt.args.groupId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.GetGroupSiteRoles(%v) = %v, want %v", tt.args.groupId, got, tt.want)
			}
		})
	}
}

func TestKeycloak_CreateGroup(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		want    *types.KeycloakGroup
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.CreateGroup("test", tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.CreateGroup(%v) error = %v, wantErr %v", tt.args.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.CreateGroup(%v) = %v, want %v", tt.args.name, got, tt.want)
			}
		})
	}
}

func TestKeycloak_GetGroup(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		want    *types.KeycloakGroup
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.GetGroup("test", tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.GetGroup(%v) error = %v, wantErr %v", tt.args.id, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.GetGroup(%v) = %v, want %v", tt.args.id, got, tt.want)
			}
		})
	}
}

func TestKeycloak_GetGroupByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		want    []*types.KeycloakGroup
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.GetGroupByName("test", tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.GetGroupByName(%v) error = %v, wantErr %v", tt.args.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.GetGroupByName(%v) = %v, want %v", tt.args.name, got, tt.want)
			}
		})
	}
}

func TestKeycloak_GetGroupSubgroups(t *testing.T) {
	type args struct {
		groupId string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		want    []*types.KeycloakGroup
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.GetGroupSubgroups("test", tt.args.groupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.GetGroupSubgroups(%v) error = %v, wantErr %v", tt.args.groupId, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.GetGroupSubgroups(%v) = %v, want %v", tt.args.groupId, got, tt.want)
			}
		})
	}
}

func TestKeycloak_DeleteGroup(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.k.DeleteGroup("test", tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.DeleteGroup(%v) error = %v, wantErr %v", tt.args.id, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloak_UpdateGroup(t *testing.T) {
	type args struct {
		id   string
		name string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.k.UpdateGroup("test", tt.args.id, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.UpdateGroup(%v, %v) error = %v, wantErr %v", tt.args.id, tt.args.name, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloak_CreateOrGetSubGroup(t *testing.T) {
	type args struct {
		groupExternalId string
		subGroupName    string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		want    *types.KeycloakGroup
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.CreateOrGetSubGroup("test", tt.args.groupExternalId, tt.args.subGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.CreateOrGetSubGroup(%v, %v) error = %v, wantErr %v", tt.args.groupExternalId, tt.args.subGroupName, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.CreateOrGetSubGroup(%v, %v) = %v, want %v", tt.args.groupExternalId, tt.args.subGroupName, got, tt.want)
			}
		})
	}
}

func TestKeycloak_AddRolesToGroup(t *testing.T) {
	type args struct {
		id    string
		roles []*types.KeycloakRole
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.k.AddRolesToGroup("test", tt.args.id, tt.args.roles); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.AddRolesToGroup(%v, %v) error = %v, wantErr %v", tt.args.id, tt.args.roles, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloak_AddUserToGroup(t *testing.T) {
	type args struct {
		userId  string
		groupId string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.k.AddUserToGroup("test", tt.args.userId, tt.args.groupId); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.AddUserToGroup(%v, %v) error = %v, wantErr %v", tt.args.userId, tt.args.groupId, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloak_DeleteUserFromGroup(t *testing.T) {
	type args struct {
		userId  string
		groupId string
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.k.DeleteUserFromGroup("test", tt.args.userId, tt.args.groupId); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.DeleteUserFromGroup(%v, %v) error = %v, wantErr %v", tt.args.userId, tt.args.groupId, err, tt.wantErr)
			}
		})
	}
}

// func TestKeycloak_SendCommand(t *testing.T) {
// 	type fields struct {
// 		C  *KeycloakClient
// 		handlerId string
// 		Ch chan<- KeycloakCommand
// 	}
// 	type args struct {
// 		cmdType KeycloakCommandType
// 		params  KeycloakParams
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    KeycloakResponse
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			k := &Keycloak{
// 				C:  tt.fields.C,
// 			}
// 			got, err := k.SendCommand(tt.args.cmdType, tt.args.params)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Keycloak.SendCommand(%v, %v) error = %v, wantErr %v", tt.args.cmdType, tt.args.params, err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Keycloak.SendCommand(%v, %v) = %v, want %v", tt.args.cmdType, tt.args.params, got, tt.want)
// 			}
// 		})
// 	}
// }

func TestAuthCommand_GetClientId(t *testing.T) {
	tests := []struct {
		name string
		cmd  AuthCommand
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.GetClientId(); got != tt.want {
				t.Errorf("AuthCommand.GetClientId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthCommand_GetReplyChannel(t *testing.T) {
	tests := []struct {
		name string
		cmd  AuthCommand
		want interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.GetReplyChannel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthCommand.GetReplyChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitKeycloak(t *testing.T) {
	tests := []struct {
		name string
		want *Keycloak
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitKeycloak(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitKeycloak() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeycloak_RouteCommand(t *testing.T) {
	type args struct {
		cmd AuthCommand
	}
	tests := []struct {
		name    string
		s       *Keycloak
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.RouteCommand(tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.RouteCommand(%v) error = %v, wantErr %v", tt.args.cmd, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloak_Close(t *testing.T) {
	tests := []struct {
		name string
		s    *Keycloak
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Close()
		})
	}
}

func TestKeycloak_SendCommand(t *testing.T) {
	type args struct {
		cmdType int32
		request *types.AuthRequestParams
	}
	tests := []struct {
		name    string
		k       *Keycloak
		args    args
		want    AuthResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.SendCommand(tt.args.cmdType, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.SendCommand(%v, %v) error = %v, wantErr %v", tt.args.cmdType, tt.args.request, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keycloak.SendCommand(%v, %v) = %v, want %v", tt.args.cmdType, tt.args.request, got, tt.want)
			}
		})
	}
}
