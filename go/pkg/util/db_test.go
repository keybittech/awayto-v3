package util

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
)

func TestWithPagination(t *testing.T) {
	type args struct {
		query    string
		page     int
		pageSize int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Empty query", args: args{"", 1, 10}, want: " LIMIT 10 OFFSET 0"},
		{name: "Regular query", args: args{"SELECT id FROM products WHERE category = 'test'", 3, 15}, want: "SELECT id FROM products WHERE category = 'test' LIMIT 15 OFFSET 30"},
		{name: "Negative page size", args: args{"SELECT id FROM users", 1, -5}, want: "SELECT id FROM users LIMIT -5 OFFSET 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithPagination(tt.args.query, tt.args.page, tt.args.pageSize); got != tt.want {
				t.Errorf("WithPagination() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkWithPagination(b *testing.B) {
	reset(b)
	for b.Loop() {
		_ = WithPagination("SELECT id FROM products WHERE category = 'test'", 3, 15)
	}
}

func TestNewBatchable(t *testing.T) {
	type args struct {
		pool     *pgxpool.Pool
		sub      string
		groupId  string
		roleBits int64
	}
	tests := []struct {
		name string
		args args
		want *Batchable
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := NewBatchable(tt.args.pool, tt.args.sub, tt.args.groupId, tt.args.roleBits); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewBatchable(%v, %v, %v, %v) = %v, want %v", tt.args.pool, tt.args.sub, tt.args.groupId, tt.args.roleBits, got, tt.want)
			// }
		})
	}
}

func TestBatchable_Reset(t *testing.T) {
	type args struct {
		knownOpSize []int32
	}
	tests := []struct {
		name string
		b    *Batchable
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.Reset(tt.args.knownOpSize...)
		})
	}
}

func TestBatchExec(t *testing.T) {
	type args struct {
		b      *Batchable
		query  string
		params []any
	}
	tests := []struct {
		name string
		args args
		want *pgconn.CommandTag
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BatchExec(tt.args.b, tt.args.query, tt.args.params...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchExec(%v, %v, %v) = %v, want %v", tt.args.b, tt.args.query, tt.args.params, got, tt.want)
			}
		})
	}
}

func TestBatchQuery(t *testing.T) {
	type args struct {
		b      *Batchable
		query  string
		params []any
	}
	tests := []struct {
		name string
		args args
		want *[]proto.Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BatchQuery[proto.Message](tt.args.b, tt.args.query, tt.args.params...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchQuery(%v, %v, %v) = %v, want %v", tt.args.b, tt.args.query, tt.args.params, got, tt.want)
			}
		})
	}
}

func TestBatchQueryRow(t *testing.T) {
	type args struct {
		b      *Batchable
		query  string
		params []any
	}
	tests := []struct {
		name string
		args args
		want **proto.Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BatchQueryRow[proto.Message](tt.args.b, tt.args.query, tt.args.params...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchQueryRow(%v, %v, %v) = %v, want %v", tt.args.b, tt.args.query, tt.args.params, got, tt.want)
			}
		})
	}
}

func TestBatchQueryMap(t *testing.T) {
	type args struct {
		b      *Batchable
		mapKey string
		query  string
		params []any
	}
	tests := []struct {
		name string
		args args
		want *map[string]*proto.Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BatchQueryMap[proto.Message](tt.args.b, tt.args.mapKey, tt.args.query, tt.args.params...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchQueryMap(%v, %v, %v, %v) = %v, want %v", tt.args.b, tt.args.mapKey, tt.args.query, tt.args.params, got, tt.want)
			}
		})
	}
}

func TestBatchable_Send(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		b    *Batchable
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.Send(tt.args.ctx)
		})
	}
}
