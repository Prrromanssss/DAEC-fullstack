package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	agentapp "github.com/Prrromanssss/DAEC-fullstack/internal/app/agent"
	"github.com/Prrromanssss/DAEC-fullstack/internal/config"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/logcleaner"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/setup"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage"
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
	dbCfg := storage.NewStorage(log, cfg.StorageURL)

	// Configuration Agent
	application, err := agentapp.New(log, cfg, dbCfg, cancel)
	if err != nil {
		panic(err)
	}

	go application.MustRun(ctxWithCancel)

	// Graceful shotdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping agent", slog.String("signal", sign.String()))

	application.Stop(ctxWithCancel)

	log.Info("agent stopped")
}
