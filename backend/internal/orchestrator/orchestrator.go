package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/brokers"
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"

	"github.com/streadway/amqp"
)

type Orchestrator struct {
	log                  *slog.Logger
	dbConfig             *storage.Storage
	InactiveTimeForAgent int32
	mu                   *sync.Mutex
	kill                 context.CancelFunc
}

// NewOrchestrator creates new Orchestrator.
func NewOrchestrator(
	log *slog.Logger,
	dbCfg *storage.Storage,
	inactiveTimeForAgent int32,
	kill context.CancelFunc,
) (*Orchestrator, error) {

	return &Orchestrator{
		log:                  log,
		dbConfig:             dbCfg,
		InactiveTimeForAgent: inactiveTimeForAgent,
		mu:                   &sync.Mutex{},
		kill:                 kill,
	}, nil
}

// AddTask publish message to agents.
func (o *Orchestrator) AddTask(
	expressionMessage messages.ExpressionMessage,
	producer brokers.Producer,
) {
	const fn = "orchestrator.AddTask"

	o.log.Info("orchestrator ready to publish message to queue")

	tokens := parser.GetTokens(o.log, expressionMessage.Expression)
	for _, token := range tokens {
		err := producer.PublishExpressionMessage(&messages.ExpressionMessage{
			ExpressionID: expressionMessage.ExpressionID,
			Token:        token,
			Expression:   expressionMessage.Expression,
			UserID:       expressionMessage.UserID,
		})
		if err != nil {
			o.log.Error("can't publish token to queue", sl.Err(err), slog.String("fn", fn))
			// TODO: think about it. Should I kill orchestrator?
			o.kill()
		}
	}
}

// ReloadComputingExpressions add not completed expressions again to queue.
func (o *Orchestrator) ReloadComputingExpressions(
	ctx context.Context,
	producer brokers.Producer,
) error {
	const fn = "orchestrator.ReloadComputingExpressions"

	expressions, err := o.dbConfig.Queries.GetComputingExpressions(ctx)
	if err != nil {
		return fmt.Errorf("orhestrator Error: %v, fn: %s", err, fn)
	}

	for _, expr := range expressions {
		msgToQueue := messages.ExpressionMessage{
			ExpressionID: expr.ExpressionID,
			Expression:   expr.ParseData,
			UserID:       expr.UserID,
		}
		o.AddTask(msgToQueue, producer)
	}

	return nil
}

// CheckPing checks pings from agents and terminates them if there were no pings from them.
// Then if agent was terminated sends expressions by this agent again to queue
func (o *Orchestrator) CheckPing(ctx context.Context, producer brokers.Producer) error {
	const fn = "orchestrator.CheckPing"

	log := o.log.With(
		slog.String("fn", fn),
	)

	tx, err := o.dbConfig.DB.Begin()
	if err != nil {
		return err
	}

	qtx := o.dbConfig.Queries.WithTx(tx)

	agentIDs, err := qtx.TerminateAgents(
		ctx,
		strconv.Itoa(int(o.InactiveTimeForAgent)),
	)
	if err != nil {
		log.Error("can't make agents terminated", sl.Err(err))
		errRollback := tx.Rollback()
		if errRollback != nil {
			log.Error("can't rollback transaction")

			return errRollback
		}
		return err
	}

	if len(agentIDs) == 0 {
		log.Info("all agents are activate")
		err = tx.Commit()
		if err != nil {
			log.Error("can't commit transaction", sl.Err(err))

			return err
		}

		return nil
	}

	for _, agentID := range agentIDs {
		err := qtx.MakeExpressionsTerminated(ctx, sql.NullInt32{
			Int32: agentID,
			Valid: true,
		})
		if err != nil {
			log.Error("can't make expressions by this agent terminated",
				slog.Int("agent ID", int(agentID)), sl.Err(err))
			errRollback := tx.Rollback()
			if errRollback != nil {
				log.Error("can't rollback transaction")

				return errRollback
			}
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Error("can't commit transaction", sl.Err(err))

		return err
	}

	expressions, err := o.dbConfig.Queries.GetTerminatedExpressions(ctx)
	if err != nil {
		log.Error("can't get expressions", sl.Err(err))

		return err
	}

	for _, expr := range expressions {
		msgToQueue := messages.ExpressionMessage{
			ExpressionID: expr.ExpressionID,
			Expression:   expr.ParseData,
			UserID:       expr.UserID,
		}
		o.AddTask(msgToQueue, producer)
	}

	return nil
}

// HandlePing accepts ping from agent.
func (o *Orchestrator) HandlePing(ctx context.Context, agentID int32) error {
	const fn = "orchestrator.HandlePing"

	err := o.dbConfig.Queries.UpdateAgentLastPing(
		ctx,
		postgres.UpdateAgentLastPingParams{
			AgentID:  agentID,
			LastPing: time.Now().UTC(),
		})
	if err != nil {
		return fmt.Errorf("can't update last ping: %v, fn: %s", err, fn)
	}

	return nil
}

// HandleExpressionFromAgents makes expressions ready or publishes it again to queue.
func (o *Orchestrator) HandleExpression(
	ctx context.Context,
	exprMsg messages.ExpressionMessage,
	producer brokers.Producer,
) error {
	const fn = "orchestrator.HandleExpressionFromAgents"

	newResultAndToken, err := o.UpdateExpressionFromAgents(ctx, exprMsg)
	if err != nil {
		return fmt.Errorf("orchestrator error: %v, fn: %s", err, fn)
	}

	result, err := strconv.Atoi(newResultAndToken.Result)

	if err == nil &&
		parser.IsNumber(newResultAndToken.Result) ||
		(newResultAndToken.Result[0] == '-' && parser.IsNumber(newResultAndToken.Result[1:])) {
		err := o.UpdateExpressionToReady(ctx, result, exprMsg.ExpressionID)
		if err != nil {
			return fmt.Errorf("orchestrator error: %v, fn: %s", err, fn)
		}

		return nil
	}
	if newResultAndToken.Token != "" {
		err := producer.PublishExpressionMessage(&messages.ExpressionMessage{
			ExpressionID: exprMsg.ExpressionID,
			Token:        newResultAndToken.Token,
			Expression:   newResultAndToken.Result,
			UserID:       exprMsg.UserID,
		})
		if err != nil {
			return fmt.Errorf("orchestrator error: %v, fn: %s", err, fn)
		}
	}

	return nil
}

// UpdateExpressionFromAgents parses expression with new token and updates it in the database.
func (o *Orchestrator) UpdateExpressionFromAgents(
	ctx context.Context,
	exprMsg messages.ExpressionMessage,
) (messages.ResultAndTokenMessage, error) {
	const fn = "orchestrator.UpdateExpressionFromAgents"

	expression, err := o.dbConfig.Queries.GetExpressionByID(
		ctx,
		exprMsg.ExpressionID,
	)
	if err != nil {
		return messages.ResultAndTokenMessage{},
			fmt.Errorf("can't get expression by id: %v, fn: %s", err, fn)
	}

	resAndTokenMsg, err := parser.InsertResultToToken(
		expression.ParseData,
		exprMsg.Token,
		exprMsg.Result,
	)
	if err != nil {
		return messages.ResultAndTokenMessage{},
			fmt.Errorf("can't insert tokens to expression: %v, fn: %s", err, fn)
	}

	err = o.dbConfig.Queries.UpdateExpressionParseData(
		ctx,
		postgres.UpdateExpressionParseDataParams{
			ExpressionID: exprMsg.ExpressionID,
			ParseData:    resAndTokenMsg.Result,
		})
	if err != nil {
		return messages.ResultAndTokenMessage{},
			fmt.Errorf("can't update expression data: %v, fn: %s", err, fn)
	}

	return resAndTokenMsg, nil
}

// UpdateExpressionToReady updates expression to ready.
func (o *Orchestrator) UpdateExpressionToReady(
	ctx context.Context,
	result int,
	exprID int32,
) error {
	const fn = "orchestrator.UpdateExpressionToReady"

	err := o.dbConfig.Queries.MakeExpressionReady(
		ctx,
		postgres.MakeExpressionReadyParams{
			ParseData:    "",
			Result:       int32(result),
			UpdatedAt:    time.Now().UTC(),
			ExpressionID: exprID,
		})
	if err != nil {
		return fmt.Errorf("can't make expression ready: %v, fn: %s", err, fn)
	}

	return nil
}

// HandleMessagesFromAgents consumes message from agents.
// If it is ping handle it with HandlePing method.
// If it is expression handle it with HandleExpression method.
func (o *Orchestrator) HandleMessagesFromAgents(
	ctx context.Context,
	msgFromAgents amqp.Delivery,
	producer brokers.Producer,
) error {
	const fn = "orchestrator.ConsumeMessagesFromAgents"

	log := o.log.With(
		slog.String("fn", fn),
	)

	log.Info("orchestrator consumes message from agent", slog.String("msg", string(msgFromAgents.Body)))

	err := msgFromAgents.Ack(false)
	if err != nil {
		log.Error("error acknowledging message", sl.Err(err))
		return err
	}

	var exprMsg messages.ExpressionMessage
	if err := json.Unmarshal(msgFromAgents.Body, &exprMsg); err != nil {
		log.Error("failed to parse JSON", sl.Err(err))
		return err
	}

	if exprMsg.IsPing {
		err := o.HandlePing(ctx, exprMsg.AgentID)
		if err != nil {
			log.Error("orchestrator error", sl.Err(err))
			return err
		}
	} else {
		err := o.HandleExpression(ctx, exprMsg, producer)
		if err != nil {
			log.Error("", sl.Err(err))
			return err
		}
	}

	return nil
}
