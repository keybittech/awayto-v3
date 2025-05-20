package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	microsecondsPerMinuteInternal = 60 * 1_000_000
	microsecondsPerHourInternal   = 60 * microsecondsPerMinuteInternal
)

var (
	simpleDurationRegexForPgInterval = regexp.MustCompile(`^P(?:([0-9]+)D)?(?:T(?:([0-9]+)H)?(?:([0-9]+)M)?)?$`)
	invalidInterval                  = pgtype.Interval{Valid: false}
)

type Interval string

func (ts *Interval) ScanInterval(v pgtype.Interval) error {
	if !v.Valid {
		ts = nil
		return nil
	}

	var sb strings.Builder
	sb.WriteString("P")

	var hasComponent bool

	if v.Days > 0 {
		sb.WriteString(strconv.Itoa(int(v.Days)))
		sb.WriteString("D")
		hasComponent = true
	}

	if v.Microseconds > 0 {
		sb.WriteString("T")

		var hasTimeComponent bool

		hours := v.Microseconds / microsecondsPerHourInternal
		if hours > 0 {
			sb.WriteString(strconv.FormatInt(hours, 10))
			sb.WriteString("M")
			hasTimeComponent = true
		}

		minutes := (v.Microseconds % microsecondsPerHourInternal) / microsecondsPerMinuteInternal
		if minutes > 0 {
			sb.WriteString(strconv.FormatInt(minutes, 10))
			sb.WriteString("M")
			hasTimeComponent = true
		}

		if !hasTimeComponent {
			sb.WriteString("0M")
		}
		hasComponent = true
	}

	if !hasComponent {
		sb.WriteString("0D")
	}
	*ts = Interval(sb.String())
	return nil
}

func (ts *Interval) IntervalValue() (pgtype.Interval, error) {
	matches := simpleDurationRegexForPgInterval.FindStringSubmatch(*(*string)(ts))

	if matches == nil || (matches[1] == "" && matches[2] == "" && matches[3] == "") {
		return pgtype.Interval{Valid: false}, fmt.Errorf("invalid or empty duration format: %s", *(*string)(ts))
	}

	var daysVal int32
	var dv, hoursVal, minutesVal int64
	var err error

	if matches[1] != "" {
		dv, err = strconv.ParseInt(matches[1], 10, 32)
		if err != nil {
			return invalidInterval, err
		}
		daysVal, err = I64to32(dv)
		if err != nil {
			return invalidInterval, err
		}
	}
	if matches[2] != "" {
		hoursVal, _ = strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			return invalidInterval, err
		}
	}
	if matches[3] != "" {
		minutesVal, _ = strconv.ParseInt(matches[3], 10, 64)
		if err != nil {
			return invalidInterval, err
		}
	}

	return pgtype.Interval{
		Days:         daysVal,
		Microseconds: (hoursVal*3600 + minutesVal*60) * 1_000_000,
		Valid:        true,
	}, nil
}

func TryWrapIntervalEncodePlan(value any) (plan pgtype.WrappedEncodePlanNextSetter, nextValue any, ok bool) {
	switch value := value.(type) {
	case *string:
		return &wrapIntervalEncodePlan{}, (*Interval)(value), true
	}

	return nil, nil, false
}

type wrapIntervalEncodePlan struct {
	next pgtype.EncodePlan
}

func (plan *wrapIntervalEncodePlan) SetNext(next pgtype.EncodePlan) {
	plan.next = next
}

func (plan *wrapIntervalEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return plan.next.Encode((*Interval)(value.(*string)), buf)
}

func TryWrapIntervalScanPlan(target any) (plan pgtype.WrappedScanPlanNextSetter, nextDst any, ok bool) {
	switch target := target.(type) {
	case *string:
		return &wrapIntervalScanPlan{}, (*Interval)(target), true
	}

	return nil, nil, false
}

type wrapIntervalScanPlan struct {
	next pgtype.ScanPlan
}

func (plan *wrapIntervalScanPlan) SetNext(next pgtype.ScanPlan) {
	plan.next = next
}

func (plan *wrapIntervalScanPlan) Scan(src []byte, dst any) error {
	return plan.next.Scan(src, (*Interval)(dst.(*string)))
}

type IntervalCodec struct {
	pgtype.IntervalCodec
}

func (IntervalCodec) DecodeValue(tm *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		return nil, nil
	}

	var target *string
	scanPlan := tm.PlanScan(oid, format, &target)
	if scanPlan == nil {
		return nil, fmt.Errorf("PlanScan did not find a plan for interval")
	}

	err := scanPlan.Scan(src, &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func RegisterInterval(tm *pgtype.Map) {
	tm.TryWrapEncodePlanFuncs = append([]pgtype.TryWrapEncodePlanFunc{TryWrapIntervalEncodePlan}, tm.TryWrapEncodePlanFuncs...)
	tm.TryWrapScanPlanFuncs = append([]pgtype.TryWrapScanPlanFunc{TryWrapIntervalScanPlan}, tm.TryWrapScanPlanFuncs...)

	tm.RegisterType(&pgtype.Type{
		Name:  "interval",
		OID:   pgtype.IntervalOID,
		Codec: &IntervalCodec{},
	})
}
