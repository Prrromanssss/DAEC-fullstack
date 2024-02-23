package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/config"
	"github.com/Prrromanssss/DAEE-fullstack/handlers"
	"github.com/Prrromanssss/DAEE-fullstack/pkg/agent"
	"github.com/Prrromanssss/DAEE-fullstack/pkg/logcleaner"
	"github.com/Prrromanssss/DAEE-fullstack/pkg/reload"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal("Can't get pwd")
	}
	rootPath := filepath.Dir(filepath.Dir(path))
	logPath := fmt.Sprintf("%s/daee.log", rootPath)

	// Configuration log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Log file is not found in environment")
	} else {
		log.SetOutput(file)
	}
	defer file.Close()

	go logcleaner.CleanLog(10*time.Minute, logPath, 100)

	// Load env variables
	err = godotenv.Load(fmt.Sprintf("%s/.env", filepath.Dir(rootPath)))
	if err != nil {
		log.Fatalf("Can't load env variables: %v", err)
	}

	// Configuration database
	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		log.Fatal("DB_URL is not found in environment")
	}

	dbCfg := config.NewDBConfig(dbURL)

	// Configuration RabbitMQ
	rabbitMQURL := os.Getenv("RABBIT_MQ_URL")

	if rabbitMQURL == "" {
		log.Fatal("RABBITMQ_URL is not found in environment")
	}

	queueForSendToAgentsString := "Queue for sending expressions to agents"
	queueForConsumeFromAgentsString := "Queue for consuming results and pings from agents"

	agentAgregator, err := agent.NewAgentAgregator(
		rabbitMQURL,
		dbCfg,
		queueForSendToAgentsString,
		queueForConsumeFromAgentsString,
	)
	if err != nil {
		log.Fatalf("Agent Agregator Error: %v", err)
	}

	go agent.AgregateAgents(agentAgregator)

	// Reload computing expressions
	err = reload.ReloadComputingExpressions(dbCfg, agentAgregator)
	if err != nil {
		log.Fatalf("Can't reload computin expressions: %v", err)
	}

	// Create operation
	config.ConfigOperation(dbCfg)

	// Delete previous agents
	err = dbCfg.DB.DeleteAgents(context.Background())
	if err != nil {
		log.Fatalf("Can't delete previous agents: %v", err)
	}

	// Create Agent1
	agent1, err := agent.NewAgent(
		rabbitMQURL,
		dbCfg,
		queueForSendToAgentsString,
		queueForConsumeFromAgentsString,
		5,
		200,
	)
	if err != nil {
		log.Fatalf("Can't create agent1: %v", err)
	}

	go agent.AgentService(agent1)

	// Create Agent2
	agent2, err := agent.NewAgent(
		rabbitMQURL,
		dbCfg,
		queueForSendToAgentsString,
		queueForConsumeFromAgentsString,
		5,
		200,
	)
	if err != nil {
		log.Fatalf("Can't create agent2: %v", err)
	}

	go agent.AgentService(agent2)

	// Create Agent3
	agent3, err := agent.NewAgent(
		rabbitMQURL,
		dbCfg,
		queueForSendToAgentsString,
		queueForConsumeFromAgentsString,
		5,
		200,
	)
	if err != nil {
		log.Fatalf("Can't create agent2: %v", err)
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

	v1Router.Post("/expressions", handlers.MiddlewareAgentAgregatorAndDBConfig(
		handlers.HandlerCreateExpression,
		dbCfg,
		agentAgregator,
	))
	v1Router.Get("/expressions", handlers.MiddlewareApiConfig(handlers.HandlerGetExpressions, dbCfg))

	v1Router.Get("/operations", handlers.MiddlewareApiConfig(handlers.HandlerGetOperations, dbCfg))
	v1Router.Patch("/operations", handlers.MiddlewareApiConfig(handlers.HandlerUpdateOperation, dbCfg))

	v1Router.Get("/agents", handlers.MiddlewareApiConfig(handlers.HandlerGetAgents, dbCfg))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":3000",
	}

	log.Printf("Server starting on port %v", 3000)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
