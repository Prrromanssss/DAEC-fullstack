package config

import (
	"Prrromanssss/DAEE/internal/database"
	"context"
	"log"

	"github.com/google/uuid"
)

func createOperation(dbCfg *DBConfig, operationType string, executionTime int64) {
	err := dbCfg.DB.CreateOperation(context.Background(), database.CreateOperationParams{
		ID:            uuid.New(),
		OperationType: operationType,
		ExecutionTime: executionTime,
	})
	if err != nil {
		log.Printf("Can't create operation: %v", err)
	}
}

func ConfigOperation(dbCfg *DBConfig) {
	_, err := dbCfg.DB.GetOperationByType(context.Background(), "+")
	if err.Error() == "sql: no rows in result set" {
		createOperation(dbCfg, "+", 10)
		createOperation(dbCfg, "-", 10)
		createOperation(dbCfg, "*", 10)
		createOperation(dbCfg, "/", 10)
	}

}
