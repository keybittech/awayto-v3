package handlers

import (
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func TestHandlers_PostSiteFeedback(t *testing.T) {
	h, info, err := setupTestEnv()
	if err != nil {
		t.Fatal(util.ErrCheck(err))
	}
	defer info.Tx.Rollback(info.Ctx)

	type args struct {
		info ReqInfo
		data *types.PostSiteFeedbackRequest
	}
	tests := []struct {
		name     string
		h        *Handlers
		args     args
		want     *types.PostSiteFeedbackResponse
		wantErr  bool
		validErr bool
	}{
		{
			"Valid feedback submission",
			h,
			args{info, &types.PostSiteFeedbackRequest{Feedback: &types.IFeedback{FeedbackMessage: "test"}}},
			&types.PostSiteFeedbackResponse{Success: true},
			false,
			false,
		},
		{
			"Empty feedback submission",
			h,
			args{info, &types.PostSiteFeedbackRequest{Feedback: &types.IFeedback{FeedbackMessage: ""}}},
			nil,
			true,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := validateData(t, tt.args.data, tt.name, tt.validErr)
			if done {
				return
			}

			got, err := tt.h.PostSiteFeedback(tt.args.info, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostSiteFeedback(%v, %v) error = %v, wantErr %v", tt.args.info, tt.args.data, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostSiteFeedback(%v, %v) = %v, want %v", tt.args.info, tt.args.data, got, tt.want)
			}
		})
	}
}

func BenchmarkHandlers_PostSiteFeedback(b *testing.B) {
	h, info, err := setupTestEnv()
	if err != nil {
		b.Fatal(util.ErrCheck(err))
	}
	defer info.Tx.Rollback(info.Ctx)

	// Create a valid feedback request that we'll reuse for benchmarking
	validRequest := &types.PostSiteFeedbackRequest{
		Feedback: &types.IFeedback{
			FeedbackMessage: "This is a benchmark test feedback message",
		},
	}

	reset(b)

	for i := 0; i < b.N; i++ {
		response, err := h.PostSiteFeedback(info, validRequest)
		if err != nil {
			b.Fatal(util.ErrCheck(err))
		}
		if !response.Success {
			b.Fatal("Expected successful response")
		}
	}
}
