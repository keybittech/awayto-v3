package clients

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"context"
	"database/sql"
	"net"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type IAi interface {
	GetPromptResponse(ctx context.Context, promptParts []string, promptType types.IPrompts) (string, error)
}

type IDatabase interface {
	Client() IDatabaseClient
	AdminSub() string
	AdminRoleId() string
	InitDBSocketConnection(tx IDatabaseTx, userSub string, connId string) (func(), error)
	GetSocketAllowances(tx IDatabaseTx, userSub string) ([]util.IdStruct, error)
	GetTopicMessageParticipants(tx IDatabaseTx, topic string) (SocketParticipants, error)
	GetSocketParticipantDetails(tx IDatabaseTx, participants SocketParticipants) (SocketParticipants, error)
	StoreTopicMessage(tx IDatabaseTx, connId string, message SocketMessage) error
	GetTopicMessages(tx IDatabaseTx, topic string, page, pageSize int) ([][]byte, error)
	QueryRows(protoStructSlice interface{}, query string, args ...interface{}) error
	TxExec(doFunc func(IDatabaseTx) error, ids ...string) error
}

type IDatabaseClient interface {
	Conn(ctx context.Context) (*sql.Conn, error)
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (IRows, error)
	QueryRow(query string, args ...any) IRow
	Begin() (IDatabaseTx, error)
}

type IDatabaseTx interface {
	Commit() error
	Rollback() error
	Exec(query string, args ...any) (sql.Result, error)
	SetDbVar(string, string) error
	Query(query string, args ...any) (IRows, error)
	QueryRow(query string, args ...any) IRow
	QueryRows(protoStructSlice interface{}, query string, args ...interface{}) error
}

type IRow interface {
	Scan(dest ...interface{}) error
}

type IRows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
}

type IRedis interface {
	Client() IRedisClient
	InitKeys()
	InitRedisSocketConnection(socketId string) error
	HandleUnsub(socketId string) (map[string][]string, error)
	RemoveTopicFromConnection(socketId, topic string) error
	GetCachedParticipants(ctx context.Context, topic string) SocketParticipants
	GetParticipantTargets(participants SocketParticipants) []string
	TrackTopicParticipant(ctx context.Context, topic, socketId string)
	GetGroupSessionVersion(ctx context.Context, groupId string) (int64, error)
	SetGroupSessionVersion(ctx context.Context, groupId string) (int64, error)
	GetSession(ctx context.Context, userSub string) (*UserSession, error)
	SetSession(ctx context.Context, session *UserSession) error
	DeleteSession(ctx context.Context, userSub string) error
}

type IRedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}

type IKeycloak interface {
	Chan() chan<- KeycloakCommand
	Client() *KeycloakClient
	UpdateUser(id, firstName, lastName string) error
	GetUserTokenValid(token string) (*KeycloakUser, error)
	GetUserInfoById(id string) (*KeycloakUser, error)
	GetGroupAdminRoles() []KeycloakRole
	GetGroupSiteRoles(groupId string) []ClientRoleMappingRole
	CreateOrGetSubGroup(groupExternalId, subGroupName string) (*KeycloakGroup, error)
	CreateGroup(name string) (*KeycloakGroup, error)
	GetGroup(id string) (*KeycloakGroup, error)
	GetGroupByName(name string) (*[]KeycloakGroup, error)
	GetGroupSubgroups(groupId string) (*[]KeycloakGroup, error)
	DeleteGroup(id string) error
	UpdateGroup(id, name string) error
	AddRolesToGroup(id string, roles []KeycloakRole) error
	DeleteRolesFromGroup(id string, roles []KeycloakRole) error
	AddUserToGroup(userId, groupId string) error
	DeleteUserFromGroup(userId, groupId string) error
}

type ISocket interface {
	Chan() chan<- SocketCommand
	InitConnection(conn net.Conn, userSub string, ticket string) (func(), error)
	GetSocketTicket(session *UserSession) (string, error)
	SendMessageBytes(targets []string, messageBytes []byte) error
	SendMessage(targets []string, message SocketMessage) error
	GetSubscriberByTicket(ticket string) (*Subscriber, error)
	AddSubscribedTopic(userSub, topic string, existingCids []string)
	GetSubscribedTopicTargets(userSub, topic string) []string
	DeleteSubscribedTopic(userSub, topic string)
	HasTopicSubscription(userSub, topic string) bool
	NotifyTopicUnsub(topic, socketId string, targets []string)
	RoleCall(userSub string) error
}
