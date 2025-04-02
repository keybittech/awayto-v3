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

	"github.com/keybittech/awayto-v3/go/pkg/util"

	"database/sql"

	_ "github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Database struct {
	DatabaseClient      IDatabaseClient
	DatabaseAdminSub    string
	DatabaseAdminRoleId string
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

var colTypes *ColTypes

func InitDatabase() IDatabase {
	dbDriver := os.Getenv("DB_DRIVER")
	pgUser := os.Getenv("PG_WORKER")
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	pgDb := os.Getenv("PG_DB")
	pgPass, err := util.EnvFile(os.Getenv("PG_WORKER_PASS_FILE"))
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	connString := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", dbDriver, pgUser, pgPass, pgHost, pgPort, pgDb)

	db, err := sql.Open(dbDriver, connString)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	dbc := &Database{
		DatabaseClient: &DBWrapper{db},
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

	var adminRoleId, adminSub string

	err = dbc.TxExec(func(tx IDatabaseTx) error {
		var txErr error
		txErr = tx.QueryRow(`
			SELECT u.sub, r.id FROM dbtable_schema.users u
			JOIN dbtable_schema.roles r ON r.name = 'Admin'
			WHERE u.username = 'system_owner'
		`).Scan(&adminSub, &adminRoleId)
		if txErr != nil {
			return util.ErrCheck(txErr)
		}

		txErr = tx.SetDbVar("sock_topic", "")
		if txErr != nil {
			return util.ErrCheck(txErr)
		}

		_, txErr = tx.Exec(`
			DELETE FROM dbtable_schema.sock_connections
			USING dbtable_schema.sock_connections sc
			LEFT OUTER JOIN dbtable_schema.topic_messages tm ON tm.connection_id = sc.connection_id
			WHERE dbtable_schema.sock_connections.id = sc.id AND tm.id IS NULL
		`)
		if txErr != nil {
			return util.ErrCheck(txErr)
		}

		return nil
	}, "worker", "", "")
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	dbc.DatabaseAdminSub = adminSub
	dbc.DatabaseAdminRoleId = adminRoleId

	println(fmt.Sprintf("Database Initialized\nAdmin Sub: %s\nAdmin Role Id: %s", dbc.AdminSub(), dbc.AdminRoleId()))

	return dbc
}

func (db *Database) Client() IDatabaseClient {
	return db.DatabaseClient
}

func (db *Database) AdminSub() string {
	return db.DatabaseAdminSub
}

func (db *Database) AdminRoleId() string {
	return db.DatabaseAdminRoleId
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

// DB Wrappers
type DBWrapper struct {
	*sql.DB
}

func (db *DBWrapper) Conn(ctx context.Context) (*sql.Conn, error) {
	return db.Conn(ctx)
}

func (db *DBWrapper) Begin() (IDatabaseTx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &TxWrapper{tx}, nil
}

func (db *DBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (IDatabaseTx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &TxWrapper{tx}, nil
}

func (db *DBWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.DB.PrepareContext(ctx, query)
}

func (db *DBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *DBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

func (db *DBWrapper) Query(query string, args ...interface{}) (IRows, error) {
	return db.DB.Query(query, args...)
}

func (db *DBWrapper) QueryRow(query string, args ...interface{}) IRow {
	return db.DB.QueryRow(query, args...)
}

// TX Wrappers
type TxWrapper struct {
	*sql.Tx
}

func (tx *TxWrapper) Prepare(stmt string) (*sql.Stmt, error) {
	return tx.Tx.Prepare(stmt)
}

func (tx *TxWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return tx.Tx.PrepareContext(ctx, query)
}

func (tx *TxWrapper) Commit() error {
	return tx.Tx.Commit()
}

func (tx *TxWrapper) Rollback() error {
	return tx.Tx.Rollback()
}

func (tx *TxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.Exec(query, args...)
}

func (tx *TxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {

	return tx.Tx.ExecContext(ctx, query, args...)
}

func (tx *TxWrapper) Query(query string, args ...interface{}) (IRows, error) {
	return tx.Tx.Query(query, args...)
}

func (tx *TxWrapper) QueryRow(query string, args ...interface{}) IRow {
	return tx.Tx.QueryRow(query, args...)
}

func (tx *TxWrapper) SetDbVar(prop, value string) error {
	_, err := tx.Exec(fmt.Sprintf("SET SESSION app_session.%s = '%s'", prop, value))
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

// Row Wrappers
type RowWrapper struct {
	*sql.Row
}

func (r *RowWrapper) Scan(dest ...interface{}) error {
	return r.Row.Scan(dest...)
}

// Rows Wrappers
type IRowsWrapper struct {
	*sql.Rows
}

func (r *IRowsWrapper) Next() bool {
	return r.Rows.Next()
}

func (r *IRowsWrapper) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r *IRowsWrapper) Close() error {
	return r.Rows.Close()
}

func (r *IRowsWrapper) Err() error {
	return r.Rows.Err()
}

func (r *IRowsWrapper) Columns() ([]string, error) {
	return r.Rows.Columns()
}

func (r *IRowsWrapper) ColumnTypes() ([]*sql.ColumnType, error) {
	return r.Rows.ColumnTypes()
}

var emptyString string
var setSessionVariablesSQL = `SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::VARCHAR)`

func (db *Database) TxExec(doFunc func(IDatabaseTx) error, ids ...string) error {
	if ids == nil || len(ids) != 3 {
		return util.ErrCheck(errors.New("improperly structured TxExec ids"))
	}

	tx, err := db.Client().Begin()
	if err != nil {
		return util.ErrCheck(err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(setSessionVariablesSQL, ids[0], ids[1], ids[2])
	if err != nil {
		return util.ErrCheck(err)
	}

	err = doFunc(tx)
	if err != nil {
		return util.ErrCheck(err)
	}

	_, err = tx.Exec(setSessionVariablesSQL, emptyString, emptyString, emptyString)
	if err != nil {
		return util.ErrCheck(err)
	}

	err = tx.Commit()
	if err != nil {
		return util.ErrCheck(err)
	}

	return nil
}

func (db *Database) QueryRows(protoStructSlice interface{}, query string, args ...interface{}) error {

	protoValue := reflect.ValueOf(protoStructSlice)
	if protoValue.Kind() != reflect.Ptr || protoValue.Elem().Kind() != reflect.Slice {
		return errors.New("must provide a pointer to a slice")
	}

	protoType := protoValue.Elem().Type().Elem()

	rows, err := db.Client().Query(query, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		newElem := reflect.New(protoType.Elem())
		values := make([]interface{}, len(columns))
		deferrals := make([]func(), 0)

		for i, col := range columnTypes {

			colType := col.DatabaseTypeName()

			for k := 0; k < protoType.Elem().NumField(); k++ {
				fName := strings.Split(protoType.Elem().Field(k).Tag.Get("json"), ",")[0]

				if fName != columns[i] {
					continue
				}

				fVal := newElem.Elem().Field(k)

				safeVal := reflect.New(mapTypeToNullType(colType))
				values[i] = safeVal.Interface()

				deferrals = append(deferrals, func() {
					extractValue(fVal, safeVal)
				})

				break
			}
		}

		if err := rows.Scan(values...); err != nil {
			return err
		}

		for _, d := range deferrals {
			d()
		}

		protoValue.Elem().Set(reflect.Append(protoValue.Elem(), newElem.Elem().Addr()))
	}

	return nil
}

func (tx *TxWrapper) QueryRows(protoStructSlice interface{}, query string, args ...interface{}) error {

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

	columns, err := rows.Columns()
	if err != nil {
		return util.ErrCheck(err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return util.ErrCheck(err)
	}

	for rows.Next() {
		newElem := reflect.New(protoType.Elem())
		var values []interface{}
		deferrals := make([]func() error, 0)

		for i, column := range columns {
			index, ok := indexes[column]
			if ok {
				safeVal := reflect.New(mapTypeToNullType(columnTypes[i].DatabaseTypeName()))
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
