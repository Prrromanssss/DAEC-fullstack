package middleware

import (
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/internal/agent"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
)

type agentAgregatorAndDBConfigHandler func(http.ResponseWriter, *http.Request, *storage.Storage, *agent.AgentAgregator)

func MiddlewareAgentAgregatorAndDBConfig(
	handler agentAgregatorAndDBConfigHandler,
	dbCfg *storage.Storage,
	agentAgr *agent.AgentAgregator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbCfg, agentAgr)
	}
}
