package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"svpb-tmpl/pkg/llm"

	"github.com/gotd/td/tg"
	"github.com/pocketbase/pocketbase/core"
	"go.uber.org/zap"
)

// Handler processes incoming Telegram messages and saves jobs to PocketBase.
type Handler struct {
	app      core.App
	analyzer *llm.Analyzer
	logger   *zap.Logger
	filter   *KeywordFilter
	notifier func(ctx context.Context, text string) error
	ownerID  string
}

// NewHandler creates a new message handler.
func NewHandler(app core.App, analyzer *llm.Analyzer, logger *zap.Logger) *Handler {
	return &Handler{
		app:      app,
		analyzer: analyzer,
		logger:   logger,
		filter:   NewKeywordFilter(),
	}
}

// SetOwnerID sets the default owner for new job records.
func (h *Handler) SetOwnerID(ownerID string) {
	h.ownerID = ownerID
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
		},
		minLength: 50, // Job postings are usually longer than 50 chars
	}
}

// ShouldProcess returns true if the message passes keyword filtering.
func (f *KeywordFilter) ShouldProcess(text string) bool {
	if len(text) < f.minLength {
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
		zap.Int("textLength", len(text)),
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

	parsed, err := h.analyzer.AnalyzeVacancy(ctx, text)
	if err != nil {
		h.logger.Error("LLM analysis failed",
			zap.Error(err),
			zap.Int64("channelId", channelID),
		)
		return nil
	}

	// if !parsed.IsJob {
	// 	h.logger.Debug("LLM determined not a job posting",
	// 		zap.Int64("channelId", channelID),
	// 		zap.Int("msgId", msg.ID),
	// 	)
	// 	return nil
	// }

	if !parsed.IsVacancy {
		h.logger.Debug("LLM determined not a vacancy",
			zap.Int64("channelId", channelID),
			zap.Int("msgId", msg.ID),
		)
		return nil
	}

	// Fallback for empty title if LLM missed it
	if parsed.Title == "" {
		lines := strings.Split(text, "\n")
		if len(lines) > 0 && len(lines[0]) > 0 {
			parsed.Title = lines[0]
			if len(parsed.Title) > 100 {
				parsed.Title = parsed.Title[:97] + "..."
			}
		} else {
			parsed.Title = "Untitled Vacancy"
		}
		h.logger.Warn("LLM returned empty title, using fallback",
			zap.String("fallback_title", parsed.Title),
			zap.Int64("channelId", channelID),
		)
	}

	if err := h.saveJob(parsed, text, channelID, msg.ID, hash, msg); err != nil {
		h.logger.Error("Failed to save job",
			zap.Error(err),
			zap.String("title", parsed.Title),
		)
		return nil
	}

	h.logger.Info("Job saved successfully",
		zap.String("title", parsed.Title),
		zap.String("company", parsed.Company),
		zap.Int("salaryMin", parsed.SalaryMin),
		zap.Int("salaryMax", parsed.SalaryMax),
	)

	// Send notification if notifier is set
	if h.notifier != nil {
		url := fmt.Sprintf("https://t.me/c/%d/%d", channelID, msg.ID)
		notification := h.formatNotification(parsed, url)
		if err := h.notifier(ctx, notification); err != nil {
			h.logger.Error("Failed to send notification", zap.Error(err))
		}
	}

	return nil
}

func (h *Handler) formatNotification(parsed llm.JobParsedData, url string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🚀 НОВАЯ ВАКАНСИЯ: %s\n", strings.ToUpper(parsed.Title)))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	if parsed.Company != "" {
		sb.WriteString(fmt.Sprintf("🏢 Компания: %s\n", parsed.Company))
	}

	salary := ""
	if parsed.SalaryMin > 0 || parsed.SalaryMax > 0 {
		if parsed.SalaryMin > 0 && parsed.SalaryMax > 0 {
			salary = fmt.Sprintf("%d — %d %s", parsed.SalaryMin, parsed.SalaryMax, parsed.Currency)
		} else if parsed.SalaryMin > 0 {
			salary = fmt.Sprintf("от %d %s", parsed.SalaryMin, parsed.Currency)
		} else {
			salary = fmt.Sprintf("до %d %s", parsed.SalaryMax, parsed.Currency)
		}
	}
	if salary != "" {
		sb.WriteString(fmt.Sprintf("💰 Зарплата: %s\n", salary))
	}

	if parsed.Grade != "" {
		sb.WriteString(fmt.Sprintf("📊 Грейд: %s\n", parsed.Grade))
	}

	remoteStr := "Нет"
	if parsed.IsRemote {
		remoteStr = "Да ✅"
	}
	sb.WriteString(fmt.Sprintf("🌍 Удаленка: %s\n", remoteStr))

	if len(parsed.Skills) > 0 {
		sb.WriteString(fmt.Sprintf("🛠 Стек: %s\n", strings.Join(parsed.Skills, ", ")))
	}

	if parsed.Location != "" {
		sb.WriteString(fmt.Sprintf("📍 Локация: %s\n", parsed.Location))
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Ссылка: %s", url))

	return sb.String()
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

func (h *Handler) saveJob(parsed llm.JobParsedData, originalText string, channelID int64, msgID int, hash string, rawMsg *tg.Message) error {
	collection, err := h.app.FindCollectionByNameOrId("jobs")
	if err != nil {
		return fmt.Errorf("jobs collection not found: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", parsed.Title)
	record.Set("company", parsed.Company)
	record.Set("salaryMin", parsed.SalaryMin)
	record.Set("salaryMax", parsed.SalaryMax)
	record.Set("currency", parsed.Currency)
	record.Set("grade", parsed.Grade)
	record.Set("location", parsed.Location)
	record.Set("isRemote", parsed.IsRemote)
	record.Set("description", parsed.Description)
	record.Set("skills", parsed.Skills)
	record.Set("originalText", originalText)
	record.Set("channelId", fmt.Sprintf("%d", channelID))
	record.Set("messageId", msgID)
	record.Set("hash", hash)
	record.Set("raw", rawMsg)
	if h.ownerID != "" {
		record.Set("owner", h.ownerID)
	}

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
