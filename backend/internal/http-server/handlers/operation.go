package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/jwt"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
)

// HandlerGetOperations is a http.Handler to get all operations from storage.
func HandlerGetOperations(log *slog.Logger, dbCfg *storage.Storage, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "hadlers.HandlerGetOperations"

		log := log.With(
			slog.String("fn", fn),
		)

		userID, err := jwt.GetUidFromJWT(r, secret)
		if err != nil {
			respondWithError(log, w, 403, "Status Forbidden")
			return
		}

		operations, err := dbCfg.Queries.GetOperations(r.Context(), userID)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("can't get operations: %v", err))
			return
		}

		respondWithJson(log, w, 200, postgres.DatabaseOperationsToOperations(operations))
	}
}

// HandlerUpdateOperation is a http.Handler to update execution time of the certain operation type.
func HandlerUpdateOperation(log *slog.Logger, dbCfg *storage.Storage, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerUpdateOperation"

		log := log.With(
			slog.String("fn", fn),
		)

		userID, err := jwt.GetUidFromJWT(r, secret)
		if err != nil {
			respondWithError(log, w, 403, "Status Forbidden")
			return
		}

		type parametrs struct {
			OperationType string `json:"operation_type"`
			ExecutionTime int32  `json:"execution_time"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parametrs{}
		err = decoder.Decode(&params)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("error parsing JSON: %v", err))
		}

		operation, err := dbCfg.Queries.UpdateOperationTime(r.Context(), postgres.UpdateOperationTimeParams{
			OperationType: params.OperationType,
			ExecutionTime: params.ExecutionTime,
			UserID:        userID,
		})

		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("can't update operation: %v", err))
			return
		}

		respondWithJson(log, w, 200, postgres.DatabaseOperationToOperation(operation))
	}
}
