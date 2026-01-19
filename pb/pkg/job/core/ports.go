package core

import "context"

// JobStatus represents the processing state of a job.
type JobStatus string

const (
	StatusRaw        JobStatus = "raw"
	StatusProcessing JobStatus = "processing"
	StatusProcessed  JobStatus = "processed"
	StatusRejected   JobStatus = "rejected"
)

// RawJobInput contains data needed to create a new raw job.
type RawJobInput struct {
	OriginalText string
	ChannelID    int64
	MessageID    int
	Hash         string
	RawData      any
}

// ParsedData represents structured output from LLM extraction.
type ParsedData struct {
	IsVacancy   bool     `json:"isVacancy"`
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	SalaryMin   int      `json:"salaryMin"`
	SalaryMax   int      `json:"salaryMax"`
	Currency    string   `json:"currency"`
	Skills      []string `json:"skills"`
	IsRemote    bool     `json:"isRemote"`
	Grade       string   `json:"grade"`
	Location    string   `json:"location"`
	Description string   `json:"description"`
}

// --- Service Interface (driving port) ---

// JobService is the main interface for job module operations.
// Used by collector module and adapters/in.
type JobService interface {
	// SubmitRaw creates a new job in raw state. Returns job ID.
	SubmitRaw(ctx context.Context, input RawJobInput) (string, error)

	// Process runs LLM extraction on a raw job.
	Process(ctx context.Context, jobID string) error

	// GenerateOffer creates a personalized offer message for a job.
	GenerateOffer(ctx context.Context, jobID, userID string) (string, error)

	// CheckDuplicate returns true if a job with same channelID/messageID or hash exists.
	CheckDuplicate(ctx context.Context, channelID int64, messageID int, hash string) (bool, error)
}

// --- Driven Ports (implemented in adapters/out) ---

// JobExtractor extracts structured data from job posting text.
type JobExtractor interface {
	Extract(ctx context.Context, text string) (ParsedData, error)
}

// OfferGenerator generates personalized offer messages.
type OfferGenerator interface {
	Generate(ctx context.Context, cv, jobDescription string) (string, error)
}
