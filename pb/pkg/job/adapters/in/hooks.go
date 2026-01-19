package in

import (
	"context"

	"svpb-tmpl/pkg/job/core"

	"github.com/pocketbase/pocketbase"
	pbcore "github.com/pocketbase/pocketbase/core"
	"go.uber.org/zap"
)

// Hooks handles PocketBase lifecycle events for job module.
type Hooks struct {
	service core.JobService
	logger  *zap.Logger
}

// NewHooks creates a new Hooks adapter.
func NewHooks(service core.JobService, logger *zap.Logger) *Hooks {
	return &Hooks{
		service: service,
		logger:  logger,
	}
}

// Register registers all PocketBase hooks.
func (h *Hooks) Register(app *pocketbase.PocketBase) {
	// Process raw jobs after creation
	app.OnRecordAfterCreateSuccess("jobs").BindFunc(h.onJobCreated)
}

// onJobCreated triggers LLM processing when a raw job is created.
func (h *Hooks) onJobCreated(e *pbcore.RecordEvent) error {
	record := e.Record

	// Only process raw jobs
	if record.GetString("status") != string(core.StatusRaw) {
		return e.Next()
	}

	// Run processing in background to not block the request
	go func() {
		if err := h.service.Process(context.Background(), record.Id); err != nil {
			h.logger.Error("Job processing failed",
				zap.Error(err),
				zap.String("jobId", record.Id),
			)
		}
	}()

	return e.Next()
}
