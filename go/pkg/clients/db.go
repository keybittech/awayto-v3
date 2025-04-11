package clients

import (
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
var setSessionVariablesSQL = `SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::VARCHAR)`

type Database struct {
	DatabaseClient      *sql.DB
	DatabaseAdminSub    string
	DatabaseAdminRoleId string
}

func (db *Database) Client() *sql.DB {
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
	// pgHost := os.Getenv("PG_HOST")
	// pgPort := os.Getenv("PG_PORT")
	pgDb := os.Getenv("PG_DB")
	pgPass, err := util.EnvFile(os.Getenv("PG_WORKER_PASS_FILE"))
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	// connString := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", dbDriver, pgUser, pgPass, pgHost, pgPort, pgDb)
	connString2 := fmt.Sprintf("%s://%s:%s@/%s?host=%s&sslmode=disable", dbDriver, pgUser, pgPass, pgDb, os.Getenv("UNIX_SOCK_DIR"))

	db, err := sql.Open(dbDriver, connString2)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	dbc := &Database{
		DatabaseClient: db,
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

	err = dbc.TxExec(func(tx *sql.Tx) error {
		var txErr error
		txErr = tx.QueryRow(`
			SELECT u.sub, r.id FROM dbtable_schema.users u
			JOIN dbtable_schema.roles r ON r.name = 'Admin'
			WHERE u.username = 'system_owner'
		`).Scan(&adminSub, &adminRoleId)
		if txErr != nil {
			return util.ErrCheck(txErr)
		}

		txErr = dbc.SetDbVar(tx, "sock_topic", "")
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

func (db *Database) SetDbVar(tx *sql.Tx, prop, value string) error {
	_, err := tx.Exec(fmt.Sprintf("SET SESSION app_session.%s = '%s'", prop, value))
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

func (db *Database) TxExec(doFunc func(*sql.Tx) error, ids ...string) error {
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

func (db *Database) QueryRows(tx *sql.Tx, protoStructSlice interface{}, query string, args ...interface{}) error {

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
