package orchestratorapp

import (
	"context"
	"log/slog"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"

	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator"
	"github.com/Prrromanssss/DAEE-fullstack/internal/rabbitmq"
)

type App struct {
	log             *slog.Logger
	OrchestratorApp *orchestrator.Orchestrator
	amqpConfig      rabbitmq.AMQPConfig
	Producer        brokers.Producer
	Consumer        brokers.Consumer
}

// MustRun runs Orchestrator and panics if any error occurs.
func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}

// New creates new Orchestrator app.
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

	producer, err := rabbitmq.NewAMQPProducer(log, amqpCfg, cfg.QueueForExpressionsToAgents)
	if err != nil {
		log.Error("can't create NewAMQPProducer", sl.Err(err))
		return nil, err
	}

	consumer, err := rabbitmq.NewAMQPConsumer(log, amqpCfg, cfg.QueueForResultsFromAgents)
	if err != nil {
		log.Error("can't create NewAMQPConsumer", sl.Err(err))
		return nil, err
	}

	orc, err := orchestrator.NewOrchestrator(
		log,
		dbCfg,
		cfg.InactiveTimeForAgent,
		cancel,
	)
	if err != nil {
		log.Error("orchestrator error", sl.Err(err))
		return nil, err
	}

	return &App{
		log:             log,
		OrchestratorApp: orc,
		amqpConfig:      *amqpCfg,
		Producer:        producer,
		Consumer:        consumer,
	}, nil
}

// RunOrchestrator agregates agents,
// consumes messages from client, manages their job.
func (a *App) Run(ctx context.Context) error {
	defer func() {
		a.amqpConfig.Close()
		a.Producer.Close()
		a.Consumer.Close()
	}()

	const fn = "orchestratorapp.Run"

	log := a.log.With(
		slog.String("fn", fn),
	)

	// Reload not completed expressions.
	err := a.OrchestratorApp.ReloadComputingExpressions(ctx, a.Producer)
	if err != nil {
		log.Error("can't reload computing expressions", sl.Err(err))

		return err
	}

	ticker := time.NewTicker(time.Duration(a.OrchestratorApp.InactiveTimeForAgent) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msgFromAgents := <-a.Consumer.GetMessages():
			go a.OrchestratorApp.HandleMessagesFromAgents(ctx, msgFromAgents, a.Producer)
		case <-ticker.C:
			err := a.OrchestratorApp.CheckPing(ctx, a.Producer)
			if err != nil {
				log.Warn("can't check pings from agents", sl.Err(err))
			}
		case <-ctx.Done():
			log.Error("orchestrator stopped")

			return err
		}
	}
}
