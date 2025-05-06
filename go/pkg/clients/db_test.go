package clients

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keybittech/awayto-v3/go/pkg/types"
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

	reset(b)
	for c := 0; c < b.N; c++ {
		tx, err := db.Begin(ctx)
		if err != nil {
			b.Fatal(err)
		}

		batch := &pgx.Batch{}

		batch.Queue(setSessionVariablesSQL, "worker", emptyString, emptyInteger, emptyString)
		batch.Queue(selectAdminRoleIdSQL)
		batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)
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

func TestDatabaseClient_OpenPoolSessionGroupTx(t *testing.T) {
	type args struct {
		ctx     context.Context
		session *types.UserSession
	}
	tests := []struct {
		name    string
		dc      *DatabaseClient
		args    args
		want    *PoolTx
		want1   *types.UserSession
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.dc.OpenPoolSessionGroupTx(tt.args.ctx, tt.args.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseClient.OpenPoolSessionGroupTx(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.session, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseClient.OpenPoolSessionGroupTx(%v, %v) got = %v, want %v", tt.args.ctx, tt.args.session, got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("DatabaseClient.OpenPoolSessionGroupTx(%v, %v) got1 = %v, want %v", tt.args.ctx, tt.args.session, got1, tt.want1)
			}
		})
	}
}

func TestDatabaseClient_OpenPoolSessionTx(t *testing.T) {
	type args struct {
		ctx     context.Context
		session *types.UserSession
	}
	tests := []struct {
		name    string
		dc      *DatabaseClient
		args    args
		want    *PoolTx
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dc.OpenPoolSessionTx(tt.args.ctx, tt.args.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseClient.OpenPoolSessionTx(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.session, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseClient.OpenPoolSessionTx(%v, %v) = %v, want %v", tt.args.ctx, tt.args.session, got, tt.want)
			}
		})
	}
}

func TestDatabaseClient_ClosePoolSessionTx(t *testing.T) {
	type args struct {
		ctx    context.Context
		poolTx *PoolTx
	}
	tests := []struct {
		name    string
		dc      *DatabaseClient
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dc.ClosePoolSessionTx(tt.args.ctx, tt.args.poolTx); (err != nil) != tt.wantErr {
				t.Errorf("DatabaseClient.ClosePoolSessionTx(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.poolTx, err, tt.wantErr)
			}
		})
	}
}

func TestPoolTx_SetSession(t *testing.T) {
	type args struct {
		ctx     context.Context
		session *types.UserSession
	}
	tests := []struct {
		name    string
		ptx     *PoolTx
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ptx.SetSession(tt.args.ctx, tt.args.session); (err != nil) != tt.wantErr {
				t.Errorf("PoolTx.SetSession(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.session, err, tt.wantErr)
			}
		})
	}
}

func TestPoolTx_UnsetSession(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		ptx     *PoolTx
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ptx.UnsetSession(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("PoolTx.UnsetSession(%v) error = %v, wantErr %v", tt.args.ctx, err, tt.wantErr)
			}
		})
	}
}

func TestNewGroupDbSession(t *testing.T) {
	type args struct {
		pool    *pgxpool.Pool
		session *types.UserSession
	}
	tests := []struct {
		name string
		args args
		want DbSession
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := NewGroupDbSession(tt.args.pool, tt.args.session); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewGroupDbSession(%v, %v) = %v, want %v", tt.args.pool, tt.args.session, got, tt.want)
			// }
		})
	}
}

func TestDbSession_SessionBatch(t *testing.T) {
	type args struct {
		ctx          context.Context
		primaryQuery string
		params       []interface{}
	}
	tests := []struct {
		name string
		ds   DbSession
		args args
		want pgx.BatchResults
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ds.SessionBatch(tt.args.ctx, tt.args.primaryQuery, tt.args.params...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DbSession.SessionBatch(%v, %v, %v) = %v, want %v", tt.args.ctx, tt.args.primaryQuery, tt.args.params, got, tt.want)
			}
		})
	}
}

func TestDbSession_SessionBatchExec(t *testing.T) {
	type args struct {
		ctx    context.Context
		query  string
		params []interface{}
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		want    pgconn.CommandTag
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ds.SessionBatchExec(tt.args.ctx, tt.args.query, tt.args.params...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbSession.SessionBatchExec(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.query, tt.args.params, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DbSession.SessionBatchExec(%v, %v, %v) = %v, want %v", tt.args.ctx, tt.args.query, tt.args.params, got, tt.want)
			}
		})
	}
}

func TestDbSession_SessionBatchQuery(t *testing.T) {
	type args struct {
		ctx    context.Context
		query  string
		params []interface{}
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		want    pgx.Rows
		want1   func()
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, got1, err := tt.ds.SessionBatchQuery(tt.args.ctx, tt.args.query, tt.args.params...)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("DbSession.SessionBatchQuery(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.query, tt.args.params, err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("DbSession.SessionBatchQuery(%v, %v, %v) got = %v, want %v", tt.args.ctx, tt.args.query, tt.args.params, got, tt.want)
			// }
			// if !reflect.DeepEqual(got1, tt.want1) {
			// 	t.Errorf("DbSession.SessionBatchQuery(%v, %v, %v) got1 = %v, want %v", tt.args.ctx, tt.args.query, tt.args.params, got1, tt.want1)
			// }
		})
	}
}

func TestDbSession_SessionBatchQueryRow(t *testing.T) {
	type args struct {
		ctx    context.Context
		query  string
		params []interface{}
	}
	tests := []struct {
		name    string
		ds      DbSession
		args    args
		want    pgx.Row
		want1   func()
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, got1, err := tt.ds.SessionBatchQueryRow(tt.args.ctx, tt.args.query, tt.args.params...)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("DbSession.SessionBatchQueryRow(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.query, tt.args.params, err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("DbSession.SessionBatchQueryRow(%v, %v, %v) got = %v, want %v", tt.args.ctx, tt.args.query, tt.args.params, got, tt.want)
			// }
			// if !reflect.DeepEqual(got1, tt.want1) {
			// 	t.Errorf("DbSession.SessionBatchQueryRow(%v, %v, %v) got1 = %v, want %v", tt.args.ctx, tt.args.query, tt.args.params, got1, tt.want1)
			// }
		})
	}
}

func TestDbSession_SessionOpenBatch(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		ds   DbSession
		args args
		want *pgx.Batch
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ds.SessionOpenBatch(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DbSession.SessionOpenBatch(%v) = %v, want %v", tt.args.ctx, got, tt.want)
			}
		})
	}
}

func TestDbSession_SessionSendBatch(t *testing.T) {
	type args struct {
		ctx   context.Context
		batch *pgx.Batch
	}
	tests := []struct {
		name string
		ds   DbSession
		args args
		want pgx.BatchResults
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ds.SessionSendBatch(tt.args.ctx, tt.args.batch)

			if err == nil {
				t.Errorf("session send batch error %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DbSession.SessionSendBatch(%v, %v) = %v, want %v", tt.args.ctx, tt.args.batch, got, tt.want)
			}
		})
	}
}

func TestDatabase_QueryRows(t *testing.T) {
	type args struct {
		ctx              context.Context
		tx               *PoolTx
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
			if err := tt.db.QueryRows(tt.args.ctx, tt.args.tx, tt.args.protoStructSlice, tt.args.query, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Database.QueryRows(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.tx, tt.args.protoStructSlice, tt.args.query, tt.args.args, err, tt.wantErr)
			}
		})
	}
}

func Test_mapIntTypeToNullType(t *testing.T) {
	type args struct {
		t uint32
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
			if got := mapIntTypeToNullType(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapIntTypeToNullType(%v) = %v, want %v", tt.args.t, got, tt.want)
			}
		})
	}
}
