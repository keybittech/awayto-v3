package clients

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

func TestDatabase_Client(t *testing.T) {
	tests := []struct {
		name string
		db   *Database
		want *sql.DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := tt.db.Client(); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Database.Client() = %v, want %v", got, tt.want)
			// }
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

func TestInitDatabase(t *testing.T) {
	tests := []struct {
		name string
		want *Database
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

func TestDatabase_SetDbVar(t *testing.T) {
	type args struct {
		tx    *sql.Tx
		prop  string
		value string
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
			if err := tt.db.SetDbVar(tt.args.tx, tt.args.prop, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Database.SetDbVar(%v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.prop, tt.args.value, err, tt.wantErr)
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

func TestDatabase_TxExec(t *testing.T) {
	type args struct {
		doFunc func(*sql.Tx) error
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
			if err := tt.db.TxExec(tt.args.doFunc, tt.args.ids...); (err != nil) != tt.wantErr {
				// t.Errorf("Database.TxExec(%v, %v) error = %v, wantErr %v", tt.args.doFunc, tt.args.ids, err, tt.wantErr)
			}
		})
	}
}

func TestDatabase_QueryRows(t *testing.T) {
	type args struct {
		tx               *sql.Tx
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
			if err := tt.db.QueryRows(tt.args.tx, tt.args.protoStructSlice, tt.args.query, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Database.QueryRows(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.tx, tt.args.protoStructSlice, tt.args.query, tt.args.args, err, tt.wantErr)
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
