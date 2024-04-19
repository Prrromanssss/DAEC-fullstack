package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
)

// HandlerGetAgents is a http.Handler to get all agents from storage.
func HandlerGetAgents(log *slog.Logger, dbCfg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerGetAgents"

		log := log.With(
			slog.String("fn", fn),
		)

		agents, err := dbCfg.Queries.GetAgents(r.Context())
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("couldn't get agents: %v", err))
			return
		}

		respondWithJson(log, w, 200, postgres.DatabaseAgentsToAgents(agents))
	}
}
