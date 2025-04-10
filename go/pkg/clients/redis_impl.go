package clients

import (
	"context"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"

	"github.com/redis/go-redis/v9"
)

type RedisWrapper struct {
	*redis.Client
}

// Redis status cmd functions
type StatusCmdWrapper struct {
	*redis.StatusCmd
}

func (rc *RedisWrapper) Set(ctx context.Context, key string, value interface{}, duration time.Duration) interfaces.IRedisStatusCmd {
	return &StatusCmdWrapper{rc.Client.Set(ctx, key, value, duration)}
}

func (rc *RedisWrapper) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) interfaces.IRedisStatusCmd {
	return &StatusCmdWrapper{rc.Client.SetEx(ctx, key, value, expiration)}
}

func (statusCmd *StatusCmdWrapper) Err() error {
	return statusCmd.StatusCmd.Err()
}

// Redis bool cmd functions
type BoolCmdWrapper struct {
	*redis.BoolCmd
}

func (rc *RedisWrapper) Expire(ctx context.Context, key string, expiration time.Duration) interfaces.IRedisBoolCmd {
	return &BoolCmdWrapper{rc.Client.Expire(ctx, key, expiration)}
}

func (rc *RedisWrapper) SIsMember(ctx context.Context, key string, member interface{}) interfaces.IRedisBoolCmd {
	return &BoolCmdWrapper{rc.Client.SIsMember(ctx, key, member)}
}

func (boolCmd *BoolCmdWrapper) Result() (bool, error) {
	return boolCmd.BoolCmd.Result()
}

func (boolCmd *BoolCmdWrapper) Err() error {
	return boolCmd.BoolCmd.Err()
}

// Redis string cmd functions
type StringCmdWrapper struct {
	*redis.StringCmd
}

func (rc *RedisWrapper) Get(ctx context.Context, key string) interfaces.IRedisStringCmd {
	return &StringCmdWrapper{rc.Client.Get(ctx, key)}
}

func (stringCmd *StringCmdWrapper) Bytes() ([]byte, error) {
	return stringCmd.StringCmd.Bytes()
}

func (stringCmd *StringCmdWrapper) Int64() (int64, error) {
	return stringCmd.StringCmd.Int64()
}

func (stringCmd *StringCmdWrapper) Err() error {
	return stringCmd.StringCmd.Err()
}

// Redis int cmd functions
type IntCmdWrapper struct {
	*redis.IntCmd
}

func (rc *RedisWrapper) Del(ctx context.Context, key ...string) interfaces.IRedisIntCmd {
	return &IntCmdWrapper{rc.Client.Del(ctx, key...)}
}

func (rc *RedisWrapper) SAdd(ctx context.Context, key string, members ...interface{}) interfaces.IRedisIntCmd {
	return &IntCmdWrapper{rc.Client.SAdd(ctx, key, members...)}
}

func (rc *RedisWrapper) SRem(ctx context.Context, key string, members ...interface{}) interfaces.IRedisIntCmd {
	return &IntCmdWrapper{rc.Client.SRem(ctx, key, members...)}
}

func (intCmd *IntCmdWrapper) Result() (int64, error) {
	return intCmd.IntCmd.Result()
}

func (intCmd *IntCmdWrapper) Err() error {
	return intCmd.IntCmd.Err()
}

// Redis string slice cmd functions
type StringSliceCmdWrapper struct {
	*redis.StringSliceCmd
}

func (rc *RedisWrapper) SMembers(ctx context.Context, key string) interfaces.IRedisStringSliceCmd {
	return &StringSliceCmdWrapper{rc.Client.SMembers(ctx, key)}
}

func (stringSliceCmd *StringSliceCmdWrapper) Result() ([]string, error) {
	return stringSliceCmd.StringSliceCmd.Result()
}

func (stringSliceCmd *StringSliceCmdWrapper) Err() error {
	return stringSliceCmd.StringSliceCmd.Err()
}
