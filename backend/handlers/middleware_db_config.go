package handlers

import (
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/config"
)

type dbConfigHandler func(http.ResponseWriter, *http.Request, *config.DBConfig)

func MiddlewareApiConfig(handler dbConfigHandler, dbCfg *config.DBConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbCfg)
	}
}
