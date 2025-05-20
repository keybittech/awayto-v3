package clients

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	_ "github.com/lib/pq"
)

const (
	emptyString            = ""
	emptyInteger           = 0
	setSessionVariablesSQL = `SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::INTEGER, $4::VARCHAR)`
)

type Database struct {
	DatabaseAdminSub    string
	DatabaseAdminRoleId string
	DatabaseClient      *DatabaseClient
}

func (db *Database) Client() *DatabaseClient {
	return db.DatabaseClient
}

func (db *Database) AdminSub() string {
	return db.DatabaseAdminSub
}

func (db *Database) AdminRoleId() string {
	return db.DatabaseAdminRoleId
}

func InitDatabase() *Database {
	dbDriver := os.Getenv("DB_DRIVER")
	pgUser := os.Getenv("PG_WORKER")
	pgDb := os.Getenv("PG_DB")
	pgPass, err := util.GetEnvFile("PG_WORKER_PASS_FILE", 128)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	connString := fmt.Sprintf("%s://%s:%s@/%s?host=%s&sslmode=disable", dbDriver, pgUser, pgPass, pgDb, os.Getenv("UNIX_SOCK_DIR"))
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse db config: %v\n", err)
	}

	config.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		util.RegisterTimestamp(c.TypeMap())
		util.RegisterDate(c.TypeMap())
		util.RegisterInterval(c.TypeMap())
		return nil
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	// go func() {
	// 	ticker := time.NewTicker(time.Duration(5 * time.Second))
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		stats := dbpool.Stat()
	// 		fmt.Printf("[%s] DB Stats: TotalConns: %d, AcquiredConns: %d, IdleConns: %d, MaxConns: %d, AcquireCount: %d, AcquireDuration: %s\n",
	// 			time.Now().Format(time.RFC3339),
	// 			stats.TotalConns(),
	// 			stats.AcquiredConns(),
	// 			stats.IdleConns(),
	// 			stats.MaxConns(),
	// 			stats.AcquireCount(),
	// 			stats.AcquireDuration(),
	// 		)
	// 	}
	// }()

	dbc := &Database{
		DatabaseClient: &DatabaseClient{
			Pool: dbpool,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	workerDbSession := &DbSession{
		Pool: dbpool,
		ConcurrentUserSession: types.NewConcurrentUserSession(&types.UserSession{
			UserSub: "worker",
		}),
	}

	row, done, err := workerDbSession.SessionBatchQueryRow(ctx, `
		SELECT u.sub, r.id FROM dbtable_schema.users u
		JOIN dbtable_schema.roles r ON r.name = 'Admin'
		WHERE u.username = 'system_owner'
	`)
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	err = row.Scan(&dbc.DatabaseAdminSub, &dbc.DatabaseAdminRoleId)
	if err != nil {
		done()
		log.Fatal(util.ErrCheck(err))
	}
	done()

	_, err = workerDbSession.SessionBatchExec(ctx, `
		DELETE FROM dbtable_schema.sock_connections
		USING dbtable_schema.sock_connections sc
		LEFT OUTER JOIN dbtable_schema.topic_messages tm ON tm.connection_id = sc.connection_id
		WHERE dbtable_schema.sock_connections.id = sc.id AND tm.id IS NULL
	`)
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	util.DebugLog.Println(fmt.Sprintf("Database Init\nAdmin Sub: %s\nAdmin Role Id: %s", dbc.AdminSub(), dbc.AdminRoleId()))

	return dbc
}

// Go code
func (db *Database) BuildInserts(sb *strings.Builder, size, current int) {

	baseIndex := (current / size) * size

	sb.WriteString("(")

	for i := range size {
		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(baseIndex + i + 1))
		if i < size-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("),")
}

type ColTypes struct {
	reflectString    reflect.Type
	reflectInt32     reflect.Type
	reflectInt64     reflect.Type
	reflectFloat64   reflect.Type
	reflectBool      reflect.Type
	reflectJson      reflect.Type
	reflectTimestamp reflect.Type
}

type DatabaseClient struct {
	*pgxpool.Pool
}

func (dc *DatabaseClient) OpenPoolSessionGroupTx(ctx context.Context, session *types.ConcurrentUserSession) (*PoolTx, *types.ConcurrentUserSession, error) {
	groupSession := types.NewConcurrentUserSession(&types.UserSession{
		UserSub: session.GetGroupSub(),
		GroupId: session.GetGroupId(),
	})
	groupPoolTx, err := dc.OpenPoolSessionTx(ctx, groupSession)
	return groupPoolTx, groupSession, err
}

func (dc *DatabaseClient) OpenPoolSessionTx(ctx context.Context, session *types.ConcurrentUserSession) (*PoolTx, error) {
	tx, err := dc.Pool.Begin(ctx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	poolTx := &PoolTx{
		Tx: tx,
	}

	err = poolTx.SetSession(ctx, session)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return poolTx, nil
}

func (dc *DatabaseClient) ClosePoolSessionTx(ctx context.Context, poolTx *PoolTx) error {
	err := poolTx.UnsetSession(ctx)
	if err != nil {
		return util.ErrCheck(err)
	}

	err = poolTx.Commit(ctx)
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

type PoolTx struct {
	pgx.Tx
}

func (ptx *PoolTx) SetSession(ctx context.Context, session *types.ConcurrentUserSession) error {
	_, err := ptx.Exec(ctx, setSessionVariablesSQL, session.GetUserSub(), session.GetGroupId(), session.GetRoleBits(), emptyString)
	if err != nil {
		return util.ErrCheck(err)
	}
	return nil
}

func (ptx *PoolTx) UnsetSession(ctx context.Context) error {
	_, err := ptx.Exec(ctx, setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)
	if err != nil {
		return util.ErrCheck(err)
	}
	return nil
}

type DbSession struct {
	Topic string
	*types.ConcurrentUserSession
	*pgxpool.Pool
}

func NewGroupDbSession(pool *pgxpool.Pool, session *types.ConcurrentUserSession) DbSession {
	return DbSession{
		Pool: pool,
		ConcurrentUserSession: types.NewConcurrentUserSession(&types.UserSession{
			UserSub: session.GetGroupSub(),
			GroupId: session.GetGroupId(),
		}),
	}
}

func (ds DbSession) SessionBatch(ctx context.Context, primaryQuery string, params ...any) pgx.BatchResults {
	batch := &pgx.Batch{}

	batch.Queue(setSessionVariablesSQL, ds.ConcurrentUserSession.GetUserSub(), ds.ConcurrentUserSession.GetGroupId(), ds.ConcurrentUserSession.GetRoleBits(), ds.Topic)
	batch.Queue(primaryQuery, params...)
	batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)

	results := ds.SendBatch(ctx, batch)

	return results
}

var emptyTag = pgconn.CommandTag{}

func (ds DbSession) SessionBatchExec(ctx context.Context, query string, params ...any) (pgconn.CommandTag, error) {
	results := ds.SessionBatch(ctx, query, params...)
	defer results.Close()

	if _, err := results.Exec(); err != nil {
		err = results.Close()
		if err != nil {
			return emptyTag, util.ErrCheck(err)
		}
		return emptyTag, util.ErrCheck(err)
	}

	commandTag, err := results.Exec()
	if err != nil {
		err = results.Close()
		if err != nil {
			return emptyTag, util.ErrCheck(err)
		}
		return emptyTag, util.ErrCheck(err)
	}

	if _, err := results.Exec(); err != nil {
		err = results.Close()
		if err != nil {
			return emptyTag, util.ErrCheck(err)
		}
		return emptyTag, util.ErrCheck(err)
	}

	return commandTag, nil
}

func (ds DbSession) SessionBatchQuery(ctx context.Context, query string, params ...any) (pgx.Rows, func(), error) {
	results := ds.SessionBatch(ctx, query, params...)

	if _, err := results.Exec(); err != nil {
		err = results.Close()
		if err != nil {
			return nil, nil, util.ErrCheck(err)
		}
		return nil, nil, util.ErrCheck(err)
	}

	rows, err := results.Query()
	if err != nil {
		err = results.Close()
		if err != nil {
			return nil, nil, util.ErrCheck(err)
		}
		return nil, nil, util.ErrCheck(err)
	}

	done := func() {
		rows.Close()
		err = results.Close()
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
		}
	}

	return rows, done, nil
}

func (ds DbSession) SessionBatchQueryRow(ctx context.Context, query string, params ...any) (pgx.Row, func(), error) {
	results := ds.SessionBatch(ctx, query, params...)

	if _, err := results.Exec(); err != nil {
		err = results.Close()
		if err != nil {
			return nil, nil, util.ErrCheck(err)
		}
		return nil, nil, util.ErrCheck(err)
	}

	row := results.QueryRow()

	done := func() {
		err := results.Close()
		if err != nil {
			util.ErrorLog.Println(util.ErrCheck(err))
		}
	}

	return row, done, nil
}

// Open a session batch with the intention of adding multiple queries
func (ds DbSession) SessionOpenBatch(ctx context.Context) *pgx.Batch {
	batch := &pgx.Batch{}

	batch.Queue(setSessionVariablesSQL, ds.ConcurrentUserSession.GetUserSub(), ds.ConcurrentUserSession.GetGroupId(), ds.ConcurrentUserSession.GetRoleBits(), ds.Topic)

	return batch
}

// Close a batch opened with SessionOpenBatch. The caller should handle all but the first (session set) queries.
func (ds DbSession) SessionSendBatch(ctx context.Context, batch *pgx.Batch) (pgx.BatchResults, error) {
	batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)

	results := ds.SendBatch(ctx, batch)

	if _, err := results.Exec(); err != nil {
		err = results.Close()
		if err != nil {
			return nil, util.ErrCheck(err)
		}
		return nil, util.ErrCheck(err)
	}

	return results, nil
}
