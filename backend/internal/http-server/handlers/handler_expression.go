package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	agentagregator "github.com/Prrromanssss/DAEE-fullstack/internal/agent_agregator"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"

	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
)

func HandlerCreateExpression(
	w http.ResponseWriter,
	r *http.Request,
	dbCfg *storage.Storage,
	agentAgr *agentagregator.AgentAgregator,
) {
	type parametrs struct {
		Data string `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parametrs{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
	}

	parseData, err := parser.ParseExpression(params.Data)

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing expression: %v", err))
		return
	}

	expression, err := dbCfg.DB.CreateExpression(r.Context(),
		postgres.CreateExpressionParams{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Data:      params.Data,
			ParseData: parseData,
			Status:    "ready for computation",
		})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't create expression: %v", err))
		return
	}

	msgToQueue := messages.MessageFromOrchestrator{
		ExpressionID: expression.ExpressionID,
		Expression:   parseData,
	}

	agentAgr.AddTask(msgToQueue)
	log.Println("Send message from orchestrator to agent agregator")

	respondWithJson(w, 201, postgres.DatabaseExpressionToExpression(expression))
}

func HandlerGetExpressions(w http.ResponseWriter, r *http.Request, dbCfg *storage.Storage) {
	expressions, err := dbCfg.DB.GetExpressions(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get expressions: %v", err))
		return
	}
	respondWithJson(w, 200, postgres.DatabaseExpressionsToExpressions(expressions))
}
