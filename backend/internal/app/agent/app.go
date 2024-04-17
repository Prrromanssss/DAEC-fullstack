package agentapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/agent"
	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/rabbitmq"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"
)

type App struct {
	log        *slog.Logger
	AgentApp   *agent.Agent
	amqpConfig rabbitmq.AMQPConfig
	Producer   brokers.Producer
	Consumer   brokers.Consumer
}

// MustRun runs Agent and panics if any error occurs.
func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}

// New creates new Agent app.
func New(
	log *slog.Logger,
	cfg *config.Config,
	dbCfg *storage.Storage,
	cancel context.CancelFunc,
) (*App, error) {
	amqpCfg, err := rabbitmq.NewAMQPConfig(log, cfg.RabbitMQURL)
	if err != nil {
		log.Error("can't create NewAMQPConfig", sl.Err(err))
		return nil, err
	}

	producer, err := rabbitmq.NewAMQPProducer(log, amqpCfg, cfg.QueueForResultsFromAgents)
	if err != nil {
		log.Error("can't create NewAMQPProducer", sl.Err(err))
		return nil, err
	}

	consumer, err := rabbitmq.NewAMQPConsumer(log, amqpCfg, cfg.QueueForExpressionsToAgents)
	if err != nil {
		log.Error("can't create NewAMQPConsumer", sl.Err(err))
		return nil, err
	}

	ag, err := agent.NewAgent(
		log,
		dbCfg,
		postgres.Agent{},
		200,
		cancel,
	)
	if err != nil {
		log.Error("can't create agent", sl.Err(err))
		return nil, err
	}

	return &App{
		log:        log,
		AgentApp:   ag,
		amqpConfig: *amqpCfg,
		Producer:   producer,
		Consumer:   consumer,
	}, nil
}

// Run gets messages from SimpleComputers, handle these messages,
// sends pings to Agent Agregator.
func (a *App) Run(ctx context.Context) error {
	defer func() {
		a.amqpConfig.Close()
		a.Producer.Close()
		a.Consumer.Close()
		a.AgentApp.MakeExpressionsTerminated(ctx)
	}()

	// Delete previous agents
	err := a.AgentApp.DeletePreviousAgents(ctx)
	if err != nil {
		a.log.Error("can't delete previous agents", sl.Err(err))

		return err
	}

	go func() {
		for msgFromAgentAgregator := range a.Consumer.GetMessages() {
			go a.AgentApp.ConsumeMessageFromOrchestrator(ctx, msgFromAgentAgregator)
		}
	}()

	ticker := time.NewTicker(time.Duration(a.AgentApp.InactiveTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case result := <-a.AgentApp.SimpleComputers:
			go a.AgentApp.ConsumeMessageFromComputers(ctx, result, a.Producer)
		case <-ctx.Done():
			a.AgentApp.Terminate()
			return fmt.Errorf("agent terminated")
		case <-ticker.C:
			a.AgentApp.Ping(a.Producer)
		}
	}
}