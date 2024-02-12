package config

import (
	"Prrromanssss/DAEE/internal/database"
	"database/sql"
	"log"
	"net/http"
)

type ApiConfig struct {
	DB *database.Queries
}

type configHandler func(http.ResponseWriter, *http.Request, *ApiConfig)

func NewApiConfig(dbURL string) *ApiConfig {
	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Can't connect to database:", err)
	}

	db := database.New(conn)

	return &ApiConfig{
		DB: db,
	}
}

func MiddlewareConfig(handler configHandler, apiCfg *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, apiCfg)
	}
}
