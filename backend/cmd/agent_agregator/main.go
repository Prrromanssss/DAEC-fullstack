package agentagregatorservice

import (
	"context"
	"log/slog"

	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"

	agentagregator "github.com/Prrromanssss/DAEE-fullstack/internal/agent_agregator"
	"github.com/Prrromanssss/DAEE-fullstack/internal/rabbitmq"
)

// RunAgentAgregator makes all requirements needed to run agent and call AgregateAgents to run it.
func RunAgentAgregator(
	log *slog.Logger,
	cfg *config.Config,
	dbCfg *storage.Storage,
) (*agentagregator.AgentAgregator, error) {
	const fn = "agentagregatorservice.RunAgentAgregator"

	log = log.With(
		slog.String("fn", fn),
	)

	ctxWithCancel, cancel := context.WithCancel(context.Background())

	amqpCfg, err := rabbitmq.NewAMQPConfig(log, cfg.RabbitMQURL)
	if err != nil {
		cancel()
		log.Error("can't create NewAMQPConfig", sl.Err(err))
		return nil, err
	}

	producer, err := rabbitmq.NewAMQPProducer(log, amqpCfg, cfg.QueueForExpressionsToAgents)
	if err != nil {
		cancel()
		log.Error("can't create NewAMQPProducer", sl.Err(err))
		return nil, err
	}

	consumer, err := rabbitmq.NewAMQPConsumer(log, amqpCfg, cfg.QueueForResultsFromAgents)
	if err != nil {
		cancel()
		log.Error("can't create NewAMQPConsumer", sl.Err(err))
		return nil, err
	}

	agentAgregator, err := agentagregator.NewAgentAgregator(
		log,
		dbCfg,
		cancel,
	)
	if err != nil {
		cancel()
		log.Error("agent agregator error", sl.Err(err))
		return nil, err
	}

	err = agentAgregator.ReloadComputingExpressions(ctxWithCancel)
	if err != nil {
		cancel()
		log.Error("can't reload computing expressions", sl.Err(err))
		return nil, err
	}

	go AgregateAgents(ctxWithCancel, log, amqpCfg, producer, consumer, agentAgregator)

	return agentAgregator, nil
}

// AgregateAgents agregates agents,
// consumes messages from them and orchestrator, manages their job.
func AgregateAgents(
	ctx context.Context,
	log *slog.Logger,
	amqpCfg *rabbitmq.AMQPConfig,
	producer brokers.Producer,
	consumer brokers.Consumer,
	agentAg *agentagregator.AgentAgregator,
) {
	defer func() {
		amqpCfg.Close()
		producer.Close()
		consumer.Close()
	}()

	go func() {
		for msgFromAgents := range consumer.GetMessages() {
			go agentAg.ConsumeMessagesFromAgents(ctx, msgFromAgents, producer)
		}
	}()

	for {
		select {
		case expressionMessage := <-agentAg.Tasks:
			go agentAg.ConsumeMessagesFromOrchestrator(expressionMessage, producer)
		case <-ctx.Done():
			log.Error("agent agregator stopped")

			// os.Exit(1)
			return
		}

	}
}
