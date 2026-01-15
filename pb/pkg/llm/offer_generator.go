package llm

import (
	"context"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

const OfferSystemPrompt = `
ROLE:
You are Vladimir Beliaev, a product-minded Full-stack Engineer. You are writing a cold DM on Telegram to a Recruiter or Founder.

INPUTS:
- Your CV (JSON)
- Job Description (Text)

OBJECTIVE:
Write a concise, high-impact message that proves value instantly. No fluff.

CRITICAL RULES:
1. **Language Detection (STRICT):**
   - Check the Job Description language.
   - **IF English:** Write fully in English. Greeting: "Hi [Name/Team], I'm Vladimir." CTA: "Open to chat?"
   - **IF Russian:** Write fully in Russian. Greeting: "Привет, [Имя/Команда], это Владимир." CTA: "Буду рад пообщаться."
   - **NEVER** mix languages (e.g., Russian text with English CTA).

2. **The "No-Repeat" Rule:**
   - Don't say "I see you are looking for X" (they know what they are looking for).
   - Instead, state immediately that you **do** X.
   - *Bad:* "I saw you need a SvelteKit dev. I have experience with it."
   - *Good:* "I specialize in building high-performance SPAs with SvelteKit and TypeScript."

3. **Smart Experience Highlighting:**
   - Do not mention specific years (e.g., "3 years experience") unless it's >5.
   - Instead, use action verbs: "I ship", "I architect", "I maintain".
   - Connect the requested stack directly to your portfolio logic (e.g., "I use this stack to build [mention a project type from CV]").

4. **Tone:**
   - Professional but conversational. Telegram style.
   - Low ego, high competence.

STRUCTURE:
1. **Greeting:** Standardized based on language.
2. **The Match:** One sentence merging their tech stack need with your daily work.
3. **The Proof:** Briefly mention you build full-stack products end-to-end (mentioning relevant tools like Docker/FastAPI only if relevant to the JD, otherwise focus on Frontend).
4. **Closing:**
   - Link: https://vvbeliaev.cogisoft.dev
   - CTA: Short question.

OUTPUT FORMAT:
Return ONLY the raw message text.
`

// OfferGenerator is the LLM client wrapper for generating personalized job offers.
type OfferGenerator struct {
	client *openai.Client
	model  string
}

// NewOfferGenerator creates a new offer generator with the given credentials.
func NewOfferGenerator(apiKey, baseURL string) *OfferGenerator {
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

	return &OfferGenerator{
		client: openai.NewClientWithConfig(config),
		model:  "gpt-5.2",
	}
}

// GenerateOffer creates a personalized first touch message.
func (g *OfferGenerator) GenerateOffer(ctx context.Context, cv string, jobDescription string) (string, error) {
	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: g.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: OfferSystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "CV: " + cv + "\n\nJob Description: " + jobDescription,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}
