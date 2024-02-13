package handlers

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/internal/database"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func HandlerGetAgents(w http.ResponseWriter, r *http.Request, apiCfg *config.ApiConfig) {
	agents, err := apiCfg.DB.GetAgents(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get agents: %v", err))
		return
	}
	respondWithJson(w, 200, database.DatabaseAgentsToAgents(agents))
}

func HandlerGetAgentByID(w http.ResponseWriter, r *http.Request, apiCfg *config.ApiConfig) {
	agentIDString := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDString)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't parse agent id: %v", err))
		return
	}
	agent, err := apiCfg.DB.GetAgentByID(r.Context(), agentID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get agent: %v", err))
		return
	}
	respondWithJson(w, 200, database.DatabaseAgentToAgent(agent))
}
