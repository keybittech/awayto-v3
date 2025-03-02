package clients

import (
	"av3api/pkg/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"database/sql"

	_ "github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Database struct {
	DatabaseClient      IDatabaseClient
	DatabaseAdminSub    string
	DatabaseAdminRoleId string
}

type ColTypes struct {
	reflectString  reflect.Type
	reflectInt32   reflect.Type
	reflectInt64   reflect.Type
	reflectFloat64 reflect.Type
	reflectBool    reflect.Type
	reflectMap     reflect.Type
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
		reflect.TypeOf(ProtoMapSerializer{}),
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

func (tx *TxWrapper) Commit() error {
	return tx.Tx.Commit()
}

func (tx *TxWrapper) Rollback() error {
	return tx.Tx.Rollback()
}

func (tx *TxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.Exec(query, args...)
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

type ProtoMapSerializer []byte

func (pms *ProtoMapSerializer) Scan(src interface{}) error {

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

func (db *Database) TxExec(doFunc func(IDatabaseTx) error, ids ...string) error {
	if ids == nil || len(ids) != 3 {
		return util.ErrCheck(errors.New("improperly structured TxExec ids"))
	}

	tx, err := db.Client().Begin()
	if err != nil {
		return util.ErrCheck(err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::VARCHAR)
	`, ids[0], ids[1], ids[2])
	if err != nil {
		return util.ErrCheck(err)
	}

	err = doFunc(tx)
	if err != nil {
		return util.ErrCheck(err)
	}

	_, err = tx.Exec(`
		SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::VARCHAR)
	`, "", "", "")
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
		return errors.New("must provide a pointer to a slice")
	}

	protoType := protoValue.Elem().Type().Elem()

	rows, err := tx.Query(query, args...)
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

func mapTypeToNullType(t string) reflect.Type {
	switch t {
	case "VARCHAR", "CHAR", "TIMESTAMP", "DATE", "INTERVAL", "TEXT", "UUID":
		return colTypes.reflectString
	case "INT8", "INT4":
		return colTypes.reflectInt32
	case "INTEGER", "SMALLINT":
		return colTypes.reflectInt64
	case "BOOL":
		return colTypes.reflectBool
	case "JSONB":
		return colTypes.reflectMap
	default:
		return nil
	}
}

func extractValue(dst, src reflect.Value) {
	if dst.IsValid() && dst.CanSet() {
		if src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface {
			src = reflect.Indirect(src)
		}
		switch src.Type() {
		case colTypes.reflectString:
			dst.SetString(src.FieldByName("String").String())
		case colTypes.reflectInt32:
			dst.SetInt(src.FieldByName("Int32").Int())
		case colTypes.reflectInt64:
			dst.SetInt(src.FieldByName("Int64").Int())
		case colTypes.reflectBool:
			dst.SetBool(src.FieldByName("Bool").Bool())
		case colTypes.reflectMap:
			jsonData := src.Interface().(ProtoMapSerializer)

			// Check if destination is a map
			if dst.Kind() == reflect.Map {
				// Handle map<string, proto.Message> case
				mapType := dst.Type()
				valueType := mapType.Elem()

				// Create a temporary Go map to hold the unmarshaled data
				var tempMap map[string]json.RawMessage
				if err := json.Unmarshal(jsonData, &tempMap); err != nil {
					println("Error parsing JSON map:", err.Error())
					return
				}

				// Initialize the destination map if it's nil
				if dst.IsNil() {
					dst.Set(reflect.MakeMap(mapType))
				}

				// For each key-value pair in the map
				for key, rawValue := range tempMap {
					// Create a new instance of the map value type
					protoMsgValue := reflect.New(valueType.Elem()).Interface().(proto.Message)

					// Unmarshal the raw JSON into the proto message
					if err := protojson.Unmarshal(rawValue, protoMsgValue); err != nil {
						println("Error unmarshaling map value for key:", key, "type:", valueType.Elem().Name(), "error:", err.Error())
						continue
					}

					// Set the key-value pair in the destination map
					dst.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(protoMsgValue))
				}
			} else if dst.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
				// Handle regular proto.Message case
				protoMsg := reflect.New(dst.Type().Elem()).Interface().(proto.Message)

				if err := protojson.Unmarshal(jsonData, protoMsg); err != nil {
					msgTypeName := dst.Type().Elem().Name()
					println("Error unmarshaling proto message type:", msgTypeName, "error:", err.Error())
					return
				}

				dst.Set(reflect.ValueOf(protoMsg))
			} else {
				println("Destination type is neither a map nor a proto.Message:", dst.Type().String())
			}
		default:
			println("no match for extractValue, setting default")
			dst.Set(src)
		}

	}
}
