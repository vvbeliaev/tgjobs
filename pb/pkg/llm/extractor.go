package llm

import (
	"context"
	"encoding/json"
	"os"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// JobParsedData represents the structured output from LLM analysis.
type JobParsedData struct {
	IsVacancy   bool     `json:"isVacancy"`   // False if spam/advertisement/not a job posting
	Title       string   `json:"title"`       // Job title
	Company     string   `json:"company"`     // Company name if mentioned
	SalaryMin   int      `json:"salaryMin"`   // Minimum salary (0 if not specified)
	SalaryMax   int      `json:"salaryMax"`   // Maximum salary (0 if not specified)
	Currency    string   `json:"currency"`    // Currency code (USD, EUR, RUB, etc.)
	Skills      []string `json:"skills"`      // Required skills/technologies
	IsRemote    bool     `json:"isRemote"`    // Whether remote work is available
	Grade       string   `json:"grade"`       // Junior, Middle, Senior, Lead, etc.
	Location    string   `json:"location"`    // Office location if mentioned
	Description string   `json:"description"` // Brief job description
}

func (j JobParsedData) Schema() *jsonschema.Definition {
	schema, err := jsonschema.GenerateSchemaForType(j)
	if err != nil {
		panic(err)
	}
	return schema
}

// Extractor is the LLM client wrapper for job vacancy analysis (extraction).
type Extractor struct {
	client *openai.Client
	model  string
}

// NewExtractor creates a new LLM extractor with the given credentials.
func NewExtractor(apiKey, baseURL string) *Extractor {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if baseURL == "" {
		baseURL = os.Getenv("OPENAI_BASE_URL")
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	return &Extractor{
		client: openai.NewClientWithConfig(config),
		model:  "gpt-5-nano",
	}
}

// ExtractionPrompt defines the LLM's behavior for parsing job vacancies.
const ExtractionPrompt = `You are a job vacancy parser. Your task is to analyze text messages and extract structured data about job postings.

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

// AnalyzeVacancy sends the message text to LLM and returns structured job data.
func (e *Extractor) AnalyzeVacancy(ctx context.Context, text string) (JobParsedData, error) {
	resp, err := e.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: e.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: ExtractionPrompt,
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
					Schema: JobParsedData{}.Schema(),
					Strict: true,
				},
			},
		},
	)

	if err != nil {
		return JobParsedData{}, err
	}

	var result JobParsedData
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	return result, err
}
