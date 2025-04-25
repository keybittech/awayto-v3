package clients

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

var selectAdminRoleIdSQL = `SELECT id FROM dbtable_schema.roles WHERE name = 'Admin'`

func BenchmarkDbPgxQueryRow(b *testing.B) {
	db := InitDatabase().DatabaseClient.Pool
	defer db.Close()
	ctx := context.Background()
	var testStr string
	reset(b)
	for c := 0; c < b.N; c++ {
		db.QueryRow(ctx, selectAdminRoleIdSQL).Scan(&testStr)
	}
}

func BenchmarkDbPgxExec(b *testing.B) {
	db := InitDatabase().DatabaseClient.Pool
	defer db.Close()
	ctx := context.Background()
	reset(b)
	for c := 0; c < b.N; c++ {
		db.Exec(ctx, selectAdminRoleIdSQL)
	}
}

func BenchmarkDbPgxTx(b *testing.B) {
	db := InitDatabase().DatabaseClient.Pool
	defer db.Close()
	ctx := context.Background()
	var adminRoleId string
	reset(b)
	for c := 0; c < b.N; c++ {
		tx, _ := db.Begin(ctx)
		defer tx.Rollback(ctx)
		tx.QueryRow(ctx, selectAdminRoleIdSQL).Scan(&adminRoleId)
		tx.Commit(ctx)
	}
}

func BenchmarkDbPgxBatchNoCommit(b *testing.B) {
	db := InitDatabase().DatabaseClient.Pool
	defer db.Close()
	ctx := context.Background()
	var adminRoleId string

	var emptyString string
	var setSessionVariablesSQL = `SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::VARCHAR)`

	reset(b)
	for c := 0; c < b.N; c++ {
		tx, err := db.Begin(ctx)
		if err != nil {
			b.Fatal(err)
		}

		batch := &pgx.Batch{}

		batch.Queue(setSessionVariablesSQL, "worker", "", "")
		batch.Queue(selectAdminRoleIdSQL)
		batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyString)
		results := tx.SendBatch(ctx, batch)

		_, _ = results.Exec()

		err = results.QueryRow().Scan(&adminRoleId)
		if err != nil {
			results.Close()
			tx.Rollback(ctx)
			b.Fatal(err)
		}

		_, _ = results.Exec()

		err = results.Close()
		if err != nil {
			tx.Rollback(ctx)
			b.Fatal(err)
		}

		tx.Rollback(ctx)
	}
}

// func BenchmarkDbSocketGetTopicMessageParticipants(b *testing.B) {
// 	db := InitDatabase()
// 	defer db.DatabaseClient.Close()
//
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
//
// 	}
// }

func BenchmarkDbSocketGetSocketAllowances(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	bookingId := integrationTest.Bookings[1].Id
	ctx := context.Background()
	reset(b)
	for c := 0; c < b.N; c++ {
		b.StopTimer()
		ds := &DbSession{
			Pool:        db.DatabaseClient.Pool,
			UserSession: integrationTest.TestUsers[int32(c%7)].UserSession,
		}
		b.StartTimer()
		ds.GetSocketAllowances(ctx, bookingId)
	}
}

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
			if got := InitDatabase(); got == nil {
				t.Errorf("InitDatabase() = %v, want %v", got, tt.want)
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

func TestDatabase_BuildSessionQuery(t *testing.T) {
	type args struct {
		userSub string
		groupId string
		roles   string
		query   string
		args    []interface{}
	}
	tests := []struct {
		name    string
		db      *Database
		args    args
		want    string
		want1   []interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.db.BuildSessionQuery(tt.args.userSub, tt.args.groupId, tt.args.roles, tt.args.query, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.BuildSessionQuery(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.userSub, tt.args.groupId, tt.args.roles, tt.args.query, tt.args.args, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Database.BuildSessionQuery(%v, %v, %v, %v, %v) got = %v, want %v", tt.args.userSub, tt.args.groupId, tt.args.roles, tt.args.query, tt.args.args, got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Database.BuildSessionQuery(%v, %v, %v, %v, %v) got1 = %v, want %v", tt.args.userSub, tt.args.groupId, tt.args.roles, tt.args.query, tt.args.args, got1, tt.want1)
			}
		})
	}
}
