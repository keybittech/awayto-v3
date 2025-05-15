package util

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	defaultMaxOps          int32 = 4
	emptyString                  = ""
	emptyInteger                 = 0
	setSessionVariablesSQL       = `SELECT dbfunc_schema.set_session_vars($1::VARCHAR, $2::VARCHAR, $3::INTEGER, $4::VARCHAR)`
)

func appendRowsToMap[T any, M ~map[string]T](resultMap M, mapKeyTarget string, rows pgx.Rows, fn pgx.RowToFunc[T]) (M, error) {
	defer rows.Close()

	for rows.Next() {
		raw, err := fn(rows)
		if err != nil {
			return nil, err
		}

		value, ok := any(raw).(proto.Message)
		if !ok {
			return nil, ErrCheck(fmt.Errorf("result set while mapping is not a proto message, got %T", value))
		}

		var mapKey string
		value.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if fd.Name() == protoreflect.Name(mapKeyTarget) {
				mapKey = v.String()
				return false
			}
			return true
		})

		if mapKey == "" {
			return nil, ErrCheck(fmt.Errorf("the map key target %s was not found in the result set", mapKeyTarget))
		}

		resultMap[mapKey] = raw
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resultMap, nil
}

func collectRowsToMap[T any](rows pgx.Rows, keyTarget string, fn pgx.RowToFunc[T]) (map[string]T, error) {
	initMap := make(map[string]T)
	return appendRowsToMap(initMap, keyTarget, rows, fn)
}

type batchOp func(pgx.BatchResults) (any, error)

func batchOpExec(br pgx.BatchResults) (any, error) {
	return br.Exec()
}

func batchOpQuery[T any](br pgx.BatchResults) (any, error) {
	rows, err := br.Query()
	if err != nil {
		return nil, ErrCheck(err)
	}

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[T])
}

func batchOpQueryMap[T any](mapKeyTarget string) func(br pgx.BatchResults) (any, error) {
	return func(br pgx.BatchResults) (any, error) {
		rows, err := br.Query()
		if err != nil {
			return nil, ErrCheck(err)
		}

		return collectRowsToMap(rows, mapKeyTarget, pgx.RowToAddrOfStructByNameLax[T])
	}
}

func batchOpQueryRow[T any](br pgx.BatchResults) (any, error) {
	rows, err := br.Query()
	if err != nil {
		return nil, ErrCheck(err)
	}

	return pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[T])
}

type Batchable struct {
	ops          []batchOp
	outerSlice   []any
	Sub, GroupId string
	pool         *pgxpool.Pool
	batch        *pgx.Batch
	RoleBits     int64
}

// Open a batch with the intention of adding multiple queries
func NewBatchable(pool *pgxpool.Pool, sub, groupId string, roleBits int64) *Batchable {
	b := &Batchable{
		Sub:      sub,
		GroupId:  groupId,
		RoleBits: roleBits,
		pool:     pool,
	}
	b.Reset()
	return b
}

// Use a default of defaultMaxOps (4) batch ops because we have first and last op as set session ops and
// go would natrually allocate 4 spaces anyway if we didn't allocate here
func (b *Batchable) Reset(knownOpSize ...int32) {
	var opSize int32 = defaultMaxOps
	if knownOpSize != nil {
		opSize = knownOpSize[0]
	}
	b.batch = &pgx.Batch{}
	b.batch.Queue(setSessionVariablesSQL, b.Sub, b.GroupId, b.RoleBits, emptyString)

	b.ops = make([]batchOp, 0, opSize)

	b.ops = append(b.ops, batchOpExec)

	b.outerSlice = make([]any, 0, opSize)
}

func (b *Batchable) BatchOp(op batchOp, query string, params ...any) int {
	b.batch.Queue(query, params...)
	b.ops = append(b.ops, op)
	b.outerSlice = append(b.outerSlice, nil)
	return len(b.outerSlice) - 1
}

func BatchExec[T any](b *Batchable, query string, params ...any) *pgconn.CommandTag {
	var result pgconn.CommandTag
	resultPtr := &result
	idx := b.BatchOp(batchOpExec, query, params...)
	b.outerSlice[idx] = resultPtr
	return resultPtr
}

func BatchQuery[T any](b *Batchable, query string, params ...any) *[]*T {
	var result []*T
	resultPtr := &result
	idx := b.BatchOp(batchOpQuery[T], query, params...)
	b.outerSlice[idx] = resultPtr
	return resultPtr
}

func BatchQueryRow[T any](b *Batchable, query string, params ...any) **T {
	var result *T
	resultPtr := &result
	idx := b.BatchOp(batchOpQueryRow[T], query, params...)
	b.outerSlice[idx] = resultPtr
	return resultPtr
}

func BatchQueryMap[T any](b *Batchable, mapKey, query string, params ...any) *map[string]*T {
	var result map[string]*T
	resultPtr := &result
	idx := b.BatchOp(batchOpQueryMap[T](mapKey), query, params...)
	b.outerSlice[idx] = resultPtr
	return resultPtr
}

// Close a batch opened with NewBatchable. The caller needs to infer types of response slice
// Results will exclude the first and last operation, which correspond to set session ops
func (b *Batchable) Send(ctx context.Context) error {
	// Unset session vars after all queries
	b.batch.Queue(setSessionVariablesSQL, emptyString, emptyString, emptyInteger, emptyString)
	b.ops = append(b.ops, batchOpExec)

	br := b.pool.SendBatch(ctx, b.batch)

	var opErr error
	opIdx := 0
	for i, op := range b.ops {
		res, err := op(br)
		if err != nil {
			opErr = ErrCheck(err)
			break
		}

		if i > 0 && i < len(b.ops)-1 {
			resultPtr := b.outerSlice[opIdx]
			reflect.ValueOf(resultPtr).Elem().Set(reflect.ValueOf(res))
			opIdx++
		}
	}

	closeErr := br.Close()
	if opErr != nil {
		if closeErr != nil {
			ErrorLog.Println(ErrCheck(closeErr))
		}
		return ErrCheck(opErr)
	}

	if closeErr != nil {
		return ErrCheck(closeErr)
	}

	return nil
}

func WithPagination(query string, page, pageSize int) string {
	return query + " LIMIT " + strconv.Itoa(pageSize) + " OFFSET " + strconv.Itoa((page-1)*pageSize)
}
