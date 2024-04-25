package jwt

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Prrromanssss/DAEC-fullstack/internal/storage/postgres"
	"github.com/golang-jwt/jwt/v5"
)

// NewToken creates new JWT token.
func NewToken(user postgres.User, duration time.Duration) (string, error) {
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

func getTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	// Checks that header starts with "Bearer".
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	return parts[1], nil // returns token without "Bearer".
}

func GetUidFromJWT(r *http.Request, secret string) (int32, error) {
	jwtToken, err := getTokenFromHeader(r)
	if err != nil {
		return 0, err
	}

	// Parse JWT Token.
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("error in map claims")
	}

	userIDFloat, ok := claims["uid"].(float64)
	if !ok {
		return 0, errors.New("jwt token does not contain uid")
	}

	userID := int32(userIDFloat)

	if userID == 0 {
		return 0, errors.New("userID == 0")
	}

	return userID, nil
}
