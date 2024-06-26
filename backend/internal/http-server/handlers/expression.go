package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/jwt"
	"github.com/Prrromanssss/DAEC-fullstack/internal/orchestrator"

	"github.com/Prrromanssss/DAEC-fullstack/internal/orchestrator/parser"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage/postgres"
)

// HandlerCreateExpression is a http.Handler to create new expression.
func HandlerCreateExpression(
	log *slog.Logger,
	dbCfg *storage.Storage,
	secret string,
	orc *orchestrator.Orchestrator,
	producer brokers.Producer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerCreateExpression"

		log := log.With(
			slog.String("fn", fn),
		)

		userID, err := jwt.GetUidFromJWT(r, secret)
		if err != nil {
			respondWithError(log, w, 403, "Status Forbidden")
			return
		}

		type parametrs struct {
			Data string `json:"data"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parametrs{}
		err = decoder.Decode(&params)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("error parsing JSON: %v", err))
			return
		}

		parseData, err := parser.ParseExpression(params.Data)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("error parsing expression: %v", err))
			return
		}

		expression, err := dbCfg.Queries.CreateExpression(r.Context(),
			postgres.CreateExpressionParams{
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				Data:      params.Data,
				ParseData: parseData,
				Status:    "ready_for_computation",
				UserID:    userID,
			})
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("can't create expression: %v", err))
			return
		}

		msgToQueue := messages.ExpressionMessage{
			ExpressionID: expression.ExpressionID,
			Expression:   parseData,
			UserID:       userID,
		}

		orc.AddTask(msgToQueue, producer)

		log.Info("send message to orchestrator")

		respondWithJson(log, w, 201, postgres.DatabaseExpressionToExpression(expression))
	}
}

// HandlerGetExpressions is a http.Handler to get all expressions from storage.
func HandlerGetExpressions(log *slog.Logger, dbCfg *storage.Storage, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerCreateExpression"

		log := log.With(
			slog.String("fn", fn),
		)

		userID, err := jwt.GetUidFromJWT(r, secret)
		if err != nil {
			respondWithError(log, w, 403, "Status Forbidden")
			return
		}

		expressions, err := dbCfg.Queries.GetExpressions(r.Context(), userID)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("Couldn't get expressions: %v", err))
			return
		}

		respondWithJson(log, w, 200, postgres.DatabaseExpressionsToExpressions(expressions))
	}
}
