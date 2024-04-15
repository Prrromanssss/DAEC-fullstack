package storage

import (
	"database/sql"
	"log"

	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
	_ "github.com/lib/pq"
)

type Storage struct {
	DB *postgres.Queries
}

func NewStorage(dbURL string) *Storage {
	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Can't connect to database:", err)
	}

	db := postgres.New(conn)

	log.Println("Successfully connected to DB instance")

	return &Storage{
		DB: db,
	}
}
