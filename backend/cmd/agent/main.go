package agentservice

import (
	"context"
	"log/slog"
	"time"

	ag "github.com/Prrromanssss/DAEE-fullstack/internal/agent"
	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/rabbitmq"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
)

// RunAgent makes all requirements needed to run agent and call AgentService to run it.
func RunAgent(log *slog.Logger, cfg *config.Config, dbCfg *storage.Storage) {
	const fn = "agentservice.RunAgent"

	log = log.With(
		slog.String("fn", fn),
	)

	ctxWithCancel, cancel := context.WithCancel(context.Background())

	amqpCfg, err := rabbitmq.NewAMQPConfig(log, cfg.RabbitMQURL)
	if err != nil {
		cancel()
		log.Error("can't create NewAMQPConfig", sl.Err(err))
		return
	}

	producer, err := rabbitmq.NewAMQPProducer(log, amqpCfg, cfg.QueueForResultsFromAgents)
	if err != nil {
		cancel()
		log.Error("can't create NewAMQPProducer", sl.Err(err))
		return
	}

	consumer, err := rabbitmq.NewAMQPConsumer(log, amqpCfg, cfg.QueueForExpressionsToAgents)
	if err != nil {
		cancel()
		log.Error("can't create NewAMQPConsumer", sl.Err(err))
		return
	}

	agent, err := ag.NewAgent(
		log,
		dbCfg,
		postgres.Agent{},
		200,
		cancel,
	)
	if err != nil {
		cancel()
		log.Error("can't create agent", sl.Err(err))
		return
	}

	go AgentService(ctxWithCancel, log, amqpCfg, producer, consumer, agent)
}

// AgentService gets messages from SimpleComputers, handle these messages,
// sends pings to Agent Agregator.
func AgentService(
	ctx context.Context,
	log *slog.Logger,
	amqpCfg *rabbitmq.AMQPConfig,
	producer brokers.Producer,
	consumer brokers.Consumer,
	agent *ag.Agent,
) {
	defer func() {
		amqpCfg.Close()
		producer.Close()
		consumer.Close()
		agent.MakeExpressionsTerminated(ctx)
	}()

	go func() {
		for msgFromAgentAgregator := range consumer.GetMessages() {
			go agent.ConsumeMessageFromAgentAgregator(ctx, msgFromAgentAgregator)
		}
	}()

	ticker := time.NewTicker(time.Duration(agent.InactiveTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case result := <-agent.SimpleComputers:
			go agent.ConsumeMessageFromComputers(ctx, result, producer)
		case <-ctx.Done():
			agent.Terminate()
			return
		case <-ticker.C:
			agent.Ping(producer)
		}
	}
}
