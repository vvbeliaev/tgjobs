package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"svpb-tmpl/pkg/collector/core"
	jobcore "svpb-tmpl/pkg/job/core"

	"go.uber.org/zap"
)

// Service implements core.CollectorService.
type Service struct {
	jobService jobcore.JobService
	filter     *core.KeywordFilter
	logger     *zap.Logger
}

// NewService creates a new CollectorService implementation.
func NewService(jobService jobcore.JobService, logger *zap.Logger) *Service {
	return &Service{
		jobService: jobService,
		filter:     core.NewKeywordFilter(),
		logger:     logger,
	}
}

// Handle processes an incoming message.
func (s *Service) Handle(ctx context.Context, msg core.Message) error {
	if msg.Text == "" {
		return nil
	}

	// Pre-filter by keywords
	if !s.filter.ShouldProcess(msg.Text) {
		s.logger.Debug("Message filtered out by keywords",
			zap.Int64("channelId", msg.ChannelID),
			zap.Int("msgId", msg.MessageID),
		)
		return nil
	}

	s.logger.Info("Processing potential job posting",
		zap.Int64("channelId", msg.ChannelID),
		zap.Int("msgId", msg.MessageID),
		zap.Int("runesCount", utf8.RuneCountInString(msg.Text)),
	)

	// Calculate hash for deduplication
	hash := s.calculateHash(msg.Text)

	// Check for duplicates
	isDuplicate, err := s.jobService.CheckDuplicate(ctx, msg.ChannelID, msg.MessageID, hash)
	if err != nil {
		s.logger.Error("Failed to check duplicate", zap.Error(err))
	}
	if isDuplicate {
		s.logger.Debug("Duplicate message, skipping",
			zap.Int64("channelId", msg.ChannelID),
			zap.Int("msgId", msg.MessageID),
			zap.String("hash", hash),
		)
		return nil
	}

	// Submit raw job
	input := jobcore.RawJobInput{
		OriginalText: msg.Text,
		ChannelID:    msg.ChannelID,
		MessageID:    msg.MessageID,
		Hash:         hash,
		RawData:      msg.RawData,
	}

	jobID, err := s.jobService.SubmitRaw(ctx, input)
	if err != nil {
		s.logger.Error("Failed to submit raw job",
			zap.Error(err),
			zap.Int64("channelId", msg.ChannelID),
			zap.Int("msgId", msg.MessageID),
		)
		return nil
	}

	s.logger.Info("Raw job submitted successfully",
		zap.String("jobId", jobID),
		zap.Int64("channelId", msg.ChannelID),
		zap.Int("msgId", msg.MessageID),
	)

	return nil
}

// calculateHash creates a normalized hash of the message text.
func (s *Service) calculateHash(text string) string {
	normalized := strings.ToLower(strings.Join(strings.Fields(text), ""))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
