package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	orchestratorapp "github.com/Prrromanssss/DAEE-fullstack/internal/app/orchestrator"

	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/http-server/handlers"
	mwlogger "github.com/Prrromanssss/DAEE-fullstack/internal/http-server/middleware/logger"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/logcleaner"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/setup"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load Config
	cfg := config.MustLoad()

	// Configuration Logger
	log := setup.SetupLogger(cfg.Env, cfg.LogPathOrchestrator)
	log.Info(
		"start orchestrator",
		slog.String("env", cfg.Env),
		slog.String("version", "2"),
	)
	log.Debug("debug messages are enabled")

	go logcleaner.CleanLog(10*time.Minute, cfg.LogPathOrchestrator, 100)

	// Configuration Storage
	dbCfg := storage.NewStorage(cfg.StorageURL)

	// Configuration Orchestrator
	application, err := orchestratorapp.New(log, cfg, dbCfg, cancel)
	if err != nil {
		panic(err)
	}

	go application.MustRun(ctxWithCancel)

	// Configuration HTTP-Server
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()

	v1Router.Use(middleware.RequestID)
	v1Router.Use(mwlogger.New(log))

	// Expression endpoints
	v1Router.Post("/expressions", handlers.HandlerCreateExpression(
		log,
		dbCfg,
		application.OrchestratorApp,
		application.Producer,
	))
	v1Router.Get("/expressions", handlers.HandlerGetExpressions(log, dbCfg))

	// Operation endpoints
	v1Router.Get("/operations", handlers.HandlerGetOperations(log, dbCfg))
	v1Router.Patch("/operations", handlers.HandlerUpdateOperation(log, dbCfg))

	// Agent endpoints
	v1Router.Get("/agents", handlers.HandlerGetAgents(log, dbCfg))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("server starting", slog.String("host", cfg.Address))
	if err = srv.ListenAndServe(); err != nil {
		log.Error("failed to start server ", sl.Err(err))
	}

	log.Info("server stopped")
}
