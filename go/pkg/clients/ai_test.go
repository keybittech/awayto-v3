package clients

import (
	"context"
	"reflect"
	"regexp"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestInitAi(t *testing.T) {
	tests := []struct {
		name string
		want interfaces.IAi
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitAi(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitAi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAi_GetPromptResponse(t *testing.T) {
	type args struct {
		ctx         context.Context
		promptParts []string
		promptType  types.IPrompts
	}
	tests := []struct {
		name    string
		ai      *Ai
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ai.GetPromptResponse(tt.args.ctx, tt.args.promptParts, tt.args.promptType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ai.GetPromptResponse(%v, %v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.promptParts, tt.args.promptType, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ai.GetPromptResponse(%v, %v, %v) = %v, want %v", tt.args.ctx, tt.args.promptParts, tt.args.promptType, got, tt.want)
			}
		})
	}
}

func TestGetSuggestionPrompt(t *testing.T) {
	type args struct {
		prompt string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSuggestionPrompt(tt.args.prompt); got != tt.want {
				t.Errorf("GetSuggestionPrompt(%v) = %v, want %v", tt.args.prompt, got, tt.want)
			}
		})
	}
}

func TestGenerateExample(t *testing.T) {
	type args struct {
		prompt string
		result string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateExample(tt.args.prompt, tt.args.result); got != tt.want {
				t.Errorf("GenerateExample(%v, %v) = %v, want %v", tt.args.prompt, tt.args.result, got, tt.want)
			}
		})
	}
}

func TestHasSimilarKey(t *testing.T) {
	type args struct {
		obj   map[string]interface{}
		regex regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasSimilarKey(tt.args.obj, tt.args.regex); got != tt.want {
				t.Errorf("HasSimilarKey(%v, %v) = %v, want %v", tt.args.obj, tt.args.regex, got, tt.want)
			}
		})
	}
}
