package prompts

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var openaiModel = openai.GPT3Dot5Turbo

type AiPrompts map[types.IPrompts]interface{}

var aiPrompts AiPrompts = make(AiPrompts)

func init() {
	aiPrompts[types.IPrompts_CONVERT_PURPOSE] = convertPurposeMessages
	aiPrompts[types.IPrompts_SUGGEST_FEATURE] = suggestFeatureMessages
	aiPrompts[types.IPrompts_SUGGEST_ROLE] = suggestRoleMessages
	aiPrompts[types.IPrompts_SUGGEST_SERVICE] = suggestServiceMessages
	aiPrompts[types.IPrompts_SUGGEST_TIER] = suggestTierMessages
}

func GetPromptResponse(ctx context.Context, promptParts []string, promptType types.IPrompts) string {

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		util.ErrCheck(errors.New("No OpenAI Key"))
		return ""
	}

	client := openai.NewClient(openAIKey)

	promptTokens := make(map[string]string)

	promptTemplate := aiPrompts[promptType]

	for i, prompt := range promptParts {
		tokenKey := fmt.Sprintf("${prompt%d}", i+1)
		promptTokens[tokenKey] = prompt
	}

	switch promptTemplate.(type) {
	case string:
		content := promptTemplate.(string)
		for k, v := range promptTokens {
			content = strings.ReplaceAll(content, k, v)
		}

		resp, err := client.CreateCompletion(ctx, openai.CompletionRequest{
			Model:  openaiModel,
			Prompt: content,
		})

		util.Debug("str res %+v\n", resp.Choices)

		if err != nil {
			util.ErrCheck(err)
			return ""
		}

		return resp.Choices[0].Text

	case []openai.ChatCompletionMessage:
		messages := []openai.ChatCompletionMessage{}
		for _, message := range promptTemplate.([]openai.ChatCompletionMessage) {
			content := message.Content
			for k, v := range promptTokens {
				content = strings.ReplaceAll(content, k, v)
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    message.Role,
				Content: content,
			})
		}

		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    openaiModel,
			Messages: messages,
		})

		util.Debug("messages res %+v\n", resp.Choices)

		if err != nil {
			util.ErrCheck(err)
			return ""
		}

		return resp.Choices[0].Message.Content

	default:
		err := errors.New("unsupported prompt type")
		util.ErrCheck(err)
		return ""
	}
}

func getSuggestionPrompt(prompt string) string {
	return `Generate 5 ` + prompt + `; Result is 1-3 words separated by |. Here are some examples: `
}

func generateExample(prompt string, result string) string {
	return `Phrase: ` + prompt + `\nResult: ` + result
}

func hasSimilarKey(obj map[string]interface{}, regex regexp.Regexp) bool {
	has := false
	for key := range obj {
		if regex.Match([]byte(key)) {
			has = true
			break
		}
	}
	return has
}
