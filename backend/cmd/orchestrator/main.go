package main

import (
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

	// configuration log file
	file, err := os.OpenFile(fmt.Sprintf("%s/log.txt", rootPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

	// conn, err := sql.Open("postgres", dbURL)

	// if err != nil {
	// 	log.Fatal("Can't connect to database:", err)
	// }

	// db := database.New(conn)
	// apiCfg := auth.ApiConfig{
	// 	DB: db,
	// }

	go logcleaner.CleanLog(10*time.Minute, fmt.Sprintf("%s/log.txt", rootPath), 100)

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
