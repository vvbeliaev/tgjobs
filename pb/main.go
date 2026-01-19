package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"go.uber.org/zap"

	"svpb-tmpl/config"
	"svpb-tmpl/infra/llm"
	collector_in "svpb-tmpl/pkg/collector/adapters/in"
	collector_usecases "svpb-tmpl/pkg/collector/usecases"
	job_in "svpb-tmpl/pkg/job/adapters/in"
	job_out "svpb-tmpl/pkg/job/adapters/out"
	job_usecases "svpb-tmpl/pkg/job/usecases"

	_ "svpb-tmpl/migrations"
)

func main() {
	godotenv.Load("../.env")

	app := pocketbase.New()
	cfg := config.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// --- Global Infrastructure ---
	llmClient := llm.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL)

	// --- Job Module ---
	// Adapters/out (driven ports implementations)
	jobExtractor := job_out.NewExtractor(llmClient)
	offerGenerator := job_out.NewOfferGenerator(llmClient)

	// Usecase
	jobService := job_usecases.NewService(app, jobExtractor, offerGenerator, logger)

	// Adapters/in (driving ports)
	jobAPI := job_in.NewAPI(jobService)
	jobHooks := job_in.NewHooks(jobService, logger)

	// Register job module
	jobAPI.Register(app)
	jobHooks.Register(app)

	// --- Collector Module ---
	// Usecase (depends on job service interface)
	collectorService := collector_usecases.NewService(jobService, logger)

	// Adapters/in
	tgAdapter := collector_in.NewTG(cfg.Telegram, collectorService, logger)

	// Register collector module
	tgAdapter.RegisterCommand(app)

	// --- Static File Serving ---
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", func(e *core.RequestEvent) error {
			path := e.Request.PathValue("path")
			fsys := os.DirFS("./pb_public")

			// PWA treatment block - no cache for service worker files
			if path == "sw.js" || strings.HasPrefix(path, "workbox-") {
				e.Response.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				e.Response.Header().Set("Pragma", "no-cache")
				e.Response.Header().Set("Expires", "0")
				return e.FileFS(fsys, path)
			}

			// Try to serve static file
			err := e.FileFS(fsys, path)
			if err == nil {
				return nil
			}

			// SPA fallback (ignore API and admin routes)
			if !strings.HasPrefix(path, "api/") && !strings.HasPrefix(path, "_/") && !strings.Contains(path, ".") {
				return e.FileFS(fsys, "index.html")
			}

			return nil
		})

		return se.Next()
	})

	// --- Auto-migrations ---
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
		Dir:         "migrations",
	})

	// --- Start Telegram Listener ---
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if !tgAdapter.IsConfigured() {
			log.Println("Telegram not configured (TG_API_ID/TG_API_HASH missing), skipping collector")
			return se.Next()
		}

		if !tgAdapter.SessionExists() {
			log.Printf("Session file not found - run 'tg-login' first")
			return se.Next()
		}

		// Start collector in background
		go startTGCollector(tgAdapter, logger)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// startCollector runs the Telegram message listener in the background.
func startTGCollector(adapter *collector_in.TGAdapter, logger *zap.Logger) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Info("Starting Telegram collector...")
	if err := adapter.Start(ctx); err != nil {
		if err != context.Canceled {
			logger.Error("Telegram collector error", zap.Error(err))
		}
	}

	logger.Info("Telegram collector stopped")
}
