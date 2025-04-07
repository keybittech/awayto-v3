package clients

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	_ "github.com/lib/pq"
)

func TestInitKeycloak(t *testing.T) {
	tests := []struct {
		name string
		want interfaces.IKeycloak
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
			if err := tt.k.UpdateUser(tt.args.id, tt.args.firstName, tt.args.lastName); (err != nil) != tt.wantErr {
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
			if got, _ := tt.k.GetGroupAdminRoles(); !reflect.DeepEqual(got, tt.want) {
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
			if got, _ := tt.k.GetGroupSiteRoles(tt.args.groupId); !reflect.DeepEqual(got, tt.want) {
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
			got, err := tt.k.CreateGroup(tt.args.name)
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
			got, err := tt.k.GetGroup(tt.args.id)
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
			got, err := tt.k.GetGroupByName(tt.args.name)
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
			got, err := tt.k.GetGroupSubgroups(tt.args.groupId)
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
			if err := tt.k.DeleteGroup(tt.args.id); (err != nil) != tt.wantErr {
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
			if err := tt.k.UpdateGroup(tt.args.id, tt.args.name); (err != nil) != tt.wantErr {
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
			got, err := tt.k.CreateOrGetSubGroup(tt.args.groupExternalId, tt.args.subGroupName)
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
			if err := tt.k.AddRolesToGroup(tt.args.id, tt.args.roles); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.AddRolesToGroup(%v, %v) error = %v, wantErr %v", tt.args.id, tt.args.roles, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloak_DeleteRolesFromGroup(t *testing.T) {
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
			if err := tt.k.DeleteRolesFromGroup(tt.args.id, tt.args.roles); (err != nil) != tt.wantErr {
				t.Errorf("Keycloak.DeleteRolesFromGroup(%v, %v) error = %v, wantErr %v", tt.args.id, tt.args.roles, err, tt.wantErr)
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
			if err := tt.k.AddUserToGroup(tt.args.userId, tt.args.groupId); (err != nil) != tt.wantErr {
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
			if err := tt.k.DeleteUserFromGroup(tt.args.userId, tt.args.groupId); (err != nil) != tt.wantErr {
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
