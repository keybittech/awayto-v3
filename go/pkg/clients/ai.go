package clients

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

type AiPrompts map[types.IPrompts]interface{}

type Ai struct {
	Model   string
	Prompts AiPrompts
}

func InitAi() IAi {

	aiPrompts := make(AiPrompts)
	aiPrompts[types.IPrompts_CONVERT_PURPOSE] = convertPurposeMessages
	aiPrompts[types.IPrompts_SUGGEST_FEATURE] = suggestFeatureMessages
	aiPrompts[types.IPrompts_SUGGEST_ROLE] = suggestRoleMessages
	aiPrompts[types.IPrompts_SUGGEST_SERVICE] = suggestServiceMessages
	aiPrompts[types.IPrompts_SUGGEST_TIER] = suggestTierMessages

	aic := &Ai{
		Model:   openai.GPT3Dot5Turbo,
		Prompts: aiPrompts,
	}

	return aic
}

func (ai *Ai) GetPromptResponse(ctx context.Context, promptParts []string, promptType types.IPrompts) string {

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		util.ErrCheck(errors.New("No OpenAI Key"))
		return ""
	}

	client := openai.NewClient(openAIKey)

	promptTokens := make(map[string]string)

	promptTemplate := ai.Prompts[promptType]

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
			Model:  ai.Model,
			Prompt: content,
		})

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
			Model:    ai.Model,
			Messages: messages,
		})

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

func GetSuggestionPrompt(prompt string) string {
	return `Generate 5 ` + prompt + `; Result is 1-3 words separated by |. Here are some examples: `
}

func GenerateExample(prompt string, result string) string {
	return `Phrase: ` + prompt + `\nResult: ` + result
}

func HasSimilarKey(obj map[string]interface{}, regex regexp.Regexp) bool {
	has := false
	for key := range obj {
		if regex.Match([]byte(key)) {
			has = true
			break
		}
	}
	return has
}

var convertPurposeMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `You respond with a single gerund phrase, 5 to 8 words, which will complete the user's sentence. For example, "walking in the sunshine with you".`},
	{Role: openai.ChatMessageRoleAssistant, Content: `Give me an incomplete sentence that I can complete with a gerund phrase. For example, if you said "Our favorite past time is", I might respond "walking in the sunshine with you"`},
	{Role: openai.ChatMessageRoleUser, Content: `An organization named "${prompt1}" is interested in "${prompt2}" and their mission statement is`},
}

var suggestFeatureMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s\n%s\n%s",
	GetSuggestionPrompt("features of ${prompt1}"),
	GenerateExample("features of ENGL 1010 writing tutoring", "Feedback|Revisions|Brainstorming|Discussion"),
	GenerateExample("features of Standard gym membership", "Full Gym Equipment|Limited Training|Half-Day Access"),
	GenerateExample("features of Pro web hosting service", "Unlimited Sites|Unlimited Storage|1TB Bandwidth|Daily Backups"),
	GenerateExample("features of professional photography service", "Next-Day Prints|High-quality digital photos|Retouching and editing|Choice of location|Choice of outfit changes"),
)

var suggestFeatureMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", suggestFeatureMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: GenerateExample("features of ${prompt1}", "")},
}

var roleMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s",
	GetSuggestionPrompt("role names for a group named ${prompt1} which is interested in ${prompt2}"),
	GenerateExample("role names for a group named writing center which is interested in consulting on writing", "Tutor|Student|Advisor|Administrator|Consultant"),
	GenerateExample("role names for a group named city maintenance department which is interested in maintaining the facilities in the city", "Dispatcher|Engineer|Administrator|Technician|Manager"),
)

var suggestRoleMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", roleMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: GenerateExample(`role names for a group named "${prompt1}" which is interested in ${prompt2}`, "")},
}

var suggestServiceMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s",
	GetSuggestionPrompt("gerund verbs performed for the purpose of ${prompt1}"),
	GenerateExample("gerund verbs performed for the purpose of offering educational services to community college students", "Tutoring|Advising|Consulting|Instruction|Mentoring"),
	GenerateExample("gerund verbs performed for the purpose of providing banking services to the local area", "Accounting|Financing|Securities|Financial Planning|Investing"),
)

var suggestServiceMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", suggestServiceMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: GenerateExample("gerund verbs performed for the purpose of ${prompt1}", "")},
}

var suggestTierMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s\n%s\n%s\n%s\n%s",
	GetSuggestionPrompt("service level names for ${prompt1}"),
	GenerateExample("service level names for a generic service", "Small|Medium|Large"),
	GenerateExample("service level names for writing tutoring at a school writing center", "WRI 1010|WRI 1020|WRI 2010|WRI 2020|WRI 3010"),
	GenerateExample("service level names for streaming at a web media platform", "Basic|Standard|Premium"),
	GenerateExample("service level names for advising at a school learning center", "ENG 1010|WRI 1010|MAT 1010|SCI 1010|HIS 1010"),
	GenerateExample("service level names for travelling on an airline service", "Economy|Business|First Class"),
	GenerateExample("service level names for reading tutoring at a school reading center", "ESL 900|ESL 990|ENG 1010|ENG 1020|ENG 2010"),
)

var suggestTierMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", suggestTierMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: GenerateExample("service level names for ${prompt1}", "")},
}
