package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	grpcapp "github.com/Prrromanssss/DAEC-fullstack/internal/app/grpc"
	"github.com/Prrromanssss/DAEC-fullstack/internal/config"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/setup"
	"github.com/Prrromanssss/DAEC-fullstack/internal/services/auth"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage"
)

func main() {
	// Load Config
	cfg := config.MustLoad()

	// Configuration Logger
	log := setup.SetupLogger(cfg.Env, cfg.LogPathAuth)
	log.Info(
		"start grpc server",
		slog.String("env", cfg.Env),
		slog.String("version", "2"),
	)
	log.Debug("debug messages are enabled")

	// Configuration Storage
	dbCfg := storage.NewStorage(log, cfg.StorageURL)

	authService := auth.New(log, dbCfg, dbCfg, cfg.TokenTTL)

	grpcApp := grpcapp.New(log, authService, cfg.GRPCServer.Address)

	go grpcApp.MustRun()

	// Graceful shotdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	grpcApp.Stop()

	log.Info("grpc server stopped")
}
