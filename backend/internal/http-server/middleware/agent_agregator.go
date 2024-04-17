package middleware

import (
	"net/http"

	agentagregator "github.com/Prrromanssss/DAEE-fullstack/internal/agent_agregator"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
)

type agentAgregatorAndDBConfigHandler func(http.ResponseWriter, *http.Request, *storage.Storage, *agentagregator.AgentAgregator)

func MiddlewareAgentAgregatorAndDBConfig(
	handler agentAgregatorAndDBConfigHandler,
	dbCfg *storage.Storage,
	agentAgr *agentagregator.AgentAgregator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbCfg, agentAgr)
	}
}
