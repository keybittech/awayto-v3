package clients

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	_ "github.com/lib/pq"
)

var selectAdminRoleIdSQL = `SELECT id FROM dbtable_schema.roles WHERE name = 'Admin'`

func BenchmarkDbDefaultExec(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	reset(b)
	for c := 0; c < b.N; c++ {
		db.Client().Exec(selectAdminRoleIdSQL)
	}
}

func BenchmarkDbDefaultQuery(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		db.Client().QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
	}
}

func BenchmarkDbDefaultTx(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		tx, _ := db.Client().Begin()
		defer tx.Rollback()
		tx.QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
		tx.Commit()
	}
}

func BenchmarkDbTxExec(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	reset(b)
	for c := 0; c < b.N; c++ {
		var adminRoleId string
		db.TxExec(func(tx *sql.Tx) error {
			var txErr error
			txErr = tx.QueryRow(selectAdminRoleIdSQL).Scan(&adminRoleId)
			if txErr != nil {
				return util.ErrCheck(txErr)
			}

			return nil
		}, "worker", "", "")
	}
}

func BenchmarkDbSocketGetTopicMessageParticipants(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	participants := make(map[string]*types.SocketParticipant)
	topic := "test-topic"
	reset(b)
	b.StopTimer()
	for c := 0; c < b.N; c++ {
		err := db.TxExec(func(tx *sql.Tx) error {
			b.StartTimer()
			db.GetTopicMessageParticipants(tx, topic, participants)
			b.StopTimer()
			return nil
		}, "", "", "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

var workerString = "worker"

var execQuery = `
	SELECT sub, id
	FROM dbfunc_schema.session_query_13($1, $2, $3, $4)
	AS (sub uuid, id uuid)
`

var testQuery = `
	SELECT u.sub, r.id
	FROM dbtable_schema.users u
	JOIN dbtable_schema.roles r ON r.name = 'Admin'
	WHERE u.username = 'system_owner'
`

var testSub, testId string

func (db *Database) GetTestQuery() (bool, error) {
	err := db.Client().QueryRow(
		execQuery,
		workerString,
		emptyString,
		emptyString,
		testQuery,
	).Scan(&testSub, &testId)
	if err != nil {
		println(err.Error())
	}
	return true, nil
}

func BenchmarkDbSocketGetTestQuery(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()

	reset(b)
	for c := 0; c < b.N; c++ {
		db.GetTestQuery()
		// if b != nil {
		// 	println(fmt.Sprintf("error %v", b))
		// }
		//
		// println(fmt.Sprint(a))
	}
}

func BenchmarkDbSocketGetSocketAllowances(b *testing.B) {
	db := InitDatabase()
	defer db.DatabaseClient.Close()
	bookingId := integrationTest.Bookings[1].Id

	reset(b)
	for c := 0; c < b.N; c++ {
		b.StopTimer()
		user := integrationTest.TestUsers[int32(c%7)]
		b.StartTimer()
		db.GetSocketAllowances(user.UserSession, bookingId)
		// if b != nil {
		// 	println(fmt.Sprintf("error %v", b))
		// }
		//
		// println(fmt.Sprint(a))
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
