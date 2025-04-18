package clients

import (
	"database/sql"
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
			if err := tt.db.InitDbSocketConnection(tt.args.connId, tt.args.userSub, tt.args.groupId, tt.args.roles); (err != nil) != tt.wantErr {
				t.Errorf("Database.InitDbSocketConnection(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.connId, tt.args.userSub, tt.args.groupId, tt.args.roles, err, tt.wantErr)
			}
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
			if err := tt.db.RemoveDbSocketConnection(tt.args.connId); (err != nil) != tt.wantErr {
				t.Errorf("Database.RemoveDbSocketConnection(%v) error = %v, wantErr %v", tt.args.connId, err, tt.wantErr)
			}
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
			got, err := tt.db.GetSocketAllowances(tt.args.session, tt.args.handle)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetSocketAllowances(%v, %v) error = %v, wantErr %v", tt.args.session, tt.args.handle, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Database.GetSocketAllowances(%v, %v) = %v, want %v", tt.args.session, tt.args.handle, got, tt.want)
			}
		})
	}
}

func TestDatabase_GetTopicMessageParticipants(t *testing.T) {
	type args struct {
		tx           *sql.Tx
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
			if err := tt.db.GetTopicMessageParticipants(tt.args.tx, tt.args.topic, tt.args.participants); (err != nil) != tt.wantErr {
				t.Errorf("Database.GetTopicMessageParticipants(%v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.topic, tt.args.participants, err, tt.wantErr)
			}
		})
	}
}

func TestDatabase_GetSocketParticipantDetails(t *testing.T) {
	type args struct {
		tx           *sql.Tx
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
			if err := tt.db.GetSocketParticipantDetails(tt.args.tx, tt.args.participants); (err != nil) != tt.wantErr {
				t.Errorf("Database.GetSocketParticipantDetails(%v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.participants, err, tt.wantErr)
			}
		})
	}
}

func TestDatabase_StoreTopicMessage(t *testing.T) {
	type args struct {
		tx      *sql.Tx
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
			if err := tt.db.StoreTopicMessage(tt.args.tx, tt.args.connId, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Database.StoreTopicMessage(%v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.connId, tt.args.message, err, tt.wantErr)
			}
		})
	}
}

func TestDatabase_GetTopicMessages(t *testing.T) {
	type args struct {
		tx       *sql.Tx
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
			got, err := tt.db.GetTopicMessages(tt.args.tx, tt.args.topic, tt.args.page, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetTopicMessages(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.topic, tt.args.page, tt.args.pageSize, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Database.GetTopicMessages(%v, %v, %v, %v) = %v, want %v", tt.args.tx, tt.args.topic, tt.args.page, tt.args.pageSize, got, tt.want)
			}
		})
	}
}
