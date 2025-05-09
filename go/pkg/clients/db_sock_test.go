package clients

import (
	"context"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestDatabase_InitDbSocketConnection(t *testing.T) {
	type args struct {
		connId  string
		userSub string
		groupId string
		roles   string
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := tt.db.InitDbSocketConnection(tt.args.connId, tt.args.userSub, tt.args.groupId, tt.args.roles); (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.InitDbSocketConnection(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.connId, tt.args.userSub, tt.args.groupId, tt.args.roles, err, tt.wantErr)
			// }
		})
	}
}

func TestDatabase_RemoveDbSocketConnection(t *testing.T) {
	type args struct {
		connId string
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := tt.db.RemoveDbSocketConnection(tt.args.connId); (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.RemoveDbSocketConnection(%v) error = %v, wantErr %v", tt.args.connId, err, tt.wantErr)
			// }
		})
	}
}

func TestDatabase_GetSocketAllowances(t *testing.T) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	type args struct {
		session *types.UserSession
		handle  string
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, err := tt.db.GetSocketAllowances(tt.args.session, tt.args.handle)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.GetSocketAllowances(%v, %v) error = %v, wantErr %v", tt.args.session, tt.args.handle, err, tt.wantErr)
			// 	return
			// }
			// if got != tt.want {
			// 	t.Errorf("Database.GetSocketAllowances(%v, %v) = %v, want %v", tt.args.session, tt.args.handle, got, tt.want)
			// }
		})
	}
}

func TestDatabase_GetTopicMessageParticipants(t *testing.T) {
	type args struct {
		tx           *PoolTx
		topic        string
		participants map[string]*types.SocketParticipant
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := tt.db.GetTopicMessageParticipants(tt.args.tx, tt.args.topic, tt.args.participants); (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.GetTopicMessageParticipants(%v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.topic, tt.args.participants, err, tt.wantErr)
			// }
		})
	}
}

func TestDatabase_GetSocketParticipantDetails(t *testing.T) {
	type args struct {
		tx           *PoolTx
		participants map[string]*types.SocketParticipant
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := tt.db.GetSocketParticipantDetails(tt.args.tx, tt.args.participants); (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.GetSocketParticipantDetails(%v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.participants, err, tt.wantErr)
			// }
		})
	}
}

func TestDatabase_StoreTopicMessage(t *testing.T) {
	type args struct {
		tx      *PoolTx
		connId  string
		message *types.SocketMessage
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := tt.db.StoreTopicMessage(tt.args.tx, tt.args.connId, tt.args.message); (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.StoreTopicMessage(%v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.connId, tt.args.message, err, tt.wantErr)
			// }
		})
	}
}

func TestDatabase_GetTopicMessages(t *testing.T) {
	type args struct {
		tx       *PoolTx
		topic    string
		page     int
		pageSize int
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		want    [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, err := tt.db.GetTopicMessages(tt.args.tx, tt.args.topic, tt.args.page, tt.args.pageSize)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.GetTopicMessages(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.topic, tt.args.page, tt.args.pageSize, err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Database.GetTopicMessages(%v, %v, %v, %v) = %v, want %v", tt.args.tx, tt.args.topic, tt.args.page, tt.args.pageSize, got, tt.want)
			// }
		})
	}
}

func TestDbSession_InitDbSocketConnection(t *testing.T) {
	type args struct {
		ctx    context.Context
		connId string
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ds.InitDbSocketConnection(tt.args.ctx, tt.args.connId); (err != nil) != tt.wantErr {
				t.Errorf("DbSession.InitDbSocketConnection(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.connId, err, tt.wantErr)
			}
		})
	}
}

func TestDbSession_RemoveDbSocketConnection(t *testing.T) {
	type args struct {
		ctx    context.Context
		connId string
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ds.RemoveDbSocketConnection(tt.args.ctx, tt.args.connId); (err != nil) != tt.wantErr {
				t.Errorf("DbSession.RemoveDbSocketConnection(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.connId, err, tt.wantErr)
			}
		})
	}
}

func TestDbSession_GetSocketAllowances(t *testing.T) {
	type args struct {
		ctx       context.Context
		bookingId string
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ds.GetSocketAllowances(tt.args.ctx, tt.args.bookingId)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbSession.GetSocketAllowances(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.bookingId, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DbSession.GetSocketAllowances(%v, %v) = %v, want %v", tt.args.ctx, tt.args.bookingId, got, tt.want)
			}
		})
	}
}

func TestDbSession_GetTopicMessageParticipants(t *testing.T) {
	type args struct {
		ctx          context.Context
		participants map[string]*types.SocketParticipant
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ds.GetTopicMessageParticipants(tt.args.ctx, tt.args.participants); (err != nil) != tt.wantErr {
				t.Errorf("DbSession.GetTopicMessageParticipants(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.participants, err, tt.wantErr)
			}
		})
	}
}

func TestDbSession_GetSocketParticipantDetails(t *testing.T) {
	type args struct {
		ctx          context.Context
		participants map[string]*types.SocketParticipant
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ds.GetSocketParticipantDetails(tt.args.ctx, tt.args.participants); (err != nil) != tt.wantErr {
				t.Errorf("DbSession.GetSocketParticipantDetails(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.participants, err, tt.wantErr)
			}
		})
	}
}

func TestDbSession_StoreTopicMessage(t *testing.T) {
	type args struct {
		ctx     context.Context
		connId  string
		message *types.SocketMessage
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ds.StoreTopicMessage(tt.args.ctx, tt.args.connId, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("DbSession.StoreTopicMessage(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.connId, tt.args.message, err, tt.wantErr)
			}
		})
	}
}

func TestDbSession_GetTopicMessages(t *testing.T) {
	type args struct {
		ctx      context.Context
		page     int
		pageSize int
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		want    [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ds.GetTopicMessages(tt.args.ctx, tt.args.page, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbSession.GetTopicMessages(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.page, tt.args.pageSize, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DbSession.GetTopicMessages(%v, %v, %v) = %v, want %v", tt.args.ctx, tt.args.page, tt.args.pageSize, got, tt.want)
			}
		})
	}
}
