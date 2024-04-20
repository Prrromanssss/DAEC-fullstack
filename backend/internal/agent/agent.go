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

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"

	"github.com/streadway/amqp"
)

type Agent struct {
	postgres.Agent
	log             *slog.Logger
	dbConfig        *storage.Storage
	SimpleComputers chan *messages.ExpressionMessage
	mu              *sync.Mutex
	kill            context.CancelFunc
}

// NewAgent creates new Agent.
func NewAgent(
	log *slog.Logger,
	dbCfg *storage.Storage,
	kill context.CancelFunc,
) (*Agent, error) {
	const fn = "agent.NewAgent"

	agentObj, err := dbCfg.Queries.CreateAgent(context.Background(), postgres.CreateAgentParams{
		CreatedAt:                    time.Now().UTC(),
		NumberOfParallelCalculations: 5,
		LastPing:                     time.Now().UTC(),
		Status:                       "waiting",
	})
	if err != nil {
		log.Error("can't create agent", slog.String("fn", fn), sl.Err(err))

		return nil, err
	}

	log.Info("create agent succesfully", slog.String("fn", fn))

	return &Agent{
		Agent:           agentObj,
		log:             log,
		dbConfig:        dbCfg,
		SimpleComputers: make(chan *messages.ExpressionMessage),
		mu:              &sync.Mutex{},
		kill:            kill,
	}, nil
}

// DeletePreviousAgents deletes previous agents when the application is restarted.
func (a *Agent) DeletePreviousAgents(ctx context.Context) error {
	err := a.dbConfig.Queries.DeleteAgents(ctx)
	if err != nil {
		a.log.Error("can't delete previous agents", sl.Err(err))

		return err
	}

	return nil
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

// Terminate changes agent status to terminate.
func (a *Agent) Terminate() {
	const fn = "agent.Terminate"

	err := a.dbConfig.Queries.UpdateTerminateAgentByID(context.Background(), a.AgentID)
	if err != nil {
		a.log.Error("can't terminate agent", slog.String("fn", fn), slog.Int("agentID", int(a.AgentID)), sl.Err(err))

		return
	}

	a.kill()
}

// Ping sends pings to queue.
func (a *Agent) Ping(producer brokers.Producer) {
	const fn = "agent.Ping"

	log := a.log.With(
		slog.String("fn", fn),
	)

	exprMsg := messages.ExpressionMessage{
		IsPing:  true,
		AgentID: a.AgentID,
	}
	err := producer.PublishExpressionMessage(&exprMsg)
	if err != nil {
		log.Error("can't send ping", sl.Err(err))

		return
	}

	log.Info("agent sends ping to orchestrator", slog.Time("time", time.Now()))
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
	if int(exprMsg.UserID) == 0 {
		a.log.Warn("", slog.String("oper", oper), slog.Int("userID", int(exprMsg.UserID)))
	}
	time_for_oper, err := a.dbConfig.Queries.GetOperationTimeByType(ctx, postgres.GetOperationTimeByTypeParams{
		OperationType: oper,
		UserID:        exprMsg.UserID,
	})
	if err != nil {
		return fmt.Errorf("can't get execution time by operation type: %v, fn: %s", err, fn)
	}

	timer := time.NewTimer(time.Duration(time_for_oper) * time.Second)

	go simpleComputer(exprMsg, digit1, digit2, oper, timer, a.SimpleComputers)

	err = a.dbConfig.Queries.IncrementNumberOfActiveCalculations(ctx, a.AgentID)
	if err != nil {
		return fmt.Errorf("can't increment number of active calculations: %v, fn: %s", err, fn)
	}

	return nil
}

// DecrementActiveComputers decrements NumberOfActiveCalculations and changes agent Status.
func (a *Agent) DecrementActiveComputers(ctx context.Context) error {
	const fn = "agent.DecrementActiveComputers"

	err := a.dbConfig.Queries.DecrementNumberOfActiveCalculations(ctx, a.AgentID)
	if err != nil {
		return fmt.Errorf("can't decrement number of active calculations, fn: %s", fn)
	}
	atomic.AddInt32(&a.NumberOfActiveCalculations, -1)
	if a.GetSafelyNumberOfActiveCalculations() == 0 {
		err := a.dbConfig.Queries.UpdateAgentStatus(
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
		err := a.dbConfig.Queries.UpdateAgentStatus(
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
		err := a.dbConfig.Queries.UpdateAgentStatus(
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
		err := a.dbConfig.Queries.UpdateAgentStatus(
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

	err := a.dbConfig.Queries.UpdateExpressionStatus(
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

// AssignToAgent assigns expression to agent.
func (a *Agent) AssignToAgent(ctx context.Context, exprID int32) error {
	const fn = "agent.AssignToAgent"

	err := a.dbConfig.Queries.AssignExpressionToAgent(ctx, postgres.AssignExpressionToAgentParams{
		AgentID:      sql.NullInt32{Int32: a.AgentID, Valid: true},
		ExpressionID: exprID,
	})
	if err != nil {
		a.log.Error("can't assign expression to agent", slog.String("fn", fn), sl.Err(err))
	}

	return nil
}

// ConsumeMessageFromComputers handles message from simple computers.
// Producer publishes it to queue.
func (a *Agent) ConsumeMessageFromComputers(ctx context.Context, result *messages.ExpressionMessage, producer brokers.Producer) {
	const fn = "agent.ConsumeMessageFromComputers"

	log := a.log.With(
		slog.String("fn", fn),
	)

	log.Info("agent consumes message from computers", slog.Any("message", result))

	result.AgentID = a.AgentID

	err := producer.PublishExpressionMessage(result)
	if err != nil {
		producer, err = producer.Reconnect()
		if err != nil {
			log.Error("agent error", sl.Err(err))
			a.kill()
			return
		}
		err = producer.PublishExpressionMessage(result)
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

// ConsumeMessageFromOrchestrator hanldes message from Consumer.
func (a *Agent) ConsumeMessageFromOrchestrator(ctx context.Context, msgFromOrchestrator amqp.Delivery) {
	const fn = "agent.ConsumeMessageFromOrchestrator"

	log := a.log.With(
		slog.String("fn", fn),
	)

	log.Info("agent consumes msg from orchestrator", slog.String("message", string(msgFromOrchestrator.Body)))

	var exprMsg messages.ExpressionMessage
	if err := json.Unmarshal(msgFromOrchestrator.Body, &exprMsg); err != nil {
		log.Error("agent error: failed to parse JSON", sl.Err(err))
		a.kill()
		return
	}

	err := msgFromOrchestrator.Ack(false)
	if err != nil {
		log.Error("agent error: error acknowledging message", sl.Err(err))
		a.kill()
		return
	}

	log.Info("token", slog.Any("tokens", exprMsg.Token))

	err = a.AssignToAgent(ctx, exprMsg.ExpressionID)
	if err != nil {
		log.Error("agent error", sl.Err(err))
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
