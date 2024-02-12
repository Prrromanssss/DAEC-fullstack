package handlers

import (
	"Prrromanssss/DAEE/internal/jsonhandler"
	"net/http"
)

func HandlerGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	jsonhandler.RespondWithJson(w, 200, "Hello from HandlerGetExpressionByID")
}
