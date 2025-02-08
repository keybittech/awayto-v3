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

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Database struct {
	DatabaseClient      IDatabaseClient
	DatabaseAdminRoleId string
	DatabaseAdminSub    string
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
	pgPass := os.Getenv("PG_WORKER_PASS")
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	pgDb := os.Getenv("PG_DB")

	connString := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", dbDriver, pgUser, pgPass, pgHost, pgPort, pgDb)

	db, err := sql.Open(dbDriver, connString)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
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

	if err = db.QueryRow(`
		INSERT INTO dbtable_schema.users (sub, username, created_sub)
		VALUES ($1::uuid, 'system_owner', $1::uuid)
		RETURNING sub
	`, uuid.NewString()).Scan(&adminSub); err != nil {
		fmt.Printf("DB init startup %+v\n", err)

		db.QueryRow(`
			SELECT sub
			FROM dbtable_schema.users
			WHERE username = 'system_owner'
		`).Scan(&adminSub)

		db.QueryRow(`
			SELECT id
			FROM dbtable_schema.roles
			WHERE name = 'Admin'
		`).Scan(&adminRoleId)

	} else {
		db.QueryRow(`
			INSERT INTO dbtable_schema.roles (name)
			VALUES ('Admin')
			RETURNING id
   `).Scan(&adminRoleId)
	}

	dbc := &Database{}
	dbc.SetClient(&DBWrapper{db})
	dbc.SetAdminSub(adminSub)
	dbc.SetAdminRoleId(adminRoleId)

	return dbc
}

func (db *Database) Client() IDatabaseClient {
	return db.DatabaseClient
}

func (db *Database) SetClient(c IDatabaseClient) {
	db.DatabaseClient = c
}

func (db *Database) AdminSub() string {
	return db.DatabaseAdminSub
}

func (db *Database) SetAdminSub(sub string) {
	db.DatabaseAdminSub = sub
}

func (db *Database) AdminRoleId() string {
	return db.DatabaseAdminRoleId
}

func (db *Database) SetAdminRoleId(id string) {
	db.DatabaseAdminRoleId = id
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

func (tx *TxWrapper) QueryRow(query string, args ...interface{}) IRow {
	return tx.Tx.QueryRow(query, args...)
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
			protoStruct := reflect.New(dst.Type())
			json.Unmarshal(src.Bytes(), protoStruct.Interface())
			dst.Set(protoStruct.Elem())
		default:
			println("no match for extractValue, setting default")
			dst.Set(src)
		}

	}
}
