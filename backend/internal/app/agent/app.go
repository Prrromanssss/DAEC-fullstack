package agentapp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Prrromanssss/DAEC-fullstack/internal/agent"
	"github.com/Prrromanssss/DAEC-fullstack/internal/config"
	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEC-fullstack/internal/rabbitmq"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage"
	"github.com/streadway/amqp"
)

type App struct {
	log         *slog.Logger
	AgentApp    *agent.Agent
	mu          *sync.Mutex
	TimeForPing int32
	amqpConfig  rabbitmq.AMQPConfig
	Producer    brokers.Producer
	Consumer    brokers.Consumer
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
		cancel,
	)
	if err != nil {
		log.Error("can't create agent", sl.Err(err))
		return nil, err
	}

	return &App{
		log:         log,
		AgentApp:    ag,
		mu:          &sync.Mutex{},
		TimeForPing: cfg.TimeForPing,
		amqpConfig:  *amqpCfg,
		Producer:    producer,
		Consumer:    consumer,
	}, nil
}

// Run gets messages from SimpleComputers, handle these messages,
// sends pings to Agent Agregator.
func (a *App) Run(ctx context.Context) error {
	defer func() {
		a.amqpConfig.Close()
		a.Producer.Close()
		a.Consumer.Close()
	}()

	go func() {
		for msgFromOrchestrator := range a.Consumer.GetMessages() {
			a.mu.Lock()
			if a.AgentApp.NumberOfActiveCalculations >= a.AgentApp.NumberOfParallelCalculations {
				a.mu.Unlock()
				err := msgFromOrchestrator.Nack(false, true)
				if err != nil {
					a.log.Error("can't nack message", sl.Err(err))
					return
				}
				continue // skip the processing of this message and move on to the next one.
			}
			a.AgentApp.NumberOfActiveCalculations++
			a.mu.Unlock()

			go func(msgFromOrchestrator amqp.Delivery) {
				a.AgentApp.ConsumeMessageFromOrchestrator(ctx, msgFromOrchestrator)
			}(msgFromOrchestrator)
		}
	}()

	ticker := time.NewTicker(time.Duration(a.TimeForPing) * time.Second)
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

// Stop stops Agent app.
func (a *App) Stop(ctx context.Context) {
	a.AgentApp.Terminate()
}
