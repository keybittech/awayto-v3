package handlers

import (
	"context"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostSchedule(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostScheduleRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostScheduleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostSchedule(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostSchedule(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostSchedule(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PostScheduleBrackets(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PostScheduleBracketsRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostScheduleBracketsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostScheduleBrackets(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostScheduleBrackets(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostScheduleBrackets(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_PatchSchedule(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.PatchScheduleRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PatchScheduleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PatchSchedule(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PatchSchedule(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PatchSchedule(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetSchedules(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetSchedulesRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetSchedulesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetSchedules(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetSchedules(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetSchedules(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_GetScheduleById(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.GetScheduleByIdRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.GetScheduleByIdResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.GetScheduleById(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetScheduleById(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetScheduleById(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DeleteSchedule(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DeleteScheduleRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DeleteScheduleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DeleteSchedule(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteSchedule(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DeleteSchedule(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_DisableSchedule(t *testing.T) {
	type args struct {
		info ReqInfo
		data *types.DisableScheduleRequest
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.DisableScheduleResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.DisableSchedule(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DisableSchedule(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.DisableSchedule(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func TestHandlers_HandleExistingBrackets(t *testing.T) {
	type args struct {
		ctx                context.Context
		existingBracketIds []string
		brackets           map[string]*types.IScheduleBracket
		info               ReqInfo
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.h.HandleExistingBrackets(tt.args.ctx, tt.args.existingBracketIds, tt.args.brackets, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Handlers.HandleExistingBrackets(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.existingBracketIds, tt.args.brackets, tt.args.info, err, tt.wantErr)
			}
		})
	}
}

func TestHandlers_InsertNewBrackets(t *testing.T) {
	type args struct {
		ctx         context.Context
		scheduleId  string
		newBrackets map[string]*types.IScheduleBracket
		info        ReqInfo
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.h.InsertNewBrackets(tt.args.ctx, tt.args.scheduleId, tt.args.newBrackets, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Handlers.InsertNewBrackets(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.scheduleId, tt.args.newBrackets, tt.args.info, err, tt.wantErr)
			}
		})
	}
}

func Test_handleDeletedBrackets(t *testing.T) {
	type args struct {
		ctx                context.Context
		scheduleId         string
		existingBracketIds []string
		info               ReqInfo
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
			if err := handleDeletedBrackets(tt.args.ctx, tt.args.scheduleId, tt.args.existingBracketIds, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("handleDeletedBrackets(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.scheduleId, tt.args.existingBracketIds, tt.args.info, err, tt.wantErr)
			}
		})
	}
}

func Test_disableAndDeleteBrackets(t *testing.T) {
	type args struct {
		ctx               context.Context
		bracketsToDisable []string
		bracketsToDelete  []string
		info              ReqInfo
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
			if err := disableAndDeleteBrackets(tt.args.ctx, tt.args.bracketsToDisable, tt.args.bracketsToDelete, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("disableAndDeleteBrackets(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.bracketsToDisable, tt.args.bracketsToDelete, tt.args.info, err, tt.wantErr)
			}
		})
	}
}
