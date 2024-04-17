package middleware

import (
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
)

type dbConfigHandler func(http.ResponseWriter, *http.Request, *storage.Storage)

func MiddlewareApiConfig(handler dbConfigHandler, dbCfg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbCfg)
	}
}
