package orchestratorapp

import (
	"context"
	"log/slog"
	"time"

	"github.com/Prrromanssss/DAEC-fullstack/internal/config"
	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/pool"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage"
	"github.com/streadway/amqp"

	"github.com/Prrromanssss/DAEC-fullstack/internal/orchestrator"
	"github.com/Prrromanssss/DAEC-fullstack/internal/rabbitmq"
)

type App struct {
	log                *slog.Logger
	OrchestratorApp    *orchestrator.Orchestrator
	workerPool         *pool.MyPool
	amqpConfig         rabbitmq.AMQPConfig
	channelForProducer *amqp.Channel
	channelForConsumer *amqp.Channel
	Producer           brokers.Producer
	Consumer           brokers.Consumer
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
	// Create worker pool with 5 workers.
	workerPool, err := pool.NewWorkerPool(5, 10)
	if err != nil {
		log.Error("can't create worker pool", sl.Err(err))
	}

	return &App{
		log:                log,
		OrchestratorApp:    orc,
		workerPool:         workerPool,
		amqpConfig:         *amqpCfg,
		channelForProducer: producer.Channel,
		channelForConsumer: producer.Channel,
		Producer:           producer,
		Consumer:           consumer,
	}, nil
}

// RunOrchestrator agregates agents,
// consumes messages from client, manages their job.
func (a *App) Run(ctx context.Context) error {
	defer func() {
		a.amqpConfig.Close()
		a.Producer.Close()
		a.Consumer.Close()
		a.workerPool.Stop()
	}()

	const fn = "orchestratorapp.Run"

	log := a.log.With(
		slog.String("fn", fn),
	)

	a.workerPool.Start()

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
		// TODO: Need to syncronize goroutines
		case msgFromAgents := <-a.Consumer.GetMessages():
			task := orchestrator.ExecuteWrapper(a.OrchestratorApp, ctx, msgFromAgents, a.Producer)
			a.workerPool.AddWork(task)
			time.Sleep(time.Second)
		case <-ticker.C:
			err := a.OrchestratorApp.CheckPing(ctx, a.Producer)
			if err != nil {
				log.Warn("can't check pings from agents", sl.Err(err))
			}

			err = a.OrchestratorApp.FindForgottenExpressions(ctx, a.Producer)
			if err != nil {
				log.Warn("can't find forgotten expressions", sl.Err(err))
			}
		case <-ctx.Done():
			log.Error("orchestrator stopped")

			return ctx.Err()
		}
	}
}

// // Stop stops Orchestrator app.
func (a *App) Stop(ctx context.Context, cfg *config.Config) {
	if _, err := a.channelForConsumer.QueuePurge(cfg.QueueForResultsFromAgents, false); err != nil {
		a.log.Error("can't purged queue", slog.String("queue", cfg.QueueForResultsFromAgents))
	}
	if _, err := a.channelForProducer.QueuePurge(cfg.QueueForExpressionsToAgents, false); err != nil {
		a.log.Error("can't purged queue", slog.String("queue", cfg.QueueForExpressionsToAgents))
	}
	if err := a.Producer.PublishExpressionMessage(&messages.ExpressionMessage{
		Kill: true,
	}); err != nil {
		a.log.Error("can't send kill message to agent")
	}
	if err := a.Producer.PublishExpressionMessage(&messages.ExpressionMessage{
		Kill: true,
	}); err != nil {
		a.log.Error("can't send kill message to agent")
	}
	if err := a.Producer.PublishExpressionMessage(&messages.ExpressionMessage{
		Kill: true,
	}); err != nil {
		a.log.Error("can't send kill message to agent")
	}
}
