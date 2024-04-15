package orchestrator

import (
	"context"
	"fmt"

	"github.com/Prrromanssss/DAEE-fullstack/internal/agent"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
)

func ReloadComputingExpressions(dbCfg *storage.Storage, agentAgr *agent.AgentAgregator) error {
	expressions, err := dbCfg.DB.GetComputingExpressions(context.Background())
	if err != nil {
		return fmt.Errorf("orhestrator Error: %v", err)
	}
	for _, expr := range expressions {
		msgToQueue := agent.MessageFromOrchestrator{
			ExpressionID: expr.ExpressionID,
			Expression:   expr.ParseData,
		}
		agentAgr.AddTask(msgToQueue)
	}
	return nil
}
