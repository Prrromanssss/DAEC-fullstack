package agentagregator

import (
	"context"
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

type AgentAgregator struct {
	log      *slog.Logger
	dbConfig *storage.Storage
	Tasks    chan messages.MessageFromOrchestrator
	mu       *sync.Mutex
	kill     context.CancelFunc
}

// NewAgentAgregator creates new AgentAgregator.
func NewAgentAgregator(
	log *slog.Logger,
	dbCfg *storage.Storage,
	kill context.CancelFunc,
) (*AgentAgregator, error) {

	return &AgentAgregator{
		log:      log,
		dbConfig: dbCfg,
		Tasks:    make(chan messages.MessageFromOrchestrator),
		mu:       &sync.Mutex{},
		kill:     kill,
	}, nil
}

// AddTask adds task to AgentAgregator.
func (aa *AgentAgregator) AddTask(expressionMessage messages.MessageFromOrchestrator) {
	aa.Tasks <- expressionMessage
}

// ReloadComputingExpressions add not completed expressions to AgentAgregator.
func (aa *AgentAgregator) ReloadComputingExpressions(
	ctx context.Context,
) error {
	const fn = "agentagregator.ReloadComputingExpressions"

	expressions, err := aa.dbConfig.DB.GetComputingExpressions(ctx)
	if err != nil {
		return fmt.Errorf("orhestrator Error: %v, fn: %s", err, fn)
	}

	for _, expr := range expressions {
		msgToQueue := messages.MessageFromOrchestrator{
			ExpressionID: expr.ExpressionID,
			Expression:   expr.ParseData,
		}
		aa.AddTask(msgToQueue)
	}

	return nil
}

// HandlePing accepts ping from agent.
func (aa *AgentAgregator) HandlePing(ctx context.Context, agentID int32) error {
	const fn = "agentagregator.HandlePing"

	err := aa.dbConfig.DB.UpdateAgentLastPing(
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
func (aa *AgentAgregator) HandleExpression(
	ctx context.Context,
	exprMsg messages.ExpressionMessage,
	producer brokers.Producer,
) error {
	const fn = "agentagregator.HandleExpressionFromAgents"

	newResultAndToken, err := aa.UpdateExpressionFromAgents(ctx, exprMsg)
	if err != nil {
		return fmt.Errorf("agent agregator error: %v, fn: %s", err, fn)
	}

	result, err := strconv.Atoi(newResultAndToken.Result)

	if err == nil &&
		parser.IsNumber(newResultAndToken.Result) ||
		(newResultAndToken.Result[0] == '-' && parser.IsNumber(newResultAndToken.Result[1:])) {
		err := aa.UpdateExpressionToReady(ctx, result, exprMsg.ExpressionID)
		if err != nil {
			return fmt.Errorf("agent agregator error: %v, fn: %s", err, fn)
		}

		return nil
	}
	if newResultAndToken.Token != "" {
		err := producer.PublishExpressionMessage(&messages.ExpressionMessage{
			ExpressionID: exprMsg.ExpressionID,
			Token:        newResultAndToken.Token,
			Expression:   newResultAndToken.Result,
		})
		if err != nil {
			return fmt.Errorf("agent agregator error: %v, fn: %s", err, fn)
		}
	}

	return nil
}

// UpdateExpressionFromAgents consumes updated messages and update appropiate fields in database.
func (aa *AgentAgregator) UpdateExpressionFromAgents(
	ctx context.Context,
	exprMsg messages.ExpressionMessage,
) (messages.ResultAndTokenMessage, error) {
	const fn = "agentagregator.UpdateExpressionFromAgents"

	expression, err := aa.dbConfig.DB.GetExpressionByID(
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

	err = aa.dbConfig.DB.UpdateExpressionParseData(
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
func (aa *AgentAgregator) UpdateExpressionToReady(
	ctx context.Context,
	result int,
	exprID int32,
) error {
	const fn = "agentagregator.UpdateExpressionToReady"

	err := aa.dbConfig.DB.MakeExpressionReady(
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

// ConsumeMessagesFromAgents consumes message from agents.
// If it is ping handle it with HandlePing method.
// If it is expression handle it with HandleExpression method.
func (aa *AgentAgregator) ConsumeMessagesFromAgents(
	ctx context.Context,
	msgFromAgents amqp.Delivery,
	producer brokers.Producer,
) {
	const fn = "agentagregator.ConsumeMessagesFromAgents"

	log := aa.log.With(
		slog.String("fn", fn),
	)

	log.Info("agent agregator consumes message from agent", slog.String("msg", string(msgFromAgents.Body)))

	err := msgFromAgents.Ack(false)
	if err != nil {
		log.Error("error acknowledging message", sl.Err(err))
		aa.kill()
	}

	var exprMsg messages.ExpressionMessage
	if err := json.Unmarshal(msgFromAgents.Body, &exprMsg); err != nil {
		log.Error("failed to parse JSON", sl.Err(err))
		aa.kill()
	}

	if exprMsg.IsPing {
		err := aa.HandlePing(ctx, exprMsg.AgentID)
		if err != nil {
			log.Error("agent agregator error", sl.Err(err))
			aa.kill()
		}
	} else {
		err := aa.HandleExpression(ctx, exprMsg, producer)
		if err != nil {
			log.Error("", sl.Err(err))
			aa.kill()
		}
	}
}

// ConsumeMessagesFromOrchestrator consumes message from orchestrator,
// get tokens from this message,
// publishing it to queue.
func (aa *AgentAgregator) ConsumeMessagesFromOrchestrator(
	expressionMessage messages.MessageFromOrchestrator,
	producer brokers.Producer,
) {
	const fn = "agentagregator.ConsumeMessagesFromOrchestrator"

	aa.log.Info("agent agregator consumes message from orchestrator")

	tokens := parser.GetTokens(expressionMessage.Expression)
	for _, token := range tokens {
		err := producer.PublishExpressionMessage(&messages.ExpressionMessage{
			ExpressionID: expressionMessage.ExpressionID,
			Token:        token,
			Expression:   expressionMessage.Expression,
		})
		if err != nil {
			aa.log.Error("agent agregator error", sl.Err(err), slog.String("fn", fn))
			aa.kill()
		}
	}
}
