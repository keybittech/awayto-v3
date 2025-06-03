package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const (
	LLM_MODEL  = "gemini-2.0-flash"
	LLM_URL    = "https://generativelanguage.googleapis.com/v1beta/models/" + LLM_MODEL + ":generateContent?key="
	CHAT_USER  = "user"
	CHAT_MODEL = "model"
)

var stringArrayGeneration = ResponseSchema{
	ResponseMimeType: "application/json",
	ResponseSchema: Schema{
		Type:    "STRING",
		Pattern: `^[^|]*\|[^|]*\|[^|]*\|[^|]*\|[^|]*$`,
	},
}

type ResponseSchema struct {
	ResponseMimeType string `json:"responseMimeType"`
	ResponseSchema   Schema `json:"responseSchema"`
}

type Schema struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

type Part struct {
	Text string `json:"text"`
}

type ChatContent struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

type ChatRequest struct {
	SystemInstruction ChatContent    `json:"system_instruction"`
	Contents          []ChatContent  `json:"contents"`
	GenerationConfig  ResponseSchema `json:"generationConfig"`
}

type Candidate struct {
	Content ChatContent `json:"content"`
}

type CandidateResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type LLMPrompts map[types.IPrompts]ChatRequest

type LLM struct {
	Prompts       LLMPrompts
	APIKey, Model string
}

func InitLLM() *LLM {

	aiPrompts := make(LLMPrompts)
	aiPrompts[types.IPrompts_CONVERT_PURPOSE] = convertPurposeRequest
	aiPrompts[types.IPrompts_SUGGEST_FEATURE] = suggestFeaturesRequest
	aiPrompts[types.IPrompts_SUGGEST_ROLE] = suggestRolesRequest
	aiPrompts[types.IPrompts_SUGGEST_SERVICE] = suggestServicesRequest
	aiPrompts[types.IPrompts_SUGGEST_TIER] = suggestTiersRequest

	apiKey, err := util.GetEnvFilePath("AI_KEY_FILE", 40)
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
		log.Fatal(util.ErrCheck(err))
	}

	aic := &LLM{
		APIKey:  apiKey,
		Model:   LLM_MODEL,
		Prompts: aiPrompts,
	}

	util.DebugLog.Println("Ai Init")

	return aic
}

func (llm *LLM) GetCandidateResponse(ctx context.Context, promptParts []string, promptType types.IPrompts) (*CandidateResponse, error) {

	promptTokens := make(map[string]string)

	promptRequest := llm.Prompts[promptType]

	for i, prompt := range promptParts {
		tokenKey := fmt.Sprintf("${prompt%d}", i+1)
		promptTokens[tokenKey] = prompt
	}

	for _, message := range promptRequest.Contents {
		for k, v := range promptTokens {
			message.Parts[0].Text = strings.ReplaceAll(message.Parts[0].Text, k, v)
		}
	}

	jsonBytes, err := json.Marshal(promptRequest)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	resp, err := util.PostFormData(ctx, LLM_URL+llm.APIKey, http.Header{"Content-Type": {"application/json"}}, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	candidateResponse := &CandidateResponse{}
	err = json.Unmarshal(resp, &candidateResponse)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return candidateResponse, nil
}

func GetSuggestionPrompt(prompt string) string {
	return `Generate 5 ` + prompt + `; Result is 1-3 words separated by |. Here are some examples: `
}

func GenerateExample(prompt string, result string) string {
	return `Phrase: ` + prompt + `\nResult: ` + result
}

func HasSimilarKey(obj map[string]any, regex regexp.Regexp) bool {
	has := false
	for key := range obj {
		if regex.Match([]byte(key)) {
			has = true
			break
		}
	}
	return has
}

var gerundInstructions = ChatContent{
	Parts: []Part{
		{Text: `You respond with a single gerund phrase, 5 to 8 words, to be appended to the requested incomplete sentence. For example, if the request is "I remember when I went", the response could be "hiking in the alps".`},
	},
}

var optionsInstructions = ChatContent{
	Parts: []Part{
		{Text: `Respond with a single set of options which is a pipe delimited string separating 5 substrings.`},
	},
}

var convertPurposeRequest = ChatRequest{
	SystemInstruction: gerundInstructions,
	Contents: []ChatContent{
		{
			Role: CHAT_MODEL,
			Parts: []Part{
				{Text: `Give me an incomplete sentence that I can complete with a gerund phrase. For example, if you said "Our favorite past time is", I might respond "walking around downtown".`},
			},
		},
		{
			Role: CHAT_USER,
			Parts: []Part{
				{Text: `An organization named "${prompt1}" is interested in "${prompt2}" and their mission statement is`},
			},
		},
	},
}

var suggestFeaturesRequest = ChatRequest{
	SystemInstruction: optionsInstructions,
	GenerationConfig:  stringArrayGeneration,
	Contents: []ChatContent{
		{
			Role: CHAT_MODEL,
			Parts: []Part{
				{Text: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: PROMPT_TEXT Result:\" and I will complete the result.", suggestFeatureMessagesExample)},
			},
		},
		{
			Role: CHAT_USER,
			Parts: []Part{
				{Text: GenerateExample("features of ${prompt1}", "")},
			},
		},
	},
}

var suggestFeatureMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s\n%s\n%s",
	GetSuggestionPrompt("features of ${prompt1}"),
	GenerateExample("features of ENGL 1010 writing tutoring", "Feedback|Revisions|Brainstorming|Discussion"),
	GenerateExample("features of Standard gym membership", "Full Gym Equipment|Limited Training|Half-Day Access"),
	GenerateExample("features of Pro web hosting service", "Unlimited Sites|Unlimited Storage|1TB Bandwidth|Daily Backups"),
	GenerateExample("features of professional photography service", "Next-Day Prints|High-quality digital photos|Retouching and editing|Choice of location|Choice of outfit changes"),
)

var suggestRolesRequest = ChatRequest{
	SystemInstruction: optionsInstructions,
	GenerationConfig:  stringArrayGeneration,
	Contents: []ChatContent{
		{
			Role: CHAT_MODEL,
			Parts: []Part{
				{Text: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: PROMPT_TEXT Result:\" and I will complete the result.", roleMessagesExample)},
			},
		},
		{
			Role: CHAT_USER,
			Parts: []Part{
				{Text: GenerateExample(`role names for a group named "${prompt1}" which is interested in ${prompt2}`, "")},
			},
		},
	},
}

var roleMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s",
	GetSuggestionPrompt("role names for a group named ${prompt1} which is interested in ${prompt2}"),
	GenerateExample("role names for a group named writing center which is interested in consulting on writing", "Tutor|Student|Advisor|Administrator|Consultant"),
	GenerateExample("role names for a group named city maintenance department which is interested in maintaining the facilities in the city", "Dispatcher|Engineer|Administrator|Technician|Manager"),
)

var suggestServicesRequest = ChatRequest{
	SystemInstruction: optionsInstructions,
	GenerationConfig:  stringArrayGeneration,
	Contents: []ChatContent{
		{
			Role: CHAT_MODEL,
			Parts: []Part{
				{Text: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: PROMPT_TEXT Result:\" and I will complete the result.", suggestServiceMessagesExample)},
			},
		},
		{
			Role: CHAT_USER,
			Parts: []Part{
				{Text: GenerateExample("gerund verbs performed for the purpose of ${prompt1}", "")},
			},
		},
	},
}

var suggestServiceMessagesExample = fmt.Sprintf(
	"%s\n%s\n%s",
	GetSuggestionPrompt("gerund verbs performed for the purpose of ${prompt1}"),
	GenerateExample("gerund verbs performed for the purpose of offering educational services to community college students", "Tutoring|Advising|Consulting|Instructing|Mentoring"),
	GenerateExample("gerund verbs performed for the purpose of providing banking services to the local area", "Accounting|Financing|Forecasting|Financial Planning|Investing"),
)

var suggestTiersRequest = ChatRequest{
	SystemInstruction: optionsInstructions,
	GenerationConfig:  stringArrayGeneration,
	Contents: []ChatContent{
		{
			Role: CHAT_MODEL,
			Parts: []Part{
				{Text: fmt.Sprintf("Simply provide your desired prompt, and I'll fill in the result!\nHere are some examples:\n%s\nProvide the following text \"Prompt: PROMPT_TEXT Result:\" and I will complete the result.", suggestTierMessagesExample)},
			},
		},
		{
			Role: CHAT_USER,
			Parts: []Part{
				{Text: GenerateExample("service level names for ${prompt1}", "")},
			},
		},
	},
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
