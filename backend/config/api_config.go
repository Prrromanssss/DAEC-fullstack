package config

import (
	"Prrromanssss/DAEE/internal/database"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type ApiConfig struct {
	DB *database.Queries
}

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
