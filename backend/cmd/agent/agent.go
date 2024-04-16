package agentservice

import (
	"context"
	"log/slog"
	"time"

	ag "github.com/Prrromanssss/DAEE-fullstack/internal/agent"
)

func AgentService(ctx context.Context, log *slog.Logger, agent *ag.Agent) {
	defer agent.MakeExpressionsTerminated(ctx)

	go func() {
		for msgFromAgentAgregator := range agent.Consumer.Messages() {
			go agent.ConsumeMessageFromAgentAgregator(ctx, msgFromAgentAgregator)
		}
	}()

	ticker := time.NewTicker(time.Duration(agent.InactiveTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case result := <-agent.SimpleComputers:
			go agent.ConsumeMessageFromComputers(ctx, result)
		case <-ctx.Done():
			agent.Terminate()
			return
		case <-ticker.C:
			agent.Ping()
		}
	}
}
