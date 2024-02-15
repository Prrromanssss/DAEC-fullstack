package handlers

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/pkg/agent"
	"net/http"
)

type agentAgregatorAndDBConfigHandler func(http.ResponseWriter, *http.Request, *config.DBConfig, *agent.AgentAgregator)

func MiddlewareAgentAgregatorAndDBConfig(handler agentAgregatorAndDBConfigHandler, dbCfg *config.DBConfig, agentAgr *agent.AgentAgregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbCfg, agentAgr)
	}
}
