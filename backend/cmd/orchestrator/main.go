package main

import (
	"Prrromanssss/DAEE/internal/config"
	"Prrromanssss/DAEE/internal/handlers"
	"Prrromanssss/DAEE/internal/logcleaner"
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

	// configuration log file
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

	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		log.Fatal("DB_URL is not found in environment")
	}

	apiCfg := config.NewApiConfig(dbURL)

	go logcleaner.CleanLog(10*time.Minute, logPath, 100)

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

	v1Router.Post("/expressions", config.MiddlewareConfig(handlers.HandlerCreateExpression, apiCfg))
	v1Router.Get("/expressions", config.MiddlewareConfig(handlers.HandlerGetExpressions, apiCfg))
	v1Router.Get("/expressions/{expressionID}", config.MiddlewareConfig(handlers.HandlerGetExpressionByID, apiCfg))

	v1Router.Get("/operations", config.MiddlewareConfig(handlers.HandlerGetOperations, apiCfg))
	v1Router.Put("/operations", config.MiddlewareConfig(handlers.HandlerUpdateOperation, apiCfg))

	v1Router.Get("/computers", config.MiddlewareConfig(handlers.HandlerGetComputers, apiCfg))

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
