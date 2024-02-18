package handlers

import (
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/config"
	"github.com/Prrromanssss/DAEE-fullstack/pkg/agent"
)

type agentAgregatorAndDBConfigHandler func(http.ResponseWriter, *http.Request, *config.DBConfig, *agent.AgentAgregator)

func MiddlewareAgentAgregatorAndDBConfig(
	handler agentAgregatorAndDBConfigHandler,
	dbCfg *config.DBConfig,
	agentAgr *agent.AgentAgregator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbCfg, agentAgr)
	}
}
