package llm

import (
	openai "github.com/sashabaranov/go-openai"
)

// NewClient creates a new OpenAI client with the given credentials.
// This is a global infrastructure factory - domain-specific logic
// should be implemented in adapters/out of respective modules.
func NewClient(apiKey, baseURL string) *openai.Client {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	return openai.NewClientWithConfig(config)
}
