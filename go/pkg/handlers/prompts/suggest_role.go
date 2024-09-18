package prompts

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var roleMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s",
	getSuggestionPrompt("role names for a group named ${prompt1} which is interested in ${prompt2}"),
	generateExample("role names for a group named writing center which is interested in consulting on writing", "Tutor|Student|Advisor|Administrator|Consultant"),
	generateExample("role names for a group named city maintenance department which is interested in maintaining the facilities in the city", "Dispatcher|Engineer|Administrator|Technician|Manager"),
)

var suggestRoleMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", roleMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: generateExample(`role names for a group named "${prompt1}" which is interested in ${prompt2}`, "")},
}
