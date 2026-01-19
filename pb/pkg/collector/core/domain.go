package core

import (
	"strings"
	"unicode/utf8"
)

// Message represents an incoming message from a source.
type Message struct {
	Text      string
	ChannelID int64
	MessageID int
	RawData   any
}

// KeywordFilter performs pre-LLM filtering based on keywords.
type KeywordFilter struct {
	whitelist []string
	blacklist []string
	minLength int
}

// NewKeywordFilter creates a filter with default job-related keywords.
func NewKeywordFilter() *KeywordFilter {
	return &KeywordFilter{
		whitelist: []string{
			// Russian
			"вакансия", "ищем", "требуется", "работа", "зарплата",
			"оклад", "удаленка", "удалённо", "офис", "опыт работы",
			"junior", "middle", "senior", "lead", "тимлид",
			"разработчик", "developer", "программист", "инженер",
			// English
			"vacancy", "hiring", "job", "position", "salary",
			"remote", "on-site", "experience", "looking for",
			"we are hiring", "join our team", "opportunity",
			"engineer", "developer", "programmer",
			// Tech keywords
			"golang", "python", "javascript", "typescript", "react",
			"backend", "frontend", "fullstack", "devops", "sre",
			"kubernetes", "docker", "aws", "gcp", "azure",
		},
		blacklist: []string{
			"реклама", "продам", "куплю", "скидка", "акция",
			"casino", "казино", "betting", "ставки", "crypto pump",
			"#резюме",
		},
		minLength: 100,
	}
}

// ShouldProcess returns true if the message passes keyword filtering.
func (f *KeywordFilter) ShouldProcess(text string) bool {
	if utf8.RuneCountInString(text) < f.minLength {
		return false
	}

	lower := strings.ToLower(text)
	for _, kw := range f.blacklist {
		if strings.Contains(lower, kw) {
			return false
		}
	}

	for _, kw := range f.whitelist {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	return true
}
