package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pocketbase/pocketbase"
	pbcore "github.com/pocketbase/pocketbase/core"
	"go.uber.org/zap"

	"svpb-tmpl/pkg/job/core"
)

// Service implements core.JobService.
type Service struct {
	app       *pocketbase.PocketBase
	extractor core.JobExtractor
	offerGen  core.OfferGenerator
	logger    *zap.Logger
}

// NewService creates a new JobService implementation.
func NewService(
	app *pocketbase.PocketBase,
	extractor core.JobExtractor,
	offerGen core.OfferGenerator,
	logger *zap.Logger,
) *Service {
	return &Service{
		app:       app,
		extractor: extractor,
		offerGen:  offerGen,
		logger:    logger,
	}
}

// SubmitRaw creates a new job in raw state.
func (s *Service) SubmitRaw(ctx context.Context, input core.RawJobInput) (string, error) {
	collection, err := s.app.FindCollectionByNameOrId("jobs")
	if err != nil {
		return "", fmt.Errorf("jobs collection not found: %w", err)
	}

	job := core.NewRawJob(collection, input)

	if err := s.app.Save(job.Record()); err != nil {
		return "", fmt.Errorf("failed to save raw job: %w", err)
	}

	s.logger.Info("Raw job submitted",
		zap.String("jobId", job.ID()),
		zap.Int64("channelId", input.ChannelID),
		zap.Int("messageId", input.MessageID),
	)

	return job.ID(), nil
}

// Process runs LLM extraction on a raw job.
func (s *Service) Process(ctx context.Context, jobID string) error {
	record, err := s.app.FindRecordById("jobs", jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	job := core.NewJob(record)

	if err := job.MarkProcessing(); err != nil {
		return fmt.Errorf("cannot process job: %w", err)
	}

	// Save processing state
	if err := s.app.Save(job.Record()); err != nil {
		s.logger.Error("Failed to save processing state", zap.Error(err))
	}

	// Extract data using LLM
	parsed, err := s.extractor.Extract(ctx, job.OriginalText())
	if err != nil {
		s.logger.Error("LLM extraction failed",
			zap.Error(err),
			zap.String("jobId", jobID),
		)
		if rejectErr := job.Reject("LLM extraction failed"); rejectErr == nil {
			s.app.Save(job.Record())
		}
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Check if it's actually a vacancy
	if !parsed.IsVacancy {
		s.logger.Info("LLM determined not a vacancy",
			zap.String("jobId", jobID),
		)
		if err := job.Reject("not a vacancy"); err == nil {
			s.app.Save(job.Record())
		}
		return nil
	}

	// Complete with parsed data
	if err := job.Complete(parsed); err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	if err := s.app.Save(job.Record()); err != nil {
		return fmt.Errorf("failed to save processed job: %w", err)
	}

	s.logger.Info("Job processed successfully",
		zap.String("jobId", jobID),
		zap.String("title", parsed.Title),
	)

	return nil
}

// GenerateOffer creates a personalized offer message for a job.
func (s *Service) GenerateOffer(ctx context.Context, jobID, userID string) (string, error) {
	// Get job
	job, err := s.app.FindRecordById("jobs", jobID)
	if err != nil {
		return "", fmt.Errorf("job not found: %w", err)
	}

	// Get user with CV
	user, err := s.app.FindRecordById("users", userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	cv := user.Get("cv")
	cvBytes, _ := json.Marshal(cv)

	jobDescription := job.GetString("description") + "\n" + job.GetString("originalText")

	// Generate offer
	offer, err := s.offerGen.Generate(ctx, string(cvBytes), jobDescription)
	if err != nil {
		return "", fmt.Errorf("failed to generate offer: %w", err)
	}

	// Save to userJobMap
	if err := s.saveUserJobOffer(userID, jobID, offer); err != nil {
		s.logger.Error("Failed to save offer to userJobMap", zap.Error(err))
	}

	return offer, nil
}

// CheckDuplicate returns true if a job with same channelID/messageID or hash exists.
func (s *Service) CheckDuplicate(ctx context.Context, channelID int64, messageID int, hash string) (bool, error) {
	collection, err := s.app.FindCollectionByNameOrId("jobs")
	if err != nil {
		return false, err
	}

	records, err := s.app.FindRecordsByFilter(
		collection.Id,
		"(channelId = {:channelId} && messageId = {:messageId}) || hash = {:hash}",
		"",
		1,
		0,
		map[string]any{
			"channelId": fmt.Sprintf("%d", channelID),
			"messageId": messageID,
			"hash":      hash,
		},
	)
	if err != nil {
		return false, err
	}

	return len(records) > 0, nil
}

// saveUserJobOffer saves the offer to userJobMap collection.
func (s *Service) saveUserJobOffer(userID, jobID, offer string) error {
	collection, err := s.app.FindCollectionByNameOrId("userJobMap")
	if err != nil {
		return err
	}

	// Find existing or create new
	userJob, err := s.app.FindFirstRecordByFilter("userJobMap", "user = {:userId} && job = {:jobId}", map[string]any{
		"userId": userID,
		"jobId":  jobID,
	})

	if err != nil {
		userJob = pbcore.NewRecord(collection)
		userJob.Set("user", userID)
		userJob.Set("job", jobID)
	}

	userJob.Set("offer", offer)

	return s.app.Save(userJob)
}
