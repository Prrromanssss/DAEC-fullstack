package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
)

// The respondWithError function is designed to handle HTTP responses that indicate an error condition.
func respondWithError(log *slog.Logger, w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Warn("responding with 5XX error", slog.Int("code", code))
	}
	type errResponse struct {
		Error string `json:"error"`
	}

	respondWithJson(log, w, code, errResponse{
		Error: msg,
	})
}

// The respondWithJson function is a utility function designed to send HTTP responses with JSON content
func respondWithJson(log *slog.Logger, w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Error("failed to marshal JSON response", sl.Err(err), slog.Any("payload", payload))
		w.WriteHeader(500)

		return
	}

	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		log.Error("failed to write data", sl.Err(err), slog.Any("payload", payload))
		w.WriteHeader(500)

		return
	}
}
