package auth

import (
	"Prrromanssss/DAEE/internal/database"
	"Prrromanssss/DAEE/internal/jsonhandler"
	"fmt"
	"net/http"
)

type AuthedHandler func(http.ResponseWriter, *http.Request, database.User)

func (apiCfg *ApiConfig) MiddlewareAuth(handler AuthedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := GetAPIKey(r.Header)
		if err != nil {
			jsonhandler.RespondWithError(w, 403, fmt.Sprintf("Auth error: %v", err))
			return
		}

		user, err := apiCfg.DB.GetUserByAPIkey(r.Context(), apiKey)
		if err != nil {
			jsonhandler.RespondWithError(w, 400, fmt.Sprintf("Couldn't get user: %v", err))
			return
		}
		handler(w, r, user)
	}
}
