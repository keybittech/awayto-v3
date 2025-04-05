package clients

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	_ "github.com/lib/pq"
)

func TestInitDatabase(t *testing.T) {
	tests := []struct {
		name string
		want interfaces.IDatabase
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitDatabase(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabase_Client(t *testing.T) {
	tests := []struct {
		name string
		db   *Database
		want interfaces.IDatabaseClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.db.Client(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Database.Client() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabase_AdminSub(t *testing.T) {
	tests := []struct {
		name string
		db   *Database
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.db.AdminSub(); got != tt.want {
				t.Errorf("Database.AdminSub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabase_AdminRoleId(t *testing.T) {
	tests := []struct {
		name string
		db   *Database
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.db.AdminRoleId(); got != tt.want {
				t.Errorf("Database.AdminRoleId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabase_BuildInserts(t *testing.T) {
	type args struct {
		sb      *strings.Builder
		size    int
		current int
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.BuildInserts(tt.args.sb, tt.args.size, tt.args.current); (err != nil) != tt.wantErr {
				t.Errorf("Database.BuildInserts(%v, %v, %v) error = %v, wantErr %v", tt.args.sb, tt.args.size, tt.args.current, err, tt.wantErr)
			}
		})
	}
}

func TestDBWrapper_Conn(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		db      *DBWrapper
		args    args
		want    *sql.Conn
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, err := tt.db.Conn(tt.args.ctx)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("DBWrapper.Conn(%v) error = %v, wantErr %v", tt.args.ctx, err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("DBWrapper.Conn(%v) = %v, want %v", tt.args.ctx, got, tt.want)
			// }
		})
	}
}

func TestDBWrapper_Begin(t *testing.T) {
	tests := []struct {
		name    string
		db      *DBWrapper
		want    interfaces.IDatabaseTx
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.Begin()
			if (err != nil) != tt.wantErr {
				t.Errorf("DBWrapper.Begin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBWrapper.Begin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBWrapper_BeginTx(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *sql.TxOptions
	}
	tests := []struct {
		name    string
		db      *DBWrapper
		args    args
		want    interfaces.IDatabaseTx
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.BeginTx(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBWrapper.BeginTx(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.opts, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBWrapper.BeginTx(%v, %v) = %v, want %v", tt.args.ctx, tt.args.opts, got, tt.want)
			}
		})
	}
}

func TestDBWrapper_PrepareContext(t *testing.T) {
	type args struct {
		ctx   context.Context
		query string
	}
	tests := []struct {
		name    string
		db      *DBWrapper
		args    args
		want    *sql.Stmt
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.PrepareContext(tt.args.ctx, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBWrapper.PrepareContext(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.query, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBWrapper.PrepareContext(%v, %v) = %v, want %v", tt.args.ctx, tt.args.query, got, tt.want)
			}
		})
	}
}

func TestDBWrapper_ExecContext(t *testing.T) {
	type args struct {
		ctx   context.Context
		query string
		args  []interface{}
	}
	tests := []struct {
		name    string
		db      *DBWrapper
		args    args
		want    sql.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.ExecContext(tt.args.ctx, tt.args.query, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBWrapper.ExecContext(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.query, tt.args.args, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBWrapper.ExecContext(%v, %v, %v) = %v, want %v", tt.args.ctx, tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestDBWrapper_Exec(t *testing.T) {
	type args struct {
		query string
		args  []interface{}
	}
	tests := []struct {
		name    string
		db      *DBWrapper
		args    args
		want    sql.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.Exec(tt.args.query, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBWrapper.Exec(%v, %v) error = %v, wantErr %v", tt.args.query, tt.args.args, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBWrapper.Exec(%v, %v) = %v, want %v", tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestDBWrapper_Query(t *testing.T) {
	type args struct {
		query string
		args  []interface{}
	}
	tests := []struct {
		name    string
		db      *DBWrapper
		args    args
		want    interfaces.IRows
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.Query(tt.args.query, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBWrapper.Query(%v, %v) error = %v, wantErr %v", tt.args.query, tt.args.args, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBWrapper.Query(%v, %v) = %v, want %v", tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestDBWrapper_QueryRow(t *testing.T) {
	type args struct {
		query string
		args  []interface{}
	}
	tests := []struct {
		name string
		db   *DBWrapper
		args args
		want interfaces.IRow
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.db.QueryRow(tt.args.query, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBWrapper.QueryRow(%v, %v) = %v, want %v", tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestTxWrapper_Prepare(t *testing.T) {
	type args struct {
		stmt string
	}
	tests := []struct {
		name    string
		tx      *TxWrapper
		args    args
		want    *sql.Stmt
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tx.Prepare(tt.args.stmt)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.Prepare(%v) error = %v, wantErr %v", tt.args.stmt, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxWrapper.Prepare(%v) = %v, want %v", tt.args.stmt, got, tt.want)
			}
		})
	}
}

func TestTxWrapper_PrepareContext(t *testing.T) {
	type args struct {
		ctx   context.Context
		query string
	}
	tests := []struct {
		name    string
		tx      *TxWrapper
		args    args
		want    *sql.Stmt
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tx.PrepareContext(tt.args.ctx, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.PrepareContext(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.query, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxWrapper.PrepareContext(%v, %v) = %v, want %v", tt.args.ctx, tt.args.query, got, tt.want)
			}
		})
	}
}

func TestTxWrapper_Commit(t *testing.T) {
	tests := []struct {
		name    string
		tx      *TxWrapper
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tx.Commit(); (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.Commit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxWrapper_Rollback(t *testing.T) {
	tests := []struct {
		name    string
		tx      *TxWrapper
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tx.Rollback(); (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxWrapper_Exec(t *testing.T) {
	type args struct {
		query string
		args  []interface{}
	}
	tests := []struct {
		name    string
		tx      *TxWrapper
		args    args
		want    sql.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tx.Exec(tt.args.query, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.Exec(%v, %v) error = %v, wantErr %v", tt.args.query, tt.args.args, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxWrapper.Exec(%v, %v) = %v, want %v", tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestTxWrapper_ExecContext(t *testing.T) {
	type args struct {
		ctx   context.Context
		query string
		args  []interface{}
	}
	tests := []struct {
		name    string
		tx      *TxWrapper
		args    args
		want    sql.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tx.ExecContext(tt.args.ctx, tt.args.query, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.ExecContext(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.query, tt.args.args, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxWrapper.ExecContext(%v, %v, %v) = %v, want %v", tt.args.ctx, tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestTxWrapper_Query(t *testing.T) {
	type args struct {
		query string
		args  []interface{}
	}
	tests := []struct {
		name    string
		tx      *TxWrapper
		args    args
		want    interfaces.IRows
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tx.Query(tt.args.query, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.Query(%v, %v) error = %v, wantErr %v", tt.args.query, tt.args.args, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxWrapper.Query(%v, %v) = %v, want %v", tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestTxWrapper_QueryRow(t *testing.T) {
	type args struct {
		query string
		args  []interface{}
	}
	tests := []struct {
		name string
		tx   *TxWrapper
		args args
		want interfaces.IRow
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tx.QueryRow(tt.args.query, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxWrapper.QueryRow(%v, %v) = %v, want %v", tt.args.query, tt.args.args, got, tt.want)
			}
		})
	}
}

func TestTxWrapper_SetDbVar(t *testing.T) {
	type args struct {
		prop  string
		value string
	}
	tests := []struct {
		name    string
		tx      *TxWrapper
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tx.SetDbVar(tt.args.prop, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.SetDbVar(%v, %v) error = %v, wantErr %v", tt.args.prop, tt.args.value, err, tt.wantErr)
			}
		})
	}
}

func TestRowWrapper_Scan(t *testing.T) {
	type args struct {
		dest []interface{}
	}
	tests := []struct {
		name    string
		r       *RowWrapper
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Scan(tt.args.dest...); (err != nil) != tt.wantErr {
				t.Errorf("RowWrapper.Scan(%v) error = %v, wantErr %v", tt.args.dest, err, tt.wantErr)
			}
		})
	}
}

func TestIRowsWrapper_Next(t *testing.T) {
	tests := []struct {
		name string
		r    *IRowsWrapper
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Next(); got != tt.want {
				t.Errorf("IRowsWrapper.Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIRowsWrapper_Scan(t *testing.T) {
	type args struct {
		dest []interface{}
	}
	tests := []struct {
		name    string
		r       *IRowsWrapper
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Scan(tt.args.dest...); (err != nil) != tt.wantErr {
				t.Errorf("IRowsWrapper.Scan(%v) error = %v, wantErr %v", tt.args.dest, err, tt.wantErr)
			}
		})
	}
}

func TestIRowsWrapper_Close(t *testing.T) {
	tests := []struct {
		name    string
		r       *IRowsWrapper
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Close(); (err != nil) != tt.wantErr {
				t.Errorf("IRowsWrapper.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIRowsWrapper_Err(t *testing.T) {
	tests := []struct {
		name    string
		r       *IRowsWrapper
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Err(); (err != nil) != tt.wantErr {
				t.Errorf("IRowsWrapper.Err() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIRowsWrapper_Columns(t *testing.T) {
	tests := []struct {
		name    string
		r       *IRowsWrapper
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Columns()
			if (err != nil) != tt.wantErr {
				t.Errorf("IRowsWrapper.Columns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IRowsWrapper.Columns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIRowsWrapper_ColumnTypes(t *testing.T) {
	tests := []struct {
		name    string
		r       *IRowsWrapper
		want    []*sql.ColumnType
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.ColumnTypes()
			if (err != nil) != tt.wantErr {
				t.Errorf("IRowsWrapper.ColumnTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IRowsWrapper.ColumnTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabase_TxExec(t *testing.T) {
	type args struct {
		doFunc func(interfaces.IDatabaseTx) error
		ids    []string
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := tt.db.TxExec(tt.args.doFunc, tt.args.ids...); (err != nil) != tt.wantErr {
			// 	t.Errorf("Database.TxExec(%v, %v) error = %v, wantErr %v", tt.args.doFunc, tt.args.ids, err, tt.wantErr)
			// }
		})
	}
}

func TestDatabase_QueryRows(t *testing.T) {
	type args struct {
		protoStructSlice interface{}
		query            string
		args             []interface{}
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.QueryRows(tt.args.protoStructSlice, tt.args.query, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Database.QueryRows(%v, %v, %v) error = %v, wantErr %v", tt.args.protoStructSlice, tt.args.query, tt.args.args, err, tt.wantErr)
			}
		})
	}
}

func TestTxWrapper_QueryRows(t *testing.T) {
	type args struct {
		protoStructSlice interface{}
		query            string
		args             []interface{}
	}
	tests := []struct {
		name    string
		tx      *TxWrapper
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tx.QueryRows(tt.args.protoStructSlice, tt.args.query, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("TxWrapper.QueryRows(%v, %v, %v) error = %v, wantErr %v", tt.args.protoStructSlice, tt.args.query, tt.args.args, err, tt.wantErr)
			}
		})
	}
}

func Test_fieldIndexes(t *testing.T) {
	type args struct {
		structType reflect.Type
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fieldIndexes(tt.args.structType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fieldIndexes(%v) = %v, want %v", tt.args.structType, got, tt.want)
			}
		})
	}
}

func Test_cachedFieldIndexes(t *testing.T) {
	type args struct {
		structType reflect.Type
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cachedFieldIndexes(tt.args.structType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cachedFieldIndexes(%v) = %v, want %v", tt.args.structType, got, tt.want)
			}
		})
	}
}

func TestJSONSerializer_Scan(t *testing.T) {
	type args struct {
		src interface{}
	}
	tests := []struct {
		name    string
		pms     *JSONSerializer
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pms.Scan(tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.Scan(%v) error = %v, wantErr %v", tt.args.src, err, tt.wantErr)
			}
		})
	}
}

func Test_mapTypeToNullType(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapTypeToNullType(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapTypeToNullType(%v) = %v, want %v", tt.args.t, got, tt.want)
			}
		})
	}
}

func Test_extractValue(t *testing.T) {
	type args struct {
		dst reflect.Value
		src reflect.Value
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := extractValue(tt.args.dst, tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("extractValue(%v, %v) error = %v, wantErr %v", tt.args.dst, tt.args.src, err, tt.wantErr)
			}
		})
	}
}
