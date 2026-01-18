package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gotd/td/tg"
	"github.com/pocketbase/pocketbase/core"
	"go.uber.org/zap"
)

// Handler processes incoming Telegram messages and saves jobs to PocketBase.
type Handler struct {
	app      core.App
	logger   *zap.Logger
	filter   *KeywordFilter
	notifier func(ctx context.Context, text string) error
}

// NewHandler creates a new message handler.
func NewHandler(app core.App, logger *zap.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
		filter: NewKeywordFilter(),
	}
}

// SetNotifier sets the notification function (e.g., to send to Saved Messages).
func (h *Handler) SetNotifier(notifier func(ctx context.Context, text string) error) {
	h.notifier = notifier
}

// KeywordFilter performs pre-LLM filtering based on keywords.
type KeywordFilter struct {
	// Keywords that indicate a message might be a job posting
	whitelist []string
	// Keywords that indicate spam/irrelevant content
	blacklist []string
	// Minimum message length to consider
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
		minLength: 100, // Job postings are usually longer than 100 runes
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

func (h *Handler) HandleMessage(ctx context.Context, msg *tg.Message, channelID int64) error {
	text := msg.Message
	if text == "" {
		return nil
	}

	if !h.filter.ShouldProcess(text) {
		h.logger.Debug("Message filtered out by keywords",
			zap.Int64("channelId", channelID),
			zap.Int("msgId", msg.ID),
		)
		return nil
	}

	h.logger.Info("Processing potential job posting",
		zap.Int64("channelId", channelID),
		zap.Int("msgId", msg.ID),
		zap.Int("runesCount", utf8.RuneCountInString(text)),
	)

	hash := h.calculateHash(text)
	isDuplicate, err := h.checkDuplicate(channelID, msg.ID, hash)
	if err != nil {
		h.logger.Error("Failed to check duplicate", zap.Error(err))
	}
	if isDuplicate {
		h.logger.Debug("Duplicate message, skipping",
			zap.Int64("channelId", channelID),
			zap.Int("msgId", msg.ID),
			zap.String("hash", hash),
		)
		return nil
	}

	// Save raw job - LLM analysis will be done in the OnRecordCreate hook
	if err := h.saveRawJob(text, channelID, msg.ID, hash, msg); err != nil {
		h.logger.Error("Failed to save raw job",
			zap.Error(err),
			zap.Int64("channelId", channelID),
			zap.Int("msgId", msg.ID),
		)
		return nil
	}

	h.logger.Info("Raw job saved successfully",
		zap.Int64("channelId", channelID),
		zap.Int("msgId", msg.ID),
	)

	return nil
}

func (h *Handler) checkDuplicate(channelID int64, msgID int, hash string) (bool, error) {
	collection, err := h.app.FindCollectionByNameOrId("jobs")
	if err != nil {
		return false, err
	}

	records, err := h.app.FindRecordsByFilter(
		collection.Id,
		"(channelId = {:channelId} && messageId = {:messageId}) || hash = {:hash}",
		"", // sort
		1,  // limit
		0,  // offset
		map[string]any{
			"channelId": fmt.Sprintf("%d", channelID),
			"messageId": msgID,
			"hash":      hash,
		},
	)
	if err != nil {
		return false, err
	}

	return len(records) > 0, nil
}

func (h *Handler) saveRawJob(originalText string, channelID int64, msgID int, hash string, rawMsg *tg.Message) error {
	collection, err := h.app.FindCollectionByNameOrId("jobs")
	if err != nil {
		return fmt.Errorf("jobs collection not found: %w", err)
	}

	// Extract a preliminary title from first line for display purposes
	title := "Pending Analysis"
	lines := strings.Split(originalText, "\n")
	if len(lines) > 0 && len(lines[0]) > 0 {
		firstLine := lines[0]
		if utf8.RuneCountInString(firstLine) > 100 {
			runes := []rune(firstLine)
			title = string(runes[:97]) + "..."
		} else {
			title = firstLine
		}
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("originalText", originalText)
	record.Set("channelId", fmt.Sprintf("%d", channelID))
	record.Set("messageId", msgID)
	record.Set("hash", hash)
	record.Set("raw", rawMsg)
	record.Set("status", "raw")

	url := fmt.Sprintf("https://t.me/c/%d/%d", channelID, msgID)
	record.Set("url", url)

	return h.app.Save(record)
}

func (h *Handler) calculateHash(text string) string {
	// Simple normalization: remove spaces and lowercase
	normalized := strings.ToLower(strings.Join(strings.Fields(text), ""))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
