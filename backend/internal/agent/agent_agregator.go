package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
	"github.com/Prrromanssss/DAEE-fullstack/internal/rabbitmq"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage"
	"github.com/Prrromanssss/DAEE-fullstack/internal/storage/postgres"

	"github.com/streadway/amqp"
)

type AgentAgregator struct {
	log          *slog.Logger
	amqpConfig   *rabbitmq.AMQPConfig
	dbConfig     *storage.Storage
	tasks        chan MessageFromOrchestrator
	amqpProducer *rabbitmq.AMQPProducer
	amqpConsumer *rabbitmq.AMQPConsumer
	mu           *sync.Mutex
}

type MessageFromOrchestrator struct {
	ExpressionID int32  `json:"expression_id"`
	Expression   string `json:"expression"`
}

type ExpressionMessage struct {
	ExpressionID int32  `json:"expression_id"`
	Token        string `json:"token"`
	Expression   string `json:"expression"`
	Result       int    `json:"result"`
	IsPing       bool   `json:"is_ping"`
	AgentID      int32  `json:"agent_id"`
}

func NewAgentAgregator(
	log *slog.Logger,
	rabbitMQURL string,
	dbCfg *storage.Storage,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
) (*AgentAgregator, error) {
	amqpCfg, err := rabbitmq.NewAMQPConfig(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPConfig for Agent Agregator: %v", err)
	}
	amqpProd, err := rabbitmq.NewAMQPProducer(amqpCfg, titleForExpressionQueue)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPProducer for Agent Agregator: %v", err)
	}
	amqpCons, err := rabbitmq.NewAMQPConsumer(amqpCfg, titleForResultAndPingQueue)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPConsumer for Agent Agregator: %v", err)
	}

	return &AgentAgregator{
		amqpConfig:   amqpCfg,
		dbConfig:     dbCfg,
		tasks:        make(chan MessageFromOrchestrator),
		amqpProducer: amqpProd,
		amqpConsumer: amqpCons,
		mu:           &sync.Mutex{},
	}, nil
}

func (aa *AgentAgregator) AddTask(expressionMessage MessageFromOrchestrator) {
	aa.tasks <- expressionMessage
}

func (aa *AgentAgregator) PublishMessage(expressionID int32, token, expresssion string) error {
	msg := ExpressionMessage{
		ExpressionID: expressionID,
		Token:        token,
		Expression:   expresssion,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to encode message to JSON: %v", err)
	}

	err = aa.amqpProducer.ChannelForProduce.Publish(
		"",
		aa.amqpProducer.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)
	if err != nil {
		return fmt.Errorf("can't publish message to %s queue: %v", aa.amqpProducer.Queue.Name, err)
	}
	aa.log.Info("publishing message to Queue from Agent Agregator", slog.String("queue", aa.amqpProducer.Queue.Name))
	return nil
}

func (aa *AgentAgregator) HandlePing(agentID int32) error {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	err := aa.dbConfig.DB.UpdateAgentLastPing(
		context.Background(),
		postgres.UpdateAgentLastPingParams{
			AgentID:  agentID,
			LastPing: time.Now().UTC(),
		})
	if err != nil {
		return fmt.Errorf("can't update last ping: %v", err)
	}
	return nil
}

func (aa *AgentAgregator) UpdateExpressionFromAgents(exprMsg ExpressionMessage) (string, string, error) {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	expression, err := aa.dbConfig.DB.GetExpressionByID(
		context.Background(),
		exprMsg.ExpressionID,
	)
	if err != nil {
		return "", "", fmt.Errorf("can't get expression by id: %v", err)
	}

	newExpr, newToken, err := parser.InsertResultToToken(
		expression.ParseData,
		exprMsg.Token,
		exprMsg.Result,
	)
	if err != nil {
		return "", "", fmt.Errorf("can't insert tokens to expression: %v", err)
	}

	err = aa.dbConfig.DB.UpdateExpressionParseData(
		context.Background(),
		postgres.UpdateExpressionParseDataParams{
			ExpressionID: exprMsg.ExpressionID,
			ParseData:    newExpr,
		})

	if err != nil {
		return "", "", fmt.Errorf("can't update expression data: %v", err)
	}

	return newExpr, newToken, nil
}

func (aa *AgentAgregator) UpdateExpressionToReady(result int, exprID int32) error {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	err := aa.dbConfig.DB.MakeExpressionReady(
		context.Background(),
		postgres.MakeExpressionReadyParams{
			ParseData:    "",
			Result:       int32(result),
			UpdatedAt:    time.Now().UTC(),
			ExpressionID: exprID,
		})
	if err != nil {
		return fmt.Errorf("can't make expression ready: %v", err)
	}
	return nil
}

func (aa *AgentAgregator) ConsumeMessagesFromAgents(msgFromAgents amqp.Delivery) {
	aa.log.Info("agent agregator consume message from agent", slog.String("msg", string(msgFromAgents.Body)))
	err := msgFromAgents.Ack(false)
	if err != nil {
		aa.log.Error("error acknowledging message", sl.Err(err))
		os.Exit(1)
	}
	var exprMsg ExpressionMessage
	if err := json.Unmarshal(msgFromAgents.Body, &exprMsg); err != nil {
		aa.log.Error("failed to parse JSON", sl.Err(err))
		os.Exit(1)
	}

	if exprMsg.IsPing {
		err := aa.HandlePing(exprMsg.AgentID)
		if err != nil {
			aa.log.Error("agent agregator error", sl.Err(err))
			os.Exit(1)
		}
	} else {
		newExpr, newToken, err := aa.UpdateExpressionFromAgents(exprMsg)
		if err != nil {
			aa.log.Error("agent agregator error", sl.Err(err))
			os.Exit(1)
		}

		result, err := strconv.Atoi(newExpr)

		if err == nil &&
			parser.IsNumber(newExpr) ||
			(newExpr[0] == '-' && parser.IsNumber(newExpr[1:])) {
			err := aa.UpdateExpressionToReady(result, exprMsg.ExpressionID)
			if err != nil {
				aa.log.Error("agent agregator error:", sl.Err(err))
				os.Exit(1)
			}
			return
		}

		if newToken != "" {
			err := aa.PublishMessage(exprMsg.ExpressionID, newToken, newExpr)
			if err != nil {
				aa.log.Error("agent agregator error", sl.Err(err))
				os.Exit(1)
			}
		}
	}
}

func (aa *AgentAgregator) ConsumeMessagesFromOrchestrator(expressionMessage MessageFromOrchestrator) {
	aa.log.Info("agent agregator consume message from orchestrator")
	tokens := parser.GetTokens(expressionMessage.Expression)
	for _, token := range tokens {
		err := aa.PublishMessage(expressionMessage.ExpressionID, token, expressionMessage.Expression)
		if err != nil {
			aa.log.Error("agent agregator error", sl.Err(err))
			os.Exit(1)
		}
	}
}

func AgregateAgents(agentAg *AgentAgregator) {
	defer agentAg.amqpConfig.Conn.Close()
	defer agentAg.amqpConfig.ChannelForConsume.Close()
	defer agentAg.amqpConfig.ChannelForProduce.Close()

	go func() {
		for msgFromAgents := range agentAg.amqpConsumer.Messages {
			go agentAg.ConsumeMessagesFromAgents(msgFromAgents)
		}
	}()

	for {
		expressionMessage := <-agentAg.tasks
		go agentAg.ConsumeMessagesFromOrchestrator(expressionMessage)
	}
}
