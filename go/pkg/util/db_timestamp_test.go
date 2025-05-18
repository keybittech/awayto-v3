// FROM https://github.com/manniwood/pgx-protobuf-timestamp
// Package protobufts is used to help pgx scan postgres timestamps
// into Google Protobuf type *timestamppb.Timestamp.
package util

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestTimestamp_ScanTimestamp(t *testing.T) {
	type args struct {
		v pgtype.Timestamp
	}
	tests := []struct {
		name    string
		ts      *Timestamp
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ts.ScanTimestamp(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Timestamp.ScanTimestamp(%v) error = %v, wantErr %v", tt.args.v, err, tt.wantErr)
			}
		})
	}
}

func TestTimestamp_TimestampValue(t *testing.T) {
	tests := []struct {
		name    string
		ts      *Timestamp
		want    pgtype.Timestamp
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ts.TimestampValue()
			if (err != nil) != tt.wantErr {
				t.Errorf("Timestamp.TimestampValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Timestamp.TimestampValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTryWrapTimestampEncodePlan(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name          string
		args          args
		wantPlan      pgtype.WrappedEncodePlanNextSetter
		wantNextValue interface{}
		wantOk        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlan, gotNextValue, gotOk := TryWrapTimestampEncodePlan(tt.args.value)
			if !reflect.DeepEqual(gotPlan, tt.wantPlan) {
				t.Errorf("TryWrapTimestampEncodePlan(%v) gotPlan = %v, want %v", tt.args.value, gotPlan, tt.wantPlan)
			}
			if !reflect.DeepEqual(gotNextValue, tt.wantNextValue) {
				t.Errorf("TryWrapTimestampEncodePlan(%v) gotNextValue = %v, want %v", tt.args.value, gotNextValue, tt.wantNextValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("TryWrapTimestampEncodePlan(%v) gotOk = %v, want %v", tt.args.value, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_wrapTimestampEncodePlan_SetNext(t *testing.T) {
	type args struct {
		next pgtype.EncodePlan
	}
	tests := []struct {
		name string
		plan *wrapTimestampEncodePlan
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plan.SetNext(tt.args.next)
		})
	}
}

func Test_wrapTimestampEncodePlan_Encode(t *testing.T) {
	type args struct {
		value interface{}
		buf   []byte
	}
	tests := []struct {
		name       string
		plan       *wrapTimestampEncodePlan
		args       args
		wantNewBuf []byte
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewBuf, err := tt.plan.Encode(tt.args.value, tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("wrapTimestampEncodePlan.Encode(%v, %v) error = %v, wantErr %v", tt.args.value, tt.args.buf, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNewBuf, tt.wantNewBuf) {
				t.Errorf("wrapTimestampEncodePlan.Encode(%v, %v) = %v, want %v", tt.args.value, tt.args.buf, gotNewBuf, tt.wantNewBuf)
			}
		})
	}
}

func TestTryWrapTimestampScanPlan(t *testing.T) {
	type args struct {
		target interface{}
	}
	tests := []struct {
		name        string
		args        args
		wantPlan    pgtype.WrappedScanPlanNextSetter
		wantNextDst interface{}
		wantOk      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlan, gotNextDst, gotOk := TryWrapTimestampScanPlan(tt.args.target)
			if !reflect.DeepEqual(gotPlan, tt.wantPlan) {
				t.Errorf("TryWrapTimestampScanPlan(%v) gotPlan = %v, want %v", tt.args.target, gotPlan, tt.wantPlan)
			}
			if !reflect.DeepEqual(gotNextDst, tt.wantNextDst) {
				t.Errorf("TryWrapTimestampScanPlan(%v) gotNextDst = %v, want %v", tt.args.target, gotNextDst, tt.wantNextDst)
			}
			if gotOk != tt.wantOk {
				t.Errorf("TryWrapTimestampScanPlan(%v) gotOk = %v, want %v", tt.args.target, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_wrapTimestampScanPlan_SetNext(t *testing.T) {
	type args struct {
		next pgtype.ScanPlan
	}
	tests := []struct {
		name string
		plan *wrapTimestampScanPlan
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plan.SetNext(tt.args.next)
		})
	}
}

func Test_wrapTimestampScanPlan_Scan(t *testing.T) {
	type args struct {
		src []byte
		dst interface{}
	}
	tests := []struct {
		name    string
		plan    *wrapTimestampScanPlan
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.plan.Scan(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("wrapTimestampScanPlan.Scan(%v, %v) error = %v, wantErr %v", tt.args.src, tt.args.dst, err, tt.wantErr)
			}
		})
	}
}

func TestTimestampCodec_DecodeValue(t *testing.T) {
	type args struct {
		tm     *pgtype.Map
		oid    uint32
		format int16
		src    []byte
	}
	tests := []struct {
		name    string
		tr      TimestampCodec
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.DecodeValue(tt.args.tm, tt.args.oid, tt.args.format, tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("TimestampCodec.DecodeValue(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.tm, tt.args.oid, tt.args.format, tt.args.src, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TimestampCodec.DecodeValue(%v, %v, %v, %v) = %v, want %v", tt.args.tm, tt.args.oid, tt.args.format, tt.args.src, got, tt.want)
			}
		})
	}
}

func TestRegisterTimestamp(t *testing.T) {
	type args struct {
		tm *pgtype.Map
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterTimestamp(tt.args.tm)
		})
	}
}
