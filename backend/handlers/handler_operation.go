package handlers

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandlerGetOperations(w http.ResponseWriter, r *http.Request, apiCfg *config.ApiConfig) {
	operations, err := apiCfg.DB.GetOperations(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get operations: %v", err))
		return
	}
	respondWithJson(w, 200, database.DatabaseOperationsToOperations(operations))
}

func HandlerUpdateOperation(w http.ResponseWriter, r *http.Request, apiCfg *config.ApiConfig) {
	type parametrs struct {
		OperationType string `json:"operation_type"`
		ExecutionTime int64  `json:"execution_time"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parametrs{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
	}

	operation, err := apiCfg.DB.UpdateOperationTime(r.Context(), database.UpdateOperationTimeParams{
		OperationType: params.OperationType,
		ExecutionTime: params.ExecutionTime,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't update operation: %v", err))
		return
	}

	respondWithJson(w, 200, database.DatabaseOperationToOperation(operation))
}
