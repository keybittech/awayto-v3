package clients

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestDatabase_InitDBSocketConnection(t *testing.T) {
	type args struct {
		tx      interfaces.IDatabaseTx
		userSub string
		connId  string
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		want    func()
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.InitDBSocketConnection(tt.args.tx, tt.args.userSub, tt.args.connId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.InitDBSocketConnection(%v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.userSub, tt.args.connId, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Database.InitDBSocketConnection(%v, %v, %v) = %v, want %v", tt.args.tx, tt.args.userSub, tt.args.connId, got, tt.want)
			}
		})
	}
}

func TestDatabase_GetSocketAllowances(t *testing.T) {
	type args struct {
		tx          interfaces.IDatabaseTx
		userSub     string
		description string
		handle      string
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
			got, err := tt.db.GetSocketAllowances(tt.args.tx, tt.args.userSub, tt.args.description, tt.args.handle)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetSocketAllowances(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.userSub, tt.args.description, tt.args.handle, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Database.GetSocketAllowances(%v, %v, %v, %v) = %v, want %v", tt.args.tx, tt.args.userSub, tt.args.description, tt.args.handle, got, tt.want)
			}
		})
	}
}

func TestDatabase_GetTopicMessageParticipants(t *testing.T) {
	type args struct {
		tx    interfaces.IDatabaseTx
		topic string
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		want    map[string]*types.SocketParticipant
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.GetTopicMessageParticipants(tt.args.tx, tt.args.topic)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetTopicMessageParticipants(%v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.topic, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Database.GetTopicMessageParticipants(%v, %v) = %v, want %v", tt.args.tx, tt.args.topic, got, tt.want)
			}
		})
	}
}

func TestDatabase_GetSocketParticipantDetails(t *testing.T) {
	type args struct {
		tx           interfaces.IDatabaseTx
		participants map[string]*types.SocketParticipant
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		want    map[string]*types.SocketParticipant
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.GetSocketParticipantDetails(tt.args.tx, tt.args.participants)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetSocketParticipantDetails(%v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.participants, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Database.GetSocketParticipantDetails(%v, %v) = %v, want %v", tt.args.tx, tt.args.participants, got, tt.want)
			}
		})
	}
}

func TestDatabase_StoreTopicMessage(t *testing.T) {
	type args struct {
		tx      interfaces.IDatabaseTx
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
		tx       interfaces.IDatabaseTx
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
