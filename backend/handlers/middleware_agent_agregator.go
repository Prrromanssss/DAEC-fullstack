package handlers

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/pkg/agent"
	"net/http"
)

type agentAgregatorAndApiConfigHandler func(http.ResponseWriter, *http.Request, *config.ApiConfig, *agent.AgentAgregator)

func MiddlewareAgentAgregatorAndApiConfig(handler agentAgregatorAndApiConfigHandler, apiCfg *config.ApiConfig, agentAgr *agent.AgentAgregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, apiCfg, agentAgr)
	}
}
