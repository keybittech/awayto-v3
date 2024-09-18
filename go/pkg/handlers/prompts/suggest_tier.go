package prompts

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var suggestTierMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s\n%s\n%s\n%s\n%s",
	getSuggestionPrompt("service level names for ${prompt1}"),
	generateExample("service level names for a generic service", "Small|Medium|Large"),
	generateExample("service level names for writing tutoring at a school writing center", "WRI 1010|WRI 1020|WRI 2010|WRI 2020|WRI 3010"),
	generateExample("service level names for streaming at a web media platform", "Basic|Standard|Premium"),
	generateExample("service level names for advising at a school learning center", "ENG 1010|WRI 1010|MAT 1010|SCI 1010|HIS 1010"),
	generateExample("service level names for travelling on an airline service", "Economy|Business|First Class"),
	generateExample("service level names for reading tutoring at a school reading center", "ESL 900|ESL 990|ENG 1010|ENG 1020|ENG 2010"),
)

var suggestTierMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", suggestTierMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: generateExample("service level names for ${prompt1}", "")},
}
