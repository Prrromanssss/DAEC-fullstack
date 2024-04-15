package handlers

import (
	"fmt"
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
)

func HandlerGetAgents(w http.ResponseWriter, r *http.Request, dbCfg *storage.Storage) {
	agents, err := dbCfg.DB.GetAgents(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get agents: %v", err))
		return
	}
	respondWithJson(w, 200, postgres.DatabaseAgentsToAgents(agents))
}
