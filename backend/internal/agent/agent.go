package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/rabbitmq"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"

	"github.com/streadway/amqp"
)

type Producer interface {
	PublishMessage(msg *messages.ExpressionMessage) error
	Reconnect() (*rabbitmq.AMQPProducer, error)
}

type Consumer interface {
	Messages() <-chan amqp.Delivery
}

type Agent struct {
	Producer
	Consumer
	postgres.Agent
	log             *slog.Logger
	dbConfig        *storage.Storage
	InactiveTime    int32
	SimpleComputers chan *messages.ExpressionMessage
	mu              *sync.Mutex
	kill            context.CancelFunc
}

func NewAgent(
	log *slog.Logger,
	prod Producer,
	cons Consumer,
	dbCfg *storage.Storage,
	agentObj postgres.Agent,
	inactiveTime int32,
	kill context.CancelFunc,
) (*Agent, error) {
	return &Agent{
		log:             log,
		Producer:        prod,
		Consumer:        cons,
		Agent:           agentObj,
		InactiveTime:    inactiveTime,
		dbConfig:        dbCfg,
		SimpleComputers: make(chan *messages.ExpressionMessage),
		mu:              &sync.Mutex{},
		kill:            kill,
	}, nil
}

// GetSafelyNumberOfActiveCalculations gets NumberOfActiveCalculations with Lock.
func (a *Agent) GetSafelyNumberOfActiveCalculations() int32 {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.NumberOfActiveCalculations
}

// GetSafelyNumberOfParallelCalculations gets NumberOfParallelCalculations with Lock.
func (a *Agent) GetSafelyNumberOfParallelCalculations() int32 {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.NumberOfParallelCalculations
}

// Terminate changes agent status to terminate
func (a *Agent) Terminate() {
	const fn = "agent.Terminate"

	err := a.dbConfig.DB.UpdateAgentStatus(context.Background(), postgres.UpdateAgentStatusParams{
		AgentID: a.AgentID,
		Status:  "terminated",
	})
	if err != nil {
		a.log.Error("can't terminate agent", slog.String("fn", fn), slog.Int("agentID", int(a.AgentID)), sl.Err(err))

		return
	}
}

// Ping sends pings to queue
func (a *Agent) Ping() {
	const fn = "agent.Ping"

	exprMsg := messages.ExpressionMessage{
		IsPing:  true,
		AgentID: a.AgentID,
	}
	err := a.Producer.PublishMessage(&exprMsg)
	if err != nil {
		a.log.Error("can't send ping", slog.String("fn", fn), sl.Err(err))

		return
	}
}

// RunSimpleComputer parses messages.ExpressionMessage and run SimpleComputer.
func (a *Agent) RunSimpleComputer(ctx context.Context, exprMsg *messages.ExpressionMessage) error {
	const fn = "agent.RunSimpleComputer"

	tokenSplit := strings.Split(exprMsg.Token, " ")
	if len(tokenSplit) != 3 {
		return fmt.Errorf("invalid token, fn: %s", fn)
	}
	oper := tokenSplit[2]
	if !(oper == "+" || oper == "-" || oper == "/" || oper == "*") {
		return fmt.Errorf("operation in token doesn't match any of these +, -, /, *, fn: %s", fn)
	}

	digit1, err := strconv.Atoi(tokenSplit[0])
	if err != nil {
		return fmt.Errorf("can't convert int to str: %v, fn: %s", err, fn)
	}
	digit2, err := strconv.Atoi(tokenSplit[1])
	if err != nil {
		return fmt.Errorf("can't convert int to str: %v, fn: %s", err, fn)
	}

	time_for_oper, err := a.dbConfig.DB.GetOperationTimeByType(ctx, oper)
	if err != nil {
		return fmt.Errorf("can't get execution time by operation type: %v, fn: %s", err, fn)
	}

	timer := time.NewTimer(time.Duration(time_for_oper) * time.Second)

	go simpleComputer(exprMsg, digit1, digit2, oper, timer, a.SimpleComputers)

	err = a.dbConfig.DB.IncrementNumberOfActiveCalculations(ctx, a.AgentID)
	if err != nil {
		return fmt.Errorf("can't increment number of active calculations: %v, fn: %s", err, fn)
	}

	atomic.AddInt32(&a.NumberOfActiveCalculations, 1)

	return nil
}

// DecrementActiveComputers decrements NumberOfActiveCalculations and changes agent Status.
func (a *Agent) DecrementActiveComputers(ctx context.Context) error {
	const fn = "agent.DecrementActiveComputers"

	err := a.dbConfig.DB.DecrementNumberOfActiveCalculations(ctx, a.AgentID)
	if err != nil {
		return fmt.Errorf("can't decrement number of active calculations, fn: %s", fn)
	}
	atomic.AddInt32(&a.NumberOfActiveCalculations, -1)
	if a.GetSafelyNumberOfActiveCalculations() == 0 {
		err := a.dbConfig.DB.UpdateAgentStatus(
			ctx,
			postgres.UpdateAgentStatusParams{
				Status:  "waiting",
				AgentID: a.AgentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v, fn: %s", err, fn)
		}
		a.Status = "waiting"
	} else {
		err := a.dbConfig.DB.UpdateAgentStatus(
			ctx,
			postgres.UpdateAgentStatusParams{
				Status:  "running",
				AgentID: a.AgentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v, fn: %s", err, fn)
		}
		a.Status = "running"
	}

	return nil
}

// ChangeAgentStatusToRunningOrSleeping changes agent Status to "running" or "sleeping".
func (a *Agent) ChangeAgentStatusToRunningOrSleeping(ctx context.Context) error {
	const fn = "agent.ChangeAgentStatusToRunningOrSleeping"

	if a.GetSafelyNumberOfActiveCalculations() == a.GetSafelyNumberOfParallelCalculations() {
		err := a.dbConfig.DB.UpdateAgentStatus(
			ctx,
			postgres.UpdateAgentStatusParams{
				Status:  "sleeping",
				AgentID: a.AgentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v, fn: %s", err, fn)
		}
		a.Status = "sleeping"
	} else if a.Status != "running" {
		err := a.dbConfig.DB.UpdateAgentStatus(
			ctx,
			postgres.UpdateAgentStatusParams{
				Status:  "running",
				AgentID: a.AgentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v, fn: %s", err, fn)
		}
		a.Status = "running"
	}

	return nil
}

// ChangeExpressionStatus changes expression status to newStatus.
func (a *Agent) ChangeExpressionStatus(ctx context.Context, exprID int32, newStatus string) error {
	const fn = "agent.ChangeExpressionStatus"

	err := a.dbConfig.DB.UpdateExpressionStatus(
		ctx,
		postgres.UpdateExpressionStatusParams{
			ExpressionID: exprID,
			Status:       postgres.ExpressionStatus(newStatus),
		})
	if err != nil {
		return fmt.Errorf("can't update expression status: %v, fn: %s", err, fn)
	}
	return nil
}

// MakeExpressionsTerminated makes all expressions by this agent terminated.
func (a *Agent) MakeExpressionsTerminated(ctx context.Context) {
	const fn = "agent.MakeExpressionsTerminated"

	err := a.dbConfig.DB.MakeExpressionsTerminated(ctx, sql.NullInt32{Int32: a.AgentID, Valid: true})
	if err != nil {
		a.log.Error("can't make expressions terminated", slog.String("fn", fn))
	}
}

// ConsumeMessageFromComputers handles message from simple computers.
// Producer publishes it to queue.
func (a *Agent) ConsumeMessageFromComputers(ctx context.Context, result *messages.ExpressionMessage) {
	const fn = "agent.ConsumeMessageFromComputers"

	log := a.log.With(
		slog.String("fn", fn),
	)

	log.Info("agent consumes message from computers", slog.Any("message", result))

	err := a.Producer.PublishMessage(result)
	if err != nil {
		newProd, err := a.Producer.Reconnect()
		a.Producer = newProd
		if err != nil {
			log.Error("agent error", sl.Err(err))
			a.kill()
			return
		}
		err = a.Producer.PublishMessage(result)
		if err != nil {
			log.Error("agent error", sl.Err(err))
			a.kill()
			return
		}
	}

	err = a.DecrementActiveComputers(ctx)
	if err != nil {
		log.Error("agent error", sl.Err(err))
		a.kill()
		return
	}
}

// ConsumeMessageFromAgentAgregator hanldes message from Consumer.
func (a *Agent) ConsumeMessageFromAgentAgregator(ctx context.Context, msgFromAgentAgregator amqp.Delivery) {
	const fn = "agent.ConsumeMessageFromAgentAgregator"

	log := a.log.With(
		slog.String("fn", fn),
	)

	log.Info("agent consumes msg from agent agregator", slog.String("message", string(msgFromAgentAgregator.Body)))

	var exprMsg messages.ExpressionMessage
	if err := json.Unmarshal(msgFromAgentAgregator.Body, &exprMsg); err != nil {
		log.Error("agent error: failed to parse JSON", sl.Err(err))
		a.kill()
		return
	}

	if a.GetSafelyNumberOfActiveCalculations() >= a.GetSafelyNumberOfParallelCalculations() {
		err := msgFromAgentAgregator.Nack(false, true)
		if err != nil {
			log.Error("agent error", sl.Err(err))
			a.kill()
			return
		}
		return
	}

	err := msgFromAgentAgregator.Ack(false)
	if err != nil {
		log.Error("agent error: error acknowledging message", sl.Err(err))
		a.kill()
		return
	}

	err = a.ChangeExpressionStatus(ctx, exprMsg.ExpressionID, "computing")
	if err != nil {
		log.Error("agent error", sl.Err(err))
		a.kill()
		return
	}

	err = a.RunSimpleComputer(ctx, &exprMsg)
	if err != nil {
		log.Error("agent error", sl.Err(err))
		a.kill()
		return
	}

	err = a.ChangeAgentStatusToRunningOrSleeping(ctx)
	if err != nil {
		log.Error("agent error", sl.Err(err))
		a.kill()
		return
	}
}
