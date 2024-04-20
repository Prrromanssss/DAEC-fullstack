package orchestrator

import (
	"context"

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/pool"
	"github.com/streadway/amqp"
)

type orchestratorTask struct {
	orchestrator  *Orchestrator
	ctx           context.Context
	msgFromAgents amqp.Delivery
	producer      brokers.Producer
}

// ExecuteWrapper is a wrapper function to call the Execute method with the necessary arguments.
func ExecuteWrapper(o *Orchestrator, ctx context.Context, msgFromAgents amqp.Delivery, producer brokers.Producer) pool.PoolTask {
	return &orchestratorTask{o, ctx, msgFromAgents, producer}
}

// Execute implements the Execute method of the PoolTask interface
func (ot *orchestratorTask) Execute() error {
	return ot.orchestrator.HandleMessagesFromAgents(ot.ctx, ot.msgFromAgents, ot.producer)
}

// OnFailure implements the OnFailure method of the PoolTask interface
func (ot *orchestratorTask) OnFailure(err error) {
	ot.orchestrator.log.Error("orchestrator error", sl.Err(err))
	ot.orchestrator.kill()
}
