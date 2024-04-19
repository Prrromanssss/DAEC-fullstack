package storage

import (
	"context"
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
	Queries *postgres.Queries
	DB      *sql.DB
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
		Queries: db,
		DB:      conn,
	}
}

// SaveUser saves user to storage.
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int32, error) {
	userID, err := s.Queries.SaveUser(ctx, postgres.SaveUserParams{
		Email:        email,
		PasswordHash: passHash,
	})
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// User gets user from storage.
func (s *Storage) User(ctx context.Context, email string) (postgres.User, error) {
	user, err := s.Queries.GetUser(ctx, email)
	if err != nil {
		return postgres.User{}, err
	}

	return user, nil
}
