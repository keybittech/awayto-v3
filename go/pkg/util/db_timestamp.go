// FROM https://github.com/manniwood/pgx-protobuf-timestamp
// Package protobufts is used to help pgx scan postgres timestamps
// into Google Protobuf type *timestamppb.Timestamp.
package util

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const timestampFormat = "2006-01-02T15:04:05Z07:00"

type Timestamp string

// ScanTimestamp implements pgxtype.TimestampScanner interface
func (ts *Timestamp) ScanTimestamp(v pgtype.Timestamp) error {
	if !v.Valid {
		ts = nil
		return nil
	}
	// v.Time is a time.Time
	*ts = Timestamp(v.Time.Format(timestampFormat))
	return nil
}

// TimestampValue implements pgxtype.TimestampValuer interface
func (ts *Timestamp) TimestampValue() (pgtype.Timestamp, error) {
	var t time.Time
	t, err := time.Parse(timestampFormat, *(*string)(ts))
	if err != nil {
		return pgtype.Timestamp{Time: t, InfinityModifier: pgtype.Finite, Valid: false}, ErrCheck(err)
	}
	return pgtype.Timestamp{Time: t, InfinityModifier: pgtype.Finite, Valid: true}, nil
}

// pgxtype.TryWrapEncodePlanFunc is this type of function:
// type TryWrapEncodePlanFunc func(value any) (plan WrappedEncodePlanNextSetter, nextValue any, ok bool)
func TryWrapTimestampEncodePlan(value any) (plan pgtype.WrappedEncodePlanNextSetter, nextValue any, ok bool) {
	switch value := value.(type) {
	case *string:
		return &wrapTimestampEncodePlan{}, (*Timestamp)(value), true
	}

	return nil, nil, false
}

type wrapTimestampEncodePlan struct {
	next pgtype.EncodePlan
}

func (plan *wrapTimestampEncodePlan) SetNext(next pgtype.EncodePlan) {
	plan.next = next
}

func (plan *wrapTimestampEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return plan.next.Encode((*Timestamp)(value.(*string)), buf)
}

// pgxtype.TryWrapScanPlanFunc is this type of function:
// type TryWrapScanPlanFunc func(target any) (plan WrappedScanPlanNextSetter, nextTarget any, ok bool)

func TryWrapTimestampScanPlan(target any) (plan pgtype.WrappedScanPlanNextSetter, nextDst any, ok bool) {
	switch target := target.(type) {
	case *string:
		return &wrapTimestampScanPlan{}, (*Timestamp)(target), true
	}

	return nil, nil, false
}

type wrapTimestampScanPlan struct {
	next pgtype.ScanPlan
}

func (plan *wrapTimestampScanPlan) SetNext(next pgtype.ScanPlan) {
	plan.next = next
}

func (plan *wrapTimestampScanPlan) Scan(src []byte, dst any) error {
	return plan.next.Scan(src, (*Timestamp)(dst.(*string)))
}

// TimestampCodec embeds pgtype.TimestampCodec, which implements pgtype.Codec interface
type TimestampCodec struct {
	pgtype.TimestampCodec
}

// We only need to override the behavior of pgtype.TimestampCodec.DecodeValue();
// the other methods that satisfy pgtype.Codec are left implemented by pgtype.TimestampCodec
func (TimestampCodec) DecodeValue(tm *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		return nil, nil
	}

	var target *string
	scanPlan := tm.PlanScan(oid, format, &target)
	if scanPlan == nil {
		return nil, fmt.Errorf("PlanScan did not find a plan")
	}

	err := scanPlan.Scan(src, &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

/*
Here is the implementation of pgtype.TimestampCodec.DecodeValue():
func (c TimestampCodec) DecodeValue(m *Map, oid uint32, format int16, src []byte) (any, error) {
  if src == nil {
    return nil, nil
  }

  var ts Timestamp
  err := codecScan(c, m, oid, format, src, &ts)
  if err != nil {
    return nil, err
  }

  if ts.InfinityModifier != Finite {
    return ts.InfinityModifier, nil
  }

  return ts.Time, nil
}
*/

// Register registers the github.com/gofrs/uuid integration with a pgtype.Map.
func RegisterTimestamp(tm *pgtype.Map) {
	tm.TryWrapEncodePlanFuncs = append([]pgtype.TryWrapEncodePlanFunc{TryWrapTimestampEncodePlan}, tm.TryWrapEncodePlanFuncs...)
	tm.TryWrapScanPlanFuncs = append([]pgtype.TryWrapScanPlanFunc{TryWrapTimestampScanPlan}, tm.TryWrapScanPlanFuncs...)

	tm.RegisterType(&pgtype.Type{
		Name:  "timestamp",
		OID:   pgtype.TimestampOID,
		Codec: &TimestampCodec{},
	})

	tm.RegisterType(&pgtype.Type{
		Name:  "timestamptz",
		OID:   pgtype.TimestamptzOID,
		Codec: &TimestampCodec{},
	})
}
