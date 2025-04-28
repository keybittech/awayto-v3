package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostGroupUserSchedule(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostGroupUserScheduleRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostGroupUserScheduleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostGroupUserSchedule(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostGroupUserSchedule(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostGroupUserSchedule(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupUserSchedules(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupUserSchedulesRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupUserSchedulesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupUserSchedules(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupUserSchedules(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupUserSchedules(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupUserScheduleStubs(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupUserScheduleStubsRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupUserScheduleStubsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupUserScheduleStubs(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupUserScheduleStubs(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupUserScheduleStubs(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetGroupUserScheduleStubReplacement(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetGroupUserScheduleStubReplacementRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetGroupUserScheduleStubReplacementResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetGroupUserScheduleStubReplacement(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetGroupUserScheduleStubReplacement(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetGroupUserScheduleStubReplacement(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchGroupUserScheduleStubReplacement(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PatchGroupUserScheduleStubReplacementRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchGroupUserScheduleStubReplacementResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchGroupUserScheduleStubReplacement(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchGroupUserScheduleStubReplacement(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchGroupUserScheduleStubReplacement(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteGroupUserScheduleByUserScheduleId(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DeleteGroupUserScheduleByUserScheduleIdRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteGroupUserScheduleByUserScheduleIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteGroupUserScheduleByUserScheduleId(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteGroupUserScheduleByUserScheduleId(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteGroupUserScheduleByUserScheduleId(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}
