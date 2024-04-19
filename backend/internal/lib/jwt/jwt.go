package jwt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// NewToken creates new JWT token.
func NewToken(user postgres.User, duration time.Duration) (string, error) {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("can't get pwd: %v", err)
	}

	err = godotenv.Load(fmt.Sprintf("%s/local.env", filepath.Dir(filepath.Dir(filepath.Dir(path)))))
	if err != nil {
		log.Fatalf("can't parse env file: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":   user.UserID,
		"email": user.Email,
		"exp":   time.Now().Add(duration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
