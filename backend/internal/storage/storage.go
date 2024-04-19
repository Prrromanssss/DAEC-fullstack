package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"

	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
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
func NewStorage(log *slog.Logger, dbURL string) *Storage {
	const fn = "storage.NewStorage"

	log = log.With(
		slog.String("fn", fn),
	)

	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Error("can't connect to database:", sl.Err(err))
		os.Exit(1)
	}

	db := postgres.New(conn)

	log.Info("successfully connected to DB instance")

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
