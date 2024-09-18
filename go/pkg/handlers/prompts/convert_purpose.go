package prompts

import "github.com/sashabaranov/go-openai"

var convertPurposeMessages = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleSystem, Content: `You respond with a single gerund phrase, 5 to 8 words, which will complete the user's sentence. For example, "walking in the sunshine with you".`},
	{Role: openai.ChatMessageRoleAssistant, Content: `Give me an incomplete sentence that I can complete with a gerund phrase. For example, if you said "Our favorite past time is", I might respond "walking in the sunshine with you"`},
	{Role: openai.ChatMessageRoleUser, Content: `An organization named "${prompt1}" is interested in "${prompt2}" and their mission statement is`},
}
