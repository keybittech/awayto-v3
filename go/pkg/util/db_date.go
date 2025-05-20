package util

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const dateFormat = "2006-01-02"

type Date string

func (ts *Date) ScanDate(v pgtype.Date) error {
	if !v.Valid {
		ts = nil
		return nil
	}
	// v.Time is a time.Time
	*ts = Date(v.Time.Format(dateFormat))
	return nil
}

func (ts *Date) DateValue() (pgtype.Date, error) {
	var t time.Time
	t, err := time.Parse(dateFormat, *(*string)(ts))
	if err != nil {
		return pgtype.Date{Time: t, InfinityModifier: pgtype.Finite, Valid: false}, ErrCheck(err)
	}
	return pgtype.Date{Time: t, InfinityModifier: pgtype.Finite, Valid: true}, nil
}

func TryWrapDateEncodePlan(value any) (plan pgtype.WrappedEncodePlanNextSetter, nextValue any, ok bool) {
	switch value := value.(type) {
	case *string:
		return &wrapDateEncodePlan{}, (*Date)(value), true
	}

	return nil, nil, false
}

type wrapDateEncodePlan struct {
	next pgtype.EncodePlan
}

func (plan *wrapDateEncodePlan) SetNext(next pgtype.EncodePlan) {
	plan.next = next
}

func (plan *wrapDateEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return plan.next.Encode((*Date)(value.(*string)), buf)
}

func TryWrapDateScanPlan(target any) (plan pgtype.WrappedScanPlanNextSetter, nextDst any, ok bool) {
	switch target := target.(type) {
	case *string:
		return &wrapDateScanPlan{}, (*Date)(target), true
	}

	return nil, nil, false
}

type wrapDateScanPlan struct {
	next pgtype.ScanPlan
}

func (plan *wrapDateScanPlan) SetNext(next pgtype.ScanPlan) {
	plan.next = next
}

func (plan *wrapDateScanPlan) Scan(src []byte, dst any) error {
	return plan.next.Scan(src, (*Date)(dst.(*string)))
}

type DateCodec struct {
	pgtype.DateCodec
}

func (DateCodec) DecodeValue(tm *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		return nil, nil
	}

	var target *string
	scanPlan := tm.PlanScan(oid, format, &target)
	if scanPlan == nil {
		return nil, fmt.Errorf("PlanScan did not find a plan for date")
	}

	err := scanPlan.Scan(src, &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func RegisterDate(tm *pgtype.Map) {
	tm.TryWrapEncodePlanFuncs = append([]pgtype.TryWrapEncodePlanFunc{TryWrapDateEncodePlan}, tm.TryWrapEncodePlanFuncs...)
	tm.TryWrapScanPlanFuncs = append([]pgtype.TryWrapScanPlanFunc{TryWrapDateScanPlan}, tm.TryWrapScanPlanFuncs...)

	tm.RegisterType(&pgtype.Type{
		Name:  "date",
		OID:   pgtype.DateOID,
		Codec: &DateCodec{},
	})
}
