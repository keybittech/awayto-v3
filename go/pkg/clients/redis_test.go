package clients

import (
	"context"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestInitRedis(t *testing.T) {
	tests := []struct {
		name string
		want interfaces.IRedis
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitRedis(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitRedis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParticipantTopicsKey(t *testing.T) {
	type args struct {
		topic string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParticipantTopicsKey(tt.args.topic)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParticipantTopicsKey(%v) error = %v, wantErr %v", tt.args.topic, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParticipantTopicsKey(%v) = %v, want %v", tt.args.topic, got, tt.want)
			}
		})
	}
}

func TestSocketIdTopicsKey(t *testing.T) {
	type args struct {
		socketId string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SocketIdTopicsKey(tt.args.socketId)
			if (err != nil) != tt.wantErr {
				t.Errorf("SocketIdTopicsKey(%v) error = %v, wantErr %v", tt.args.socketId, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SocketIdTopicsKey(%v) = %v, want %v", tt.args.socketId, got, tt.want)
			}
		})
	}
}

func TestRedis_Client(t *testing.T) {
	tests := []struct {
		name string
		r    *Redis
		want interfaces.IRedisClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Client(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Redis.Client() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedis_InitKeys(t *testing.T) {
	tests := []struct {
		name string
		r    *Redis
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.InitKeys()
		})
	}
}

func TestRedis_InitRedisSocketConnection(t *testing.T) {
	type args struct {
		socketId string
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
			if err := tt.r.InitRedisSocketConnection(tt.args.socketId); (err != nil) != tt.wantErr {
				t.Errorf("Redis.InitRedisSocketConnection(%v) error = %v, wantErr %v", tt.args.socketId, err, tt.wantErr)
			}
		})
	}
}

func TestRedis_HandleUnsub(t *testing.T) {
	type args struct {
		socketId string
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.HandleUnsub(tt.args.socketId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.HandleUnsub(%v) error = %v, wantErr %v", tt.args.socketId, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Redis.HandleUnsub(%v) = %v, want %v", tt.args.socketId, got, tt.want)
			}
		})
	}
}

func TestRedis_RemoveTopicFromConnection(t *testing.T) {
	type args struct {
		socketId string
		topic    string
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
			if err := tt.r.RemoveTopicFromConnection(tt.args.socketId, tt.args.topic); (err != nil) != tt.wantErr {
				t.Errorf("Redis.RemoveTopicFromConnection(%v, %v) error = %v, wantErr %v", tt.args.socketId, tt.args.topic, err, tt.wantErr)
			}
		})
	}
}

func TestRedis_GetCachedParticipants(t *testing.T) {
	type args struct {
		ctx         context.Context
		topic       string
		targetsOnly bool
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		want    map[string]*types.SocketParticipant
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.r.GetCachedParticipants(tt.args.ctx, tt.args.topic, tt.args.targetsOnly)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.GetCachedParticipants(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.topic, tt.args.targetsOnly, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Redis.GetCachedParticipants(%v, %v, %v) got = %v, want %v", tt.args.ctx, tt.args.topic, tt.args.targetsOnly, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Redis.GetCachedParticipants(%v, %v, %v) got1 = %v, want %v", tt.args.ctx, tt.args.topic, tt.args.targetsOnly, got1, tt.want1)
			}
		})
	}
}

func TestRedis_TrackTopicParticipant(t *testing.T) {
	type args struct {
		ctx      context.Context
		topic    string
		socketId string
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
			if err := tt.r.TrackTopicParticipant(tt.args.ctx, tt.args.topic, tt.args.socketId); (err != nil) != tt.wantErr {
				t.Errorf("Redis.TrackTopicParticipant(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.topic, tt.args.socketId, err, tt.wantErr)
			}
		})
	}
}

func TestRedis_HasTracking(t *testing.T) {
	type args struct {
		ctx      context.Context
		topic    string
		socketId string
	}
	tests := []struct {
		name    string
		r       *Redis
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.HasTracking(tt.args.ctx, tt.args.topic, tt.args.socketId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.HasTracking(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.topic, tt.args.socketId, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Redis.HasTracking(%v, %v, %v) = %v, want %v", tt.args.ctx, tt.args.topic, tt.args.socketId, got, tt.want)
			}
		})
	}
}
