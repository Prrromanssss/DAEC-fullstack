package handlers

import (
	"fmt"
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/database"
)

func HandlerGetAgents(w http.ResponseWriter, r *http.Request, dbCfg *config.DBConfig) {
	agents, err := dbCfg.DB.GetAgents(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get agents: %v", err))
		return
	}
	respondWithJson(w, 200, database.DatabaseAgentsToAgents(agents))
}
