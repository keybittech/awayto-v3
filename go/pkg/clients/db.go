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

type ColTypes struct {
	reflectString    reflect.Type
	reflectInt32     reflect.Type
	reflectInt64     reflect.Type
	reflectFloat64   reflect.Type
	reflectBool      reflect.Type
	reflectJson      reflect.Type
	reflectTimestamp reflect.Type
}

var colTypes *ColTypes

var emptyString string
var setSessionVariablesSQL = `SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::VARCHAR, $4::VARCHAR)`
var setSessionVariablesSQLReplacer = `SELECT dbfunc_schema.set_session_vars($a::VARCHAR, $b::VARCHAR, $c::VARCHAR)`

type Pool struct {
	*pgxpool.Pool
}

type PoolTx struct {
	pgx.Tx
}

func (ptx PoolTx) Rollback() {
	ptx.Tx.Rollback(context.Background())
}

func (ptx PoolTx) Commit() error {
	return ptx.Tx.Commit(context.Background())
}

func (ptx PoolTx) Query(query string, params ...interface{}) (pgx.Rows, error) {
	return ptx.Tx.Query(context.Background(), query, params...)
}

func (ptx PoolTx) QueryRow(query string, params ...interface{}) pgx.Row {
	return ptx.Tx.QueryRow(context.Background(), query, params...)
}

func (ptx PoolTx) Exec(query string, params ...interface{}) (pgconn.CommandTag, error) {
	return ptx.Tx.Exec(context.Background(), query, params...)
}

func (ptx PoolTx) SetSession(session *types.UserSession) {
	ptx.Exec(setSessionVariablesSQL, session.UserSub, session.GroupId, session.Roles, emptyString)
}

func (ptx PoolTx) UnsetSession() {
	ptx.Exec(setSessionVariablesSQL, emptyString, emptyString, emptyString, emptyString)
}

type Database struct {
	DatabaseClient      *Pool
	DatabaseAdminSub    string
	DatabaseAdminRoleId string
}

func (db *Database) Client() *Pool {
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
		println("ERROR", err.Error())
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	dbc := &Database{
		DatabaseClient: &Pool{
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
		Ctx:  ctx,
		UserSession: &types.UserSession{
			UserSub: "worker",
		},
	}

	row, close, err := workerDbSession.SessionBatchQueryRow(`
		SELECT u.sub, r.id FROM dbtable_schema.users u
		JOIN dbtable_schema.roles r ON r.name = 'Admin'
		WHERE u.username = 'system_owner'
	`)
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}
	defer close()

	err = row.Scan(&dbc.DatabaseAdminSub, &dbc.DatabaseAdminRoleId)
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	_, err = workerDbSession.SessionBatchExec(`
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

func (db *Database) SetDbVar(prop, value string) error {
	tx, err := db.Client().Pool.Begin(context.Background())
	_, err = tx.Exec(context.Background(), fmt.Sprintf("SET SESSION app_session.%s = '%s'", prop, value))
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
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

const (
	sessionVarQuery  = "SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::VARCHAR)"
	sessionVarPrefix = "WITH session_setup AS (SELECT dbfunc_schema.set_session_vars($"
	varcharSeparator = "::VARCHAR, $"
	sessionVarSuffix = "::VARCHAR)) "
)

func (db *Database) BuildSessionQuery(userSub, groupId, roles, query string, args ...interface{}) (string, []interface{}, error) {
	argLen := len(args)

	allParams := make([]interface{}, argLen+3)

	copy(allParams, args)

	allParams[argLen] = userSub
	allParams[argLen+1] = groupId
	allParams[argLen+2] = roles

	var finalQuery strings.Builder
	finalQuery.Grow(len(sessionVarPrefix) + len(query) + 30)

	finalQuery.WriteString(sessionVarPrefix)
	finalQuery.WriteString(strconv.Itoa(argLen + 1))
	finalQuery.WriteString(varcharSeparator)
	finalQuery.WriteString(strconv.Itoa(argLen + 2))
	finalQuery.WriteString(varcharSeparator)
	finalQuery.WriteString(strconv.Itoa(argLen + 3))
	finalQuery.WriteString(sessionVarSuffix)
	finalQuery.WriteString(query)

	return finalQuery.String(), allParams, nil
}

// var batchPool = &sync.Pool{
// 	New: func() interface{} {
// 		return &pgx.Batch{}
// 	},
// }

type DbSession struct {
	*types.UserSession
	*pgxpool.Pool
	Topic string
	Ctx   context.Context
}

func (ds *DbSession) SessionBatch(primaryQuery string, params ...interface{}) (pgx.BatchResults, error) {
	if ds.Ctx == nil {
		return nil, util.ErrCheck(errors.New("context must be provided"))
	}

	batch := &pgx.Batch{}
	batch.Queue(setSessionVariablesSQL, ds.UserSession.UserSub, ds.UserSession.GroupId, ds.UserSession.Roles, ds.Topic)
	batch.Queue(primaryQuery, params...)
	batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyString, emptyString)

	return ds.SendBatch(ds.Ctx, batch), nil
}

var emptyTag = pgconn.CommandTag{}

func (ds *DbSession) SessionBatchExec(query string, params ...interface{}) (pgconn.CommandTag, error) {
	results, err := ds.SessionBatch(query, params...)
	if err != nil {
		return emptyTag, util.ErrCheck(err)
	}
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

func (ds *DbSession) SessionBatchQuery(query string, params ...interface{}) (pgx.Rows, func() error, error) {
	results, err := ds.SessionBatch(query, params...)
	if err != nil {
		return nil, nil, util.ErrCheck(err)
	}

	if _, err := results.Exec(); err != nil {
		results.Close()
		return nil, nil, util.ErrCheck(err)
	}

	rows, err := results.Query()
	if err != nil {
		results.Close()
		return nil, nil, util.ErrCheck(err)
	}

	close := func() error {
		defer results.Close()
		_, err := results.Exec()
		if err != nil {
			return util.ErrCheck(err)
		}
		return nil
	}

	return rows, close, nil
}

func (ds *DbSession) SessionBatchQueryRow(query string, params ...interface{}) (pgx.Row, func() error, error) {
	results, err := ds.SessionBatch(query, params...)
	if err != nil {
		return nil, nil, util.ErrCheck(err)
	}

	if _, err := results.Exec(); err != nil {
		results.Close()
		return nil, nil, util.ErrCheck(err)
	}

	row := results.QueryRow()

	close := func() error {
		defer results.Close()
		_, err := results.Exec()
		if err != nil {
			return util.ErrCheck(err)
		}
		return nil
	}

	return row, close, nil
}

// func (db *Database) TxExec(doFunc func(PoolTx) error, ids ...string) error {
// 	if ids == nil || len(ids) != 3 {
// 		return util.ErrCheck(errors.New("improperly structured TxExec ids"))
// 	}
//
// 	tx, err := db.Client().Begin()
// 	if err != nil {
// 		return util.ErrCheck(err)
// 	}
// 	defer tx.Rollback()
//
// 	_, err = tx.Exec(setSessionVariablesSQL, ids[0], ids[1], ids[2], emptyString)
// 	if err != nil {
// 		return util.ErrCheck(err)
// 	}
//
// 	err = doFunc(tx)
// 	if err != nil {
// 		return util.ErrCheck(err)
// 	}
//
// 	_, err = tx.Exec(setSessionVariablesSQL, emptyString, emptyString, emptyString, emptyString)
// 	if err != nil {
// 		return util.ErrCheck(err)
// 	}
//
// 	err = tx.Commit()
// 	if err != nil {
// 		return util.ErrCheck(err)
// 	}
//
// 	return nil
// }

// func (db *Database) QueryRows(protoStructSlice interface{}, query string, args ...interface{}) error {
//
// 	protoValue := reflect.ValueOf(protoStructSlice)
// 	if protoValue.Kind() != reflect.Ptr || protoValue.Elem().Kind() != reflect.Slice {
// 		return errors.New("must provide a pointer to a slice")
// 	}
//
// 	protoType := protoValue.Elem().Type().Elem()
//
// 	rows, err := db.Client().Query(query, args...)
// 	if err != nil {
// 		return err
// 	}
//
// 	defer rows.Close()
//
// 	columns, err := rows.Columns()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	columnTypes, err := rows.ColumnTypes()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	for rows.Next() {
// 		newElem := reflect.New(protoType.Elem())
// 		values := make([]interface{}, len(columns))
// 		deferrals := make([]func(), 0)
//
// 		for i, col := range columnTypes {
//
// 			colType := col.DatabaseTypeName()
//
// 			for k := 0; k < protoType.Elem().NumField(); k++ {
// 				fName := strings.Split(protoType.Elem().Field(k).Tag.Get("json"), ",")[0]
//
// 				if fName != columns[i] {
// 					continue
// 				}
//
// 				fVal := newElem.Elem().Field(k)
//
// 				safeVal := reflect.New(mapTypeToNullType(colType))
// 				values[i] = safeVal.Interface()
//
// 				deferrals = append(deferrals, func() {
// 					extractValue(fVal, safeVal)
// 				})
//
// 				break
// 			}
// 		}
//
// 		if err := rows.Scan(values...); err != nil {
// 			return err
// 		}
//
// 		for _, d := range deferrals {
// 			d()
// 		}
//
// 		protoValue.Elem().Set(reflect.Append(protoValue.Elem(), newElem.Elem().Addr()))
// 	}
//
// 	return nil
// }

func (db *Database) QueryRows(tx *PoolTx, protoStructSlice interface{}, query string, args ...interface{}) error {

	protoValue := reflect.ValueOf(protoStructSlice)
	if protoValue.Kind() != reflect.Ptr || protoValue.Elem().Kind() != reflect.Slice {
		return util.ErrCheck(errors.New("must provide a pointer to a slice"))
	}

	protoType := protoValue.Elem().Type().Elem()

	indexes := cachedFieldIndexes(protoType.Elem())

	rows, err := tx.Query(query, args...)
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
