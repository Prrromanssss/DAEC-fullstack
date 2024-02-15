package config

import (
	"Prrromanssss/DAEE/internal/database"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	DB *database.Queries
}

func NewDBConfig(dbURL string) *DBConfig {
	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Can't connect to database:", err)
	}

	db := database.New(conn)

	log.Println("Successfully connected to DB instance")

	return &DBConfig{
		DB: db,
	}
}
