package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	orchestratorapp "github.com/Prrromanssss/DAEC-fullstack/internal/app/orchestrator"
	daecv1 "github.com/Prrromanssss/DAEC-fullstack/internal/protos/gen/go/daec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Prrromanssss/DAEC-fullstack/internal/config"
	"github.com/Prrromanssss/DAEC-fullstack/internal/http-server/handlers"
	mwlogger "github.com/Prrromanssss/DAEC-fullstack/internal/http-server/middleware/logger"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/setup"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage"

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
	log := setup.SetupLogger(cfg.Env)
	log.Info(
		"start orchestrator",
		slog.String("env", cfg.Env),
		slog.String("version", "2"),
	)

	// Configuration Storage
	dbCfg := storage.NewStorage(log, cfg.StorageURL)

	// Delete terminated agents
	err := dbCfg.Queries.TerminateOldAgents(ctxWithCancel)
	if err != nil {
		log.Warn("can't delete old agents")
	}

	// Configuration Orchestrator
	application, err := orchestratorapp.New(log, cfg, dbCfg, cancel)
	if err != nil {
		panic(err)
	}

	go application.MustRun(ctxWithCancel)

	// Configuration gRPC Client
	conn, err := grpc.Dial(
		cfg.GRPCServer.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("could not connect to grpc server", sl.Err(err))
		os.Exit(1)
	}
	defer conn.Close()

	log.Info("succesfully connect to gRPC server")

	grpcClient := daecv1.NewAuthClient(conn)

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
	v1Router.Use(middleware.URLFormat)

	// Expression endpoints
	v1Router.Post("/expressions", handlers.HandlerCreateExpression(
		log,
		dbCfg,
		cfg.JWTSecret,
		application.OrchestratorApp,
		application.Producer,
	))
	v1Router.Get("/expressions", handlers.HandlerGetExpressions(log, dbCfg, cfg.JWTSecret))

	// Operation endpoints
	v1Router.Get("/operations", handlers.HandlerGetOperations(log, dbCfg, cfg.JWTSecret))
	v1Router.Patch("/operations", handlers.HandlerUpdateOperation(log, dbCfg, cfg.JWTSecret))

	// Agent endpoints
	v1Router.Get("/agents", handlers.HandlerGetAgents(log, dbCfg))

	// User endpoints
	v1Router.Post("/login", handlers.HandlerLoginUser(log, dbCfg, grpcClient))
	v1Router.Post("/register", handlers.HandlerRegisterNewUser(log, dbCfg, grpcClient))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.HTTPServer.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("server starting", slog.String("host", cfg.HTTPServer.Address))

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Error("failed to start server ", sl.Err(err))
		}
	}()

	log.Info("server stopped")

	// Graceful shotdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping agent", slog.String("signal", sign.String()))

	application.Stop(ctxWithCancel, cfg)
}
