package clients

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

// DB Wrappers
type DBWrapper struct {
	*sql.DB
}

func (db *DBWrapper) Begin() (interfaces.IDatabaseTx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &TxWrapper{tx}, nil
}

func (db *DBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (interfaces.IDatabaseTx, error) {
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

func (db *DBWrapper) Query(query string, args ...interface{}) (interfaces.IRows, error) {
	return db.DB.Query(query, args...)
}

func (db *DBWrapper) QueryRow(query string, args ...interface{}) interfaces.IRow {
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

func (tx *TxWrapper) Query(query string, args ...interface{}) (interfaces.IRows, error) {
	return tx.Tx.Query(query, args...)
}

func (tx *TxWrapper) QueryRow(query string, args ...interface{}) interfaces.IRow {
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
