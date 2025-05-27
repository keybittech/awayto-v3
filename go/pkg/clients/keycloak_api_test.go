package clients

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestKeycloakClient_BasicHeaders(t *testing.T) {
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		want           http.Header
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.keycloakClient.BasicHeaders(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.BasicHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_DirectGrantAuthentication(t *testing.T) {
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		want           *types.OIDCToken
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.DirectGrantAuthentication()
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.DirectGrantAuthentication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.DirectGrantAuthentication() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_FindResource(t *testing.T) {
	type args struct {
		resource string
		search   string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		want           []byte
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.FindResource(tt.args.resource, tt.args.search, 0, 10)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.FindResource(%v, %v) error = %v, wantErr %v", tt.args.resource, tt.args.search, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.FindResource(%v, %v) = %v, want %v", tt.args.resource, tt.args.search, got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_GetUserListInRealm(t *testing.T) {
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		want           []*types.KeycloakUser
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.GetUserListInRealm()
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.GetUserListInRealm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.GetUserListInRealm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_GetUserGroups(t *testing.T) {
	type args struct {
		userId string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		want           []*types.KeycloakUserGroup
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.GetUserGroups(tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.GetUserGroups(%v) error = %v, wantErr %v", tt.args.userId, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.GetUserGroups(%v) = %v, want %v", tt.args.userId, got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_GetGroupRoleMappings(t *testing.T) {
	type args struct {
		groupId string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		want           []*types.ClientRoleMappingRole
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.GetGroupRoleMappings(tt.args.groupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.GetGroupRoleMappings(%v) error = %v, wantErr %v", tt.args.groupId, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.GetGroupRoleMappings(%v) = %v, want %v", tt.args.groupId, got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_GetAppClientRoles(t *testing.T) {
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		want           []*types.KeycloakRole
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.GetAppClientRoles()
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.GetAppClientRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.GetAppClientRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_GetRealmClients(t *testing.T) {
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		want           []*types.KeycloakRealmClient
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.GetRealmClients()
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.GetRealmClients() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.GetRealmClients() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_UpdateUser(t *testing.T) {
	type args struct {
		userId    string
		firstName string
		lastName  string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.keycloakClient.UpdateUser(tt.args.userId, tt.args.firstName, tt.args.lastName); (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.UpdateUser(%v, %v, %v) error = %v, wantErr %v", tt.args.userId, tt.args.firstName, tt.args.lastName, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloakClient_CreateGroup(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		want           string
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.CreateGroup(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.CreateGroup(%v) error = %v, wantErr %v", tt.args.name, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KeycloakClient.CreateGroup(%v) = %v, want %v", tt.args.name, got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_DeleteGroup(t *testing.T) {
	type args struct {
		groupId string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.keycloakClient.DeleteGroup(tt.args.groupId); (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.DeleteGroup(%v) error = %v, wantErr %v", tt.args.groupId, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloakClient_UpdateGroup(t *testing.T) {
	type args struct {
		groupId   string
		groupName string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.keycloakClient.UpdateGroup(tt.args.groupId, tt.args.groupName); (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.UpdateGroup(%v, %v) error = %v, wantErr %v", tt.args.groupId, tt.args.groupName, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloakClient_GetGroup(t *testing.T) {
	type args struct {
		groupId string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		want           *types.KeycloakGroup
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.GetGroup(tt.args.groupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.GetGroup(%v) error = %v, wantErr %v", tt.args.groupId, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.GetGroup(%v) = %v, want %v", tt.args.groupId, got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_CreateSubgroup(t *testing.T) {
	type args struct {
		groupId      string
		subgroupName string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		want           *types.KeycloakGroup
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.CreateSubgroup(tt.args.groupId, tt.args.subgroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.CreateSubgroup(%v, %v) error = %v, wantErr %v", tt.args.groupId, tt.args.subgroupName, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.CreateSubgroup(%v, %v) = %v, want %v", tt.args.groupId, tt.args.subgroupName, got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_GetGroupSubgroups(t *testing.T) {
	type args struct {
		groupId string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		want           []*types.KeycloakGroup
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keycloakClient.GetGroupSubgroups(tt.args.groupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.GetGroupSubgroups(%v) error = %v, wantErr %v", tt.args.groupId, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeycloakClient.GetGroupSubgroups(%v) = %v, want %v", tt.args.groupId, got, tt.want)
			}
		})
	}
}

func TestKeycloakClient_MutateGroupRoles(t *testing.T) {
	type args struct {
		method  string
		groupId string
		roles   []*types.KeycloakRole
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.keycloakClient.MutateGroupRoles(tt.args.method, tt.args.groupId, tt.args.roles); (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.MutateGroupRoles(%v, %v, %v) error = %v, wantErr %v", tt.args.method, tt.args.groupId, tt.args.roles, err, tt.wantErr)
			}
		})
	}
}

func TestKeycloakClient_MutateUserGroupMembership(t *testing.T) {
	type args struct {
		method  string
		userId  string
		groupId string
	}
	tests := []struct {
		name           string
		keycloakClient KeycloakClient
		args           args
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.keycloakClient.MutateUserGroupMembership(tt.args.method, tt.args.userId, tt.args.groupId); (err != nil) != tt.wantErr {
				t.Errorf("KeycloakClient.MutateUserGroupMembership(%v, %v, %v) error = %v, wantErr %v", tt.args.method, tt.args.userId, tt.args.groupId, err, tt.wantErr)
			}
		})
	}
}
