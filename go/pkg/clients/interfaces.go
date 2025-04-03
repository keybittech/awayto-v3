package clients

import (
	"context"
	"database/sql"
	"net"
	"strings"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"

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
	GetSocketAllowances(tx IDatabaseTx, userSub, description, handle string) (bool, error)
	GetTopicMessageParticipants(tx IDatabaseTx, topic string) (map[string]*types.SocketParticipant, error)
	GetSocketParticipantDetails(tx IDatabaseTx, participants map[string]*types.SocketParticipant) (map[string]*types.SocketParticipant, error)
	StoreTopicMessage(tx IDatabaseTx, connId string, message *types.SocketMessage) error
	GetTopicMessages(tx IDatabaseTx, topic string, page, pageSize int) ([][]byte, error)
	QueryRows(protoStructSlice interface{}, query string, args ...interface{}) error
	TxExec(doFunc func(IDatabaseTx) error, ids ...string) error
	BuildInserts(sb *strings.Builder, size, current int) error
}

type IDatabaseClient interface {
	Conn(ctx context.Context) (*sql.Conn, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (IRows, error)
	QueryRow(query string, args ...any) IRow
	Begin() (IDatabaseTx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (IDatabaseTx, error)
}

type IDatabaseTx interface {
	Prepare(stmt string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Commit() error
	Rollback() error
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
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
	HandleUnsub(socketId string) (map[string]string, error)
	RemoveTopicFromConnection(socketId, topic string) error
	GetCachedParticipants(ctx context.Context, topic string, targetsOnly bool) (map[string]*types.SocketParticipant, string, error)
	TrackTopicParticipant(ctx context.Context, topic, socketId string) error
	GetGroupSessionVersion(ctx context.Context, groupId string) (int64, error)
	SetGroupSessionVersion(ctx context.Context, groupId string) (int64, error)
	GetSession(ctx context.Context, userSub string) (*types.UserSession, error)
	SetSession(ctx context.Context, session *types.UserSession) error
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
	GetGroupAdminRoles() []*types.KeycloakRole
	GetGroupSiteRoles(groupId string) []*types.ClientRoleMappingRole
	CreateOrGetSubGroup(groupExternalId, subGroupName string) (*types.KeycloakGroup, error)
	CreateGroup(name string) (*types.KeycloakGroup, error)
	GetGroup(id string) (*types.KeycloakGroup, error)
	GetGroupByName(name string) ([]*types.KeycloakGroup, error)
	GetGroupSubgroups(groupId string) ([]*types.KeycloakGroup, error)
	DeleteGroup(id string) error
	UpdateGroup(id, name string) error
	AddRolesToGroup(id string, roles []*types.KeycloakRole) error
	DeleteRolesFromGroup(id string, roles []*types.KeycloakRole) error
	AddUserToGroup(userId, groupId string) error
	DeleteUserFromGroup(userId, groupId string) error
}

type ISocket interface {
	Chan() chan<- SocketCommand
	InitConnection(conn net.Conn, userSub string, ticket string) (func(), error)
	GetSocketTicket(session *types.UserSession) (string, error)
	SendMessageBytes(messageBytes []byte, targets string) error
	SendMessage(message *types.SocketMessage, targets string) error
	GetSubscriberByTicket(ticket string) (*types.Subscriber, error)
	AddSubscribedTopic(userSub, topic string, targets string)
	DeleteSubscribedTopic(userSub, topic string)
	HasTopicSubscription(userSub, topic string) bool
	RoleCall(userSub string) error
}
