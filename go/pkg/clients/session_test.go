package clients

import (
	"context"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestRedis_SetGroupSessionVersion(t *testing.T) {
	type args struct {
		ctx     context.Context
		groupId string
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.SetGroupSessionVersion(tt.args.ctx, tt.args.groupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.SetGroupSessionVersion(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.groupId, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Redis.SetGroupSessionVersion(%v, %v) = %v, want %v", tt.args.ctx, tt.args.groupId, got, tt.want)
			}
		})
	}
}

func TestRedis_GetGroupSessionVersion(t *testing.T) {
	type args struct {
		ctx     context.Context
		groupId string
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.GetGroupSessionVersion(tt.args.ctx, tt.args.groupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.GetGroupSessionVersion(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.groupId, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Redis.GetGroupSessionVersion(%v, %v) = %v, want %v", tt.args.ctx, tt.args.groupId, got, tt.want)
			}
		})
	}
}

func TestRedis_GetSession(t *testing.T) {
	type args struct {
		ctx     context.Context
		userSub string
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		want    *types.UserSession
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.GetSession(tt.args.ctx, tt.args.userSub)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.GetSession(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.userSub, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Redis.GetSession(%v, %v) = %v, want %v", tt.args.ctx, tt.args.userSub, got, tt.want)
			}
		})
	}
}

func TestRedis_SetSession(t *testing.T) {
	type args struct {
		ctx     context.Context
		session *types.UserSession
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.SetSession(tt.args.ctx, tt.args.session); (err != nil) != tt.wantErr {
				t.Errorf("Redis.SetSession(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.session, err, tt.wantErr)
			}
		})
	}
}

func TestRedis_DeleteSession(t *testing.T) {
	type args struct {
		ctx     context.Context
		userSub string
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.DeleteSession(tt.args.ctx, tt.args.userSub); (err != nil) != tt.wantErr {
				t.Errorf("Redis.DeleteSession(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.userSub, err, tt.wantErr)
			}
		})
	}
}
