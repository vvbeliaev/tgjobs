package out

import (
	"context"
	"encoding/json"

	"svpb-tmpl/pkg/job/core"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const extractionPrompt = `You are a job vacancy parser. Your task is to analyze text messages and extract structured data about job postings.

IMPORTANT RULES:
1. If the text is NOT a job vacancy (e.g., advertisement, news, chat message), set isVacancy to false and leave other fields empty/default.
2. If it IS a vacancy, set isVacancy to true and EXTRACT the job title (e.g., "Golang Developer", "Product Manager").
3. If the title is not explicitly stated, infer it from the context or use the most prominent role mentioned. NEVER leave title empty if isVacancy is true.
4. Extract salary information if present. Convert to numbers only, no currency symbols.
5. Identify the currency from context (look for $, €, ₽, USD, EUR, RUB, etc.)
6. Extract required skills/technologies as a list of short keywords.
7. Determine job grade from context clues (Junior/Middle/Senior/Lead/Principal).
8. Set isRemote to true if remote work, WFH, or distributed team is mentioned.

Always respond with valid JSON matching the schema exactly.`

type Extractor struct {
	client *openai.Client
	model  string
}

func NewExtractor(client *openai.Client) *Extractor {
	return &Extractor{
		client: client,
		model:  "gpt-5-nano",
	}
}

func (e *Extractor) Extract(ctx context.Context, text string) (core.ParsedData, error) {
	resp, err := e.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: e.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: extractionPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "job_parser",
					Schema: parsedDataSchema(),
					Strict: true,
				},
			},
		},
	)

	if err != nil {
		return core.ParsedData{}, err
	}

	var result core.ParsedData
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	return result, err
}

// parsedDataSchema returns JSON schema for ParsedData.
func parsedDataSchema() *jsonschema.Definition {
	schema, err := jsonschema.GenerateSchemaForType(core.ParsedData{})
	if err != nil {
		panic(err)
	}
	return schema
}
