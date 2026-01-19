package in

import (
	"svpb-tmpl/pkg/job/core"

	"github.com/pocketbase/pocketbase"
	pbcore "github.com/pocketbase/pocketbase/core"
)

// API handles HTTP requests for job module.
type API struct {
	service core.JobService
}

// NewAPI creates a new API adapter.
func NewAPI(service core.JobService) *API {
	return &API{service: service}
}

// Register registers all HTTP routes on the PocketBase app.
func (a *API) Register(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *pbcore.ServeEvent) error {
		se.Router.POST("/api/jobs/{id}/generate-offer", a.handleGenerateOffer)
		return se.Next()
	})
}

// handleGenerateOffer generates a personalized offer for a job.
func (a *API) handleGenerateOffer(e *pbcore.RequestEvent) error {
	authRecord := e.Auth
	if authRecord == nil {
		return e.ForbiddenError("Only authenticated users can generate offers", nil)
	}

	jobID := e.Request.PathValue("id")

	offer, err := a.service.GenerateOffer(e.Request.Context(), jobID, authRecord.Id)
	if err != nil {
		return e.InternalServerError("Failed to generate offer", err)
	}

	return e.JSON(200, map[string]any{
		"offer": offer,
	})
}
