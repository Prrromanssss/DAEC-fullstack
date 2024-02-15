package main

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/handlers"
	"Prrromanssss/DAEE/pkg/agent"
	"Prrromanssss/DAEE/pkg/logcleaner"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

	godotenv.Load(fmt.Sprintf("%s/.env", filepath.Dir(rootPath)))

	portString := os.Getenv("PORT")

	if portString == "" {
		log.Fatal("PORT is not found in environment")
	}

	// Configuration database

	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		log.Fatal("DB_URL is not found in environment")
	}

	apiCfg := config.NewApiConfig(dbURL)

	go logcleaner.CleanLog(10*time.Minute, logPath, 100)

	// Configuration RabbitMQ

	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	if rabbitMQURL == "" {
		log.Fatal("RABBITMQ_URL is not found in environment")
	}

	amqpCfg, err := config.NewAMQPConfig(rabbitMQURL)

	if err != nil {
		log.Fatalf("Can't config RabbitMQ: %v", err)
	}

	agentAgregator := agent.NewAgentAgregator(
		amqpCfg,
		"Queue for sending expressions to agents",
		"Queue for consuming results and pings from agents")

	go agent.AgregateAgents(agentAgregator)

	if err != nil {
		log.Fatalf("Can't connect to RabbitMQ: %v", err)
	}

	// Configuration http server

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()

	v1Router.Post("/expressions", handlers.MiddlewareAgentAgregatorAndApiConfig(handlers.HandlerCreateExpression, apiCfg, agentAgregator))
	v1Router.Get("/expressions", handlers.MiddlewareApiConfig(handlers.HandlerGetExpressions, apiCfg))
	v1Router.Get("/expressions/{expressionID}", handlers.MiddlewareApiConfig(handlers.HandlerGetExpressionByID, apiCfg))

	v1Router.Get("/operations", handlers.MiddlewareApiConfig(handlers.HandlerGetOperations, apiCfg))
	v1Router.Put("/operations", handlers.MiddlewareApiConfig(handlers.HandlerUpdateOperation, apiCfg))

	v1Router.Get("/agents", handlers.MiddlewareApiConfig(handlers.HandlerGetAgents, apiCfg))
	v1Router.Get("/agents/{agentID}", handlers.MiddlewareApiConfig(handlers.HandlerGetAgentByID, apiCfg))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server starting on port %v", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
