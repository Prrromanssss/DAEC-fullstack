package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/agent"
	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/http-server/handlers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/http-server/middleware"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/handlers/slogpretty"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/logcleaner"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator"
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

	dbCfg := storage.NewStorage(cfg.StorageURL)

	agentAgregator, err := agent.NewAgentAgregator(
		cfg.RabbitMQURL,
		dbCfg,
		cfg.QueueForSendToAgents,
		cfg.QueueForConsumeFromAgents,
	)
	if err != nil {
		log.Error("agent agregator error", sl.Err(err))
	}

	go agent.AgregateAgents(agentAgregator)

	// Reload computing expressions
	err = orchestrator.ReloadComputingExpressions(dbCfg, agentAgregator)
	if err != nil {
		log.Error("can't reload computing expressions", sl.Err(err))
	}

	// Delete previous agents
	err = dbCfg.DB.DeleteAgents(context.Background())
	if err != nil {
		log.Error("can't delete previous agents", sl.Err(err))
	}

	// Create Agent1
	agent1, err := agent.NewAgent(
		cfg.RabbitMQURL,
		dbCfg,
		cfg.QueueForSendToAgents,
		cfg.QueueForConsumeFromAgents,
		5,
		200,
	)
	if err != nil {
		log.Error("can't create agent1", sl.Err(err))
	}

	go agent.AgentService(agent1)

	// Create Agent2
	agent2, err := agent.NewAgent(
		cfg.RabbitMQURL,
		dbCfg,
		cfg.QueueForSendToAgents,
		cfg.QueueForConsumeFromAgents,
		5,
		200,
	)
	if err != nil {
		log.Error("can't create agent2", sl.Err(err))
	}

	go agent.AgentService(agent2)

	// Create Agent3
	agent3, err := agent.NewAgent(
		cfg.RabbitMQURL,
		dbCfg,
		cfg.QueueForSendToAgents,
		cfg.QueueForConsumeFromAgents,
		5,
		200,
	)
	if err != nil {
		log.Error("can't create agent2", sl.Err(err))
	}

	go agent.AgentService(agent3)

	// Configuration http server
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

	v1Router.Post("/expressions", middleware.MiddlewareAgentAgregatorAndDBConfig(
		handlers.HandlerCreateExpression,
		dbCfg,
		agentAgregator,
	))
	v1Router.Get("/expressions", middleware.MiddlewareApiConfig(handlers.HandlerGetExpressions, dbCfg))

	v1Router.Get("/operations", middleware.MiddlewareApiConfig(handlers.HandlerGetOperations, dbCfg))
	v1Router.Patch("/operations", middleware.MiddlewareApiConfig(handlers.HandlerUpdateOperation, dbCfg))

	v1Router.Get("/agents", middleware.MiddlewareApiConfig(handlers.HandlerGetAgents, dbCfg))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("server starting", slog.String("host", cfg.Address))
	err = srv.ListenAndServe()
	if err != nil {
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
