package config

import (
	"context"
	"log"

	"github.com/Prrromanssss/DAEE-fullstack/internal/database"

	"github.com/google/uuid"
)

func createOperation(dbCfg *DBConfig, operationType string, executionTime int32) {
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
	if err != nil && err.Error() == "sql: no rows in result set" {
		createOperation(dbCfg, "+", 10)
		createOperation(dbCfg, "-", 10)
		createOperation(dbCfg, "*", 10)
		createOperation(dbCfg, "/", 10)
	}

}
