package reload

import (
	"context"
	"fmt"

	"github.com/Prrromanssss/DAEE-fullstack/config"
	"github.com/Prrromanssss/DAEE-fullstack/pkg/agent"
)

func ReloadComputingExpressions(dbCfg *config.DBConfig, agentAgr *agent.AgentAgregator) error {
	expressions, err := dbCfg.DB.GetComputingExpressions(context.Background())
	if err != nil {
		return fmt.Errorf("orhestrator Error: %v", err)
	}
	for _, expr := range expressions {
		msgToQueue := agent.MessageFromOrchestrator{
			ExpressionID: expr.ID,
			Expression:   expr.ParseData,
		}
		agentAgr.AddTask(msgToQueue)
	}
	return nil
}
