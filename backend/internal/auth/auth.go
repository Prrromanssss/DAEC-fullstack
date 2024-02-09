package auth

import (
	"Prrromanssss/DAEE/internal/database"
	"errors"
	"net/http"
	"strings"
)

type ApiConfig struct {
	DB *database.Queries
}

// GetAPIKey extracts an API Key from
// the headers of an HTTP request
// Example:
// Authorization: ApiKey {insert apikey here}
func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no authentication info found")
	}

	vals := strings.Split(val, " ")
	if len(vals) != 2 {
		return "", errors.New("malformed auth header")
	}
	if vals[0] != "ApiKey" {
		return "", errors.New("malformed first part of auth header")
	}
	return vals[1], nil
}
