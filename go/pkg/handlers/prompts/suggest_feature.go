package prompts

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var suggestFeatureMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s\n%s\n%s",
	getSuggestionPrompt("features of ${prompt1}"),
	generateExample("features of ENGL 1010 writing tutoring", "Feedback|Revisions|Brainstorming|Discussion"),
	generateExample("features of Standard gym membership", "Full Gym Equipment|Limited Training|Half-Day Access"),
	generateExample("features of Pro web hosting service", "Unlimited Sites|Unlimited Storage|1TB Bandwidth|Daily Backups"),
	generateExample("features of professional photography service", "Next-Day Prints|High-quality digital photos|Retouching and editing|Choice of location|Choice of outfit changes"),
)

var suggestFeatureMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `I, DelimitedOptions, will provide 5 options delimited by |.`},
	{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: <some prompt> Result:\" and I will complete the result.", suggestFeatureMessagesExample)},
	{Role: openai.ChatMessageRoleUser, Content: generateExample("features of ${prompt1}", "")},
}
