package handlers

import (
	"Prrromanssss/DAEE/config"
	"net/http"
)

func HandlerGetComputers(w http.ResponseWriter, r *http.Request, apiCfg *config.ApiConfig) {
	respondWithJson(w, 200, "Hello from HandlerGetComputers")
}
