package handlers

import (
	"Prrromanssss/DAEE/config"
	"net/http"
)

type apiConfigHandler func(http.ResponseWriter, *http.Request, *config.ApiConfig)

func MiddlewareApiConfig(handler apiConfigHandler, apiCfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, apiCfg)
	}
}
