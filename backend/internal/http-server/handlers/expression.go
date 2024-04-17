package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator"

	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
)

// HandlerCreateExpression is a http.Handler to create new expression.
func HandlerCreateExpression(
	log *slog.Logger,
	dbCfg *storage.Storage,
	orc *orchestrator.Orchestrator,
	producer brokers.Producer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerCreateExpression"

		log := log.With(
			slog.String("fn", fn),
		)

		type parametrs struct {
			Data string `json:"data"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parametrs{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("error parsing JSON: %v", err))
			return
		}

		parseData, err := parser.ParseExpression(params.Data)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("error parsing expression: %v", err))
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
			respondWithError(log, w, 400, fmt.Sprintf("can't create expression: %v", err))
			return
		}

		msgToQueue := messages.ExpressionMessage{
			ExpressionID: expression.ExpressionID,
			Expression:   parseData,
		}

		orc.AddTask(msgToQueue, producer)

		log.Info("send message from orchestrator to agent agregator")

		respondWithJson(log, w, 201, postgres.DatabaseExpressionToExpression(expression))
	}
}

// HandlerGetExpressions is a http.Handler to get all expressions from storage.
func HandlerGetExpressions(log *slog.Logger, dbCfg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerCreateExpression"

		log := log.With(
			slog.String("fn", fn),
		)

		expressions, err := dbCfg.DB.GetExpressions(r.Context())
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("Couldn't get expressions: %v", err))
			return
		}

		respondWithJson(log, w, 200, postgres.DatabaseExpressionsToExpressions(expressions))
	}
}
