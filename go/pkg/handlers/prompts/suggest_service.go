package prompts

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var suggestServiceMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s",
	getSuggestionPrompt("gerund verbs performed for the purpose of ${prompt1}"),
	generateExample("gerund verbs performed for the purpose of offering educational services to community college students", "Tutoring|Advising|Consulting|Instruction|Mentoring"),
	generateExample("gerund verbs performed for the purpose of providing banking services to the local area", "Accounting|Financing|Securities|Financial Planning|Investing"),
)

var suggestServiceMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", suggestServiceMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: generateExample("gerund verbs performed for the purpose of ${prompt1}", "")},
}
