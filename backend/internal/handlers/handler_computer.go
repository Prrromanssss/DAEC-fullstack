package handlers

import (
	"Prrromanssss/DAEE/internal/config"
	"Prrromanssss/DAEE/internal/jsonhandler"
	"net/http"
)

func HandlerGetComputers(w http.ResponseWriter, r *http.Request, apiCfg *config.ApiConfig) {
	jsonhandler.RespondWithJson(w, 200, "Hello from HandlerGetComputers")
}
