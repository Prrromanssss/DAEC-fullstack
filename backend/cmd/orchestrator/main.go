package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	agentservice "github.com/Prrromanssss/DAEE-fullstack/cmd/agent"
	agentagregatorservice "github.com/Prrromanssss/DAEE-fullstack/cmd/agent_agregator"

	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/http-server/handlers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/handlers/slogpretty"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/logcleaner"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Load config
	cfg := config.MustLoad()

	// Configuration logger
	log := setupLogger(cfg.Env, cfg.LogPath)
	log.Info(
		"start daee",
		slog.String("env", cfg.Env),
		slog.String("version", "2"),
	)
	log.Debug("debug messages are enabled")

	go logcleaner.CleanLog(10*time.Minute, cfg.LogPath, 100)

	// Configuration storage

	dbCfg := storage.NewStorage(cfg.StorageURL)

	// Configuration AgentAgregator
	agentAgr, err := agentagregatorservice.RunAgentAgregator(log, cfg, dbCfg)
	if err != nil {
		log.Error("can't make agent agregator", sl.Err(err))
		return
	}

	// Delete previous agents
	err = dbCfg.DB.DeleteAgents(context.Background())
	if err != nil {
		log.Error("can't delete previous agents", sl.Err(err))
	}

	// Create Agent1
	agentservice.RunAgent(log, cfg, dbCfg)

	// Create Agent2
	agentservice.RunAgent(log, cfg, dbCfg)

	// Create Agent3
	agentservice.RunAgent(log, cfg, dbCfg)

	// Configuration http-server
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

	// TODO: Get rid of agentAgr in endpoint!!!!!!!!

	// Expression endpoints
	v1Router.Post("/expressions", handlers.HandlerCreateExpression(log, dbCfg, agentAgr)) // <<<<-----
	v1Router.Get("/expressions", handlers.HandlerGetExpressions(log, dbCfg))

	// Opeartsion endpoints
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

func setupLogger(env, logPath string) *slog.Logger {
	var log *slog.Logger

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}
	defer logFile.Close()

	switch env {
	case envLocal:
		log = setupPrettySlog(logFile)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	}
	return log
}

func setupPrettySlog(logFile *os.File) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(logFile)

	return slog.New(handler)
}
