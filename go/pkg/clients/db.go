package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"

	"database/sql"

	_ "github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	colTypes *ColTypes
)

const (
	emptyString            = ""
	emptyInteger           = 0
	setSessionVariablesSQL = `SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::INTEGER, $4::VARCHAR)`
)

type Database struct {
	DatabaseClient      *DatabaseClient
	DatabaseAdminSub    string
	DatabaseAdminRoleId string
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
	pgPass, err := util.EnvFile(os.Getenv("PG_WORKER_PASS_FILE"))
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	connString2 := fmt.Sprintf("%s://%s:%s@/%s?host=%s&sslmode=disable", dbDriver, pgUser, pgPass, pgDb, os.Getenv("UNIX_SOCK_DIR"))
	dbpool, err := pgxpool.New(context.Background(), connString2)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	dbc := &Database{
		DatabaseClient: &DatabaseClient{
			Pool: dbpool,
		},
	}

	colTypes = &ColTypes{
		reflect.TypeOf(sql.NullString{}),
		reflect.TypeOf(sql.NullInt32{}),
		reflect.TypeOf(sql.NullInt64{}),
		reflect.TypeOf(sql.NullFloat64{}),
		reflect.TypeOf(sql.NullBool{}),
		reflect.TypeOf(JSONSerializer{}),
		reflect.TypeOf(&time.Time{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	workerDbSession := &DbSession{
		Pool: dbpool,
		UserSession: &types.UserSession{
			UserSub: "worker",
		},
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
func (db *Database) BuildInserts(sb *strings.Builder, size, current int) error {

	baseIndex := (current / size) * size

	_, err := sb.WriteString("(")
	if err != nil {
		return util.ErrCheck(err)
	}

	for i := 0; i < size; i++ {
		_, err = sb.WriteString("$")
		if err != nil {
			return util.ErrCheck(err)
		}
		_, err = sb.WriteString(strconv.Itoa(baseIndex + i + 1))
		if err != nil {
			return err
		}
		if i < size-1 {
			_, err = sb.WriteString(", ")
			if err != nil {
				return err
			}
		}
	}

	_, err = sb.WriteString("),")
	if err != nil {
		return err
	}

	return nil
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

func (dc *DatabaseClient) OpenPoolSessionGroupTx(ctx context.Context, session *types.UserSession) (*PoolTx, *types.UserSession, error) {
	groupSession := &types.UserSession{
		UserSub: session.GroupSub,
		GroupId: session.GroupId,
	}
	groupPoolTx, err := dc.OpenPoolSessionTx(ctx, groupSession)
	return groupPoolTx, groupSession, err
}

func (dc *DatabaseClient) OpenPoolSessionTx(ctx context.Context, session *types.UserSession) (*PoolTx, error) {
	tx, err := dc.Pool.Begin(ctx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	poolTx := &PoolTx{
		Tx: tx,
	}

	poolTx.SetSession(ctx, session)

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

func (ptx *PoolTx) SetSession(ctx context.Context, session *types.UserSession) error {
	_, err := ptx.Exec(ctx, setSessionVariablesSQL, session.UserSub, session.GroupId, session.RoleBits, emptyString)
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
	*types.UserSession
	*pgxpool.Pool
}

func NewGroupDbSession(pool *pgxpool.Pool, session *types.UserSession) DbSession {
	return DbSession{
		Pool: pool,
		UserSession: &types.UserSession{
			UserSub: session.GroupSub,
			GroupId: session.GroupId,
		},
	}
}

func (ds DbSession) SessionBatch(ctx context.Context, primaryQuery string, params ...interface{}) pgx.BatchResults {
	batch := &pgx.Batch{}

	batch.Queue(setSessionVariablesSQL, ds.UserSession.UserSub, ds.UserSession.GroupId, ds.UserSession.RoleBits, ds.Topic)
	batch.Queue(primaryQuery, params...)
	batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)

	results := ds.SendBatch(ctx, batch)

	return results
}

var emptyTag = pgconn.CommandTag{}

func (ds DbSession) SessionBatchExec(ctx context.Context, query string, params ...interface{}) (pgconn.CommandTag, error) {
	results := ds.SessionBatch(ctx, query, params...)
	defer results.Close()

	if _, err := results.Exec(); err != nil {
		results.Close()
		return emptyTag, util.ErrCheck(err)
	}

	commandTag, err := results.Exec()
	if err != nil {
		results.Close()
		return emptyTag, util.ErrCheck(err)
	}

	if _, err := results.Exec(); err != nil {
		results.Close()
		return emptyTag, util.ErrCheck(err)
	}

	return commandTag, nil
}

func (ds DbSession) SessionBatchQuery(ctx context.Context, query string, params ...interface{}) (pgx.Rows, func(), error) {
	results := ds.SessionBatch(ctx, query, params...)

	if _, err := results.Exec(); err != nil {
		results.Close()
		return nil, nil, util.ErrCheck(err)
	}

	rows, err := results.Query()
	if err != nil {
		results.Close()
		return nil, nil, util.ErrCheck(err)
	}

	done := func() {
		rows.Close()
		results.Close()
	}

	return rows, done, nil
}

func (ds DbSession) SessionBatchQueryRow(ctx context.Context, query string, params ...interface{}) (pgx.Row, func(), error) {
	results := ds.SessionBatch(ctx, query, params...)

	if _, err := results.Exec(); err != nil {
		results.Close()
		return nil, nil, util.ErrCheck(err)
	}

	row := results.QueryRow()

	done := func() {
		results.Close()
	}

	return row, done, nil
}

func (ds DbSession) SessionOpenBatch(ctx context.Context) *pgx.Batch {
	batch := &pgx.Batch{}

	batch.Queue(setSessionVariablesSQL, ds.UserSession.UserSub, ds.UserSession.GroupId, ds.UserSession.RoleBits, ds.Topic)

	return batch
}

func (ds DbSession) SessionSendBatch(ctx context.Context, batch *pgx.Batch) (pgx.BatchResults, error) {
	batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)

	results := ds.SendBatch(ctx, batch)

	if _, err := results.Exec(); err != nil {
		results.Close()
		return nil, util.ErrCheck(err)
	}

	return results, nil
}

func OpenBatch(ctx context.Context, sub, groupId string, roleBits int32) *pgx.Batch {
	batch := &pgx.Batch{}

	batch.Queue(setSessionVariablesSQL, sub, groupId, roleBits, "")

	return batch
}

func SendBatch(ctx context.Context, tx *PoolTx, batch *pgx.Batch) (pgx.BatchResults, error) {
	batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)

	results := tx.SendBatch(ctx, batch)

	if _, err := results.Exec(); err != nil {
		results.Close()
		return nil, util.ErrCheck(err)
	}

	return results, nil
}

func (db *Database) QueryRows(ctx context.Context, tx *PoolTx, protoStructSlice interface{}, query string, args ...interface{}) error {

	protoValue := reflect.ValueOf(protoStructSlice)
	if protoValue.Kind() != reflect.Ptr || protoValue.Elem().Kind() != reflect.Slice {
		return util.ErrCheck(errors.New("must provide a pointer to a slice"))
	}

	protoType := protoValue.Elem().Type().Elem()

	indexes := cachedFieldIndexes(protoType.Elem())

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return util.ErrCheck(err)
	}

	defer rows.Close()

	fds := rows.FieldDescriptions()

	var columns []string
	var columnTypes []uint32

	for _, fd := range fds {
		if fd.DataTypeOID != uint32(0) {
			columns = append(columns, fd.Name)
			columnTypes = append(columnTypes, fd.DataTypeOID)
		}
	}

	for rows.Next() {
		newElem := reflect.New(protoType.Elem())
		var values []interface{}
		deferrals := make([]func() error, 0)

		for i, column := range columns {
			index, ok := indexes[column]
			if ok {
				safeVal := reflect.New(mapIntTypeToNullType(columnTypes[i]))
				values = append(values, safeVal.Interface())
				deferrals = append(deferrals, func() error {
					return extractValue(newElem.Elem().Field(index), safeVal)
				})
			} else {
				var noMatch interface{}
				values = append(values, &noMatch)
			}
		}

		if err := rows.Scan(values...); err != nil {
			return util.ErrCheck(err)
		}

		for _, d := range deferrals {
			err := d()
			if err != nil {
				return util.ErrCheck(err)
			}
		}

		protoValue.Elem().Set(reflect.Append(protoValue.Elem(), newElem.Elem().Addr()))
	}

	return nil
}

// fieldIndexes returns a map of database column name to struct field index.
func fieldIndexes(structType reflect.Type) map[string]int {
	indexes := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := strings.Split(field.Tag.Get("json"), ",")[0]
		if tag != "" {
			indexes[tag] = i
		} else {
			indexes[field.Name] = i
		}
	}
	return indexes
}

var fieldIndexesCache sync.Map // map[reflect.Type]map[string]int

// cachedFieldIndexes is like fieldIndexes, but cached per struct type.
func cachedFieldIndexes(structType reflect.Type) map[string]int {
	if f, ok := fieldIndexesCache.Load(structType); ok {
		return f.(map[string]int)
	}
	indexes := fieldIndexes(structType)
	fieldIndexesCache.Store(structType, indexes)
	return indexes
}

type JSONSerializer []byte

func (pms *JSONSerializer) Scan(src interface{}) error {

	var source []byte

	switch s := src.(type) {
	case []byte:
		source = s
	case string:
		source = []byte(s)
	case nil:
		source = []byte("{}")
	default:
		return errors.New("incompatible type for ProtoMapSerializer")
	}

	*pms = source

	return nil
}

func mapIntTypeToNullType(t uint32) reflect.Type {
	switch t {
	case pgtype.TimestamptzOID:
		return colTypes.reflectTimestamp
	case pgtype.VarcharOID, pgtype.TimestampOID, pgtype.DateOID, pgtype.IntervalOID, pgtype.TextOID, pgtype.UUIDOID:
		return colTypes.reflectString
	case pgtype.Int8OID, pgtype.Int4OID, pgtype.Int2OID:
		return colTypes.reflectInt32
	// case INTEGER", "SMALLINT":
	// return colTypes.reflectInt64
	case pgtype.BoolOID:
		return colTypes.reflectBool
	case pgtype.JSONOID, pgtype.JSONBOID:
		return colTypes.reflectJson
	default:
		return nil
	}
}

func mapTypeToNullType(t string) reflect.Type {
	switch t {
	case "TIMESTAMPTZ":
		return colTypes.reflectTimestamp
	case "VARCHAR", "CHAR", "TIMESTAMP", "DATE", "INTERVAL", "TEXT", "UUID":
		return colTypes.reflectString
	case "INT8", "INT4", "INT2":
		return colTypes.reflectInt32
	case "INTEGER", "SMALLINT":
		return colTypes.reflectInt64
	case "BOOL":
		return colTypes.reflectBool
	case "JSON", "JSONB":
		return colTypes.reflectJson
	default:
		return nil
	}
}

// If reflect.Value errors are seen here it could mean that the protobuf value doesn't
// match something that can be serialized from its db column type. For example, if the
// db has a column of an INT type, but the protobuf is a string, we would see errors here
func extractValue(dst, src reflect.Value) error {
	if dst.IsValid() && dst.CanSet() {
		if src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface {
			src = reflect.Indirect(src)
		}
		switch src.Type() {
		case colTypes.reflectTimestamp:
			if !src.IsNil() {
				timestamp, ok := src.Interface().(*time.Time)
				if ok {
					dst.Set(reflect.ValueOf(&timestamppb.Timestamp{Seconds: timestamp.Unix()}))
				}
			}
		case colTypes.reflectString:
			dst.SetString(src.FieldByName("String").String())
		case colTypes.reflectInt32:
			dst.SetInt(src.FieldByName("Int32").Int())
		case colTypes.reflectInt64:
			dst.SetInt(src.FieldByName("Int64").Int())
		case colTypes.reflectBool:
			dst.SetBool(src.FieldByName("Bool").Bool())
		case colTypes.reflectJson:
			dstType := dst.Type()

			// The following serializes dbview JSON data into existing proto structs
			// The dbviews will select JSON with the structure of one or many (map) of an object
			// Handle map[string]*types.IExample as a top level proto struct field
			if dstType.Kind() == reflect.Map {

				elemType := dstType.Elem()

				dstMap := reflect.MakeMap(dstType)

				var tmpMap map[string]json.RawMessage
				err := json.Unmarshal(src.Bytes(), &tmpMap)
				if err != nil {
					return util.ErrCheck(err)
				}

				for key, val := range tmpMap {
					protoMessageElem := reflect.New(elemType.Elem())
					protoMessage, ok := protoMessageElem.Interface().(proto.Message)
					if !ok {
						return util.ErrCheck(errors.New("mapped element is not a proto message"))
					}

					err := protojson.Unmarshal(val, protoMessage)
					if err != nil {
						return util.ErrCheck(fmt.Errorf("failed to proto unmarshal %w %s", err, string(val)))
					}

					dstMap.SetMapIndex(reflect.ValueOf(key), protoMessageElem)
				}

				dst.Set(dstMap)

				// Handle *types.IExample as a top level proto struct field
			} else {

				newProtoMsg := reflect.New(dstType.Elem())
				msg, ok := newProtoMsg.Interface().(proto.Message)

				// It's not a top level proto struct
				if !ok {
					// Fallback to regular json unmarshal
					protoStruct := reflect.New(dstType)
					err := json.Unmarshal(src.Bytes(), protoStruct.Interface())
					if err != nil {
						return util.ErrCheck(errors.New("fallback parsing failed during query rows"))
					}
					dst.Set(protoStruct.Elem())
					return nil
				}

				if err := protojson.Unmarshal(src.Bytes(), msg); err != nil {
					println("Failed to marshal", string(src.Bytes()))
					return util.ErrCheck(err)
				}

				dst.Set(newProtoMsg)
			}
		default:
			dst.Set(src)
		}
	}
	return nil
}
