package main

import (
	"context"
	"log/slog"
	"time"

	agentapp "github.com/Prrromanssss/DAEE-fullstack/internal/app/agent"
	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/logcleaner"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/setup"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
)

func main() {
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load Config
	cfg := config.MustLoad()

	// Configuration Logger
	log := setup.SetupLogger(cfg.Env, cfg.LogPathAgent)
	log.Info(
		"start agent",
		slog.String("env", cfg.Env),
		slog.String("version", "2"),
	)
	log.Debug("debug messages are enabled")

	go logcleaner.CleanLog(10*time.Minute, cfg.LogPathAgent, 100)

	// Configuration Storage
	dbCfg := storage.NewStorage(cfg.StorageURL)

	// Configuration Agent
	application, err := agentapp.New(log, cfg, dbCfg, cancel)
	if err != nil {
		panic(err)
	}

	application.MustRun(ctxWithCancel)
}
