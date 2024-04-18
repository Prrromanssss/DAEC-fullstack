package storage

import (
	"database/sql"
	"errors"
	"log"

	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
	_ "github.com/lib/pq"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

type Storage struct {
	DB *postgres.Queries
}

// NewStorage creates new Storage.
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
