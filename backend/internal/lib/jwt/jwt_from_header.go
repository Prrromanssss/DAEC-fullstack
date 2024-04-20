package jwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

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
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, err
	}
	userID, ok := claims["uid"].(int32)
	if !ok {
		return 0, err
	}

	return userID, nil
}
