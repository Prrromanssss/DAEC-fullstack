package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/database"
	"github.com/Prrromanssss/DAEE-fullstack/pkg/orchestrator"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type AgentAgregator struct {
	amqpConfig   *config.AMQPConfig
	dbConfig     *config.DBConfig
	tasks        chan MessageFromOrchestrator
	amqpProducer *config.AMQPProducer
	amqpConsumer *config.AMQPConsumer
	mu           *sync.Mutex
}

type MessageFromOrchestrator struct {
	ExpressionID uuid.UUID `json:"expression_id"`
	Expression   string    `json:"expression"`
}

type ExpressionMessage struct {
	ExpressionID uuid.UUID `json:"expression_id"`
	Token        string    `json:"token"`
	Expression   string    `json:"expression"`
	Result       int       `json:"result"`
	IsPing       bool      `json:"is_ping"`
	AgentID      uuid.UUID `json:"agent_id"`
}

func NewAgentAgregator(
	rabbitMQURL string,
	dbCfg *config.DBConfig,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
) (*AgentAgregator, error) {
	amqpCfg, err := config.NewAMQPConfig(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPConfig for Agent Agregator: %v", err)
	}
	amqpProd, err := config.NewAMQPProducer(amqpCfg, titleForExpressionQueue)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPProducer for Agent Agregator: %v", err)
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, titleForResultAndPingQueue)
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

func (aa *AgentAgregator) PublishMessage(expressionID uuid.UUID, token, expresssion string) error {
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
	log.Printf("Publishing message to Queue: %s from Agent Agregator", aa.amqpProducer.Queue.Name)
	return nil
}

func (aa *AgentAgregator) HandlePing(agentID uuid.UUID) error {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	err := aa.dbConfig.DB.UpdateAgentLastPing(
		context.Background(),
		database.UpdateAgentLastPingParams{
			ID:       agentID,
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

	newExpr, newToken, err := orchestrator.InsertResultToToken(
		expression.ParseData,
		exprMsg.Token,
		exprMsg.Result,
	)
	if err != nil {
		return "", "", fmt.Errorf("can't insert tokens to expression: %v", err)
	}

	err = aa.dbConfig.DB.UpdateExpressionParseData(
		context.Background(),
		database.UpdateExpressionParseDataParams{
			ID:        exprMsg.ExpressionID,
			ParseData: newExpr,
		})

	if err != nil {
		return "", "", fmt.Errorf("can't update expression data: %v", err)
	}

	return newExpr, newToken, nil
}

func (aa *AgentAgregator) UpdateExpressionToReady(result int, exprID uuid.UUID) error {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	err := aa.dbConfig.DB.MakeExpressionReady(
		context.Background(),
		database.MakeExpressionReadyParams{
			ParseData: "",
			Result:    int32(result),
			UpdatedAt: time.Now().UTC(),
			ID:        exprID,
		})
	if err != nil {
		return fmt.Errorf("can't make expression ready: %v", err)
	}
	return nil
}

func (aa *AgentAgregator) ConsumeMessagesFromAgents(msgFromAgents amqp.Delivery) {
	log.Println("Agent agregator consume message from agent", msgFromAgents.Body)
	err := msgFromAgents.Ack(false)
	if err != nil {
		log.Fatalf("Error acknowledging message: %v", err)
	}
	var exprMsg ExpressionMessage
	if err := json.Unmarshal(msgFromAgents.Body, &exprMsg); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	if exprMsg.IsPing {
		err := aa.HandlePing(exprMsg.AgentID)
		if err != nil {
			log.Fatalf("Agent Agregator Error: %v", err)
		}
	} else {
		newExpr, newToken, err := aa.UpdateExpressionFromAgents(exprMsg)
		if err != nil {
			log.Fatalf("Agent Agregator Error: %v", err)
		}

		result, err := strconv.Atoi(newExpr)

		if err == nil &&
			orchestrator.IsNumber(newExpr) ||
			(newExpr[0] == '-' && orchestrator.IsNumber(newExpr[1:])) {
			err := aa.UpdateExpressionToReady(result, exprMsg.ExpressionID)
			if err != nil {
				log.Fatalf("Agent Agregator Error: %v", err)
			}
			return
		}

		if newToken != "" {
			err := aa.PublishMessage(exprMsg.ExpressionID, newToken, newExpr)
			if err != nil {
				log.Fatalf("Agent Agregator Error: %v", err)
			}
		}
	}
}

func (aa *AgentAgregator) ConsumeMessagesFromOrchestrator(expressionMessage MessageFromOrchestrator) {
	log.Println("Agent agregator consume message from orchestrator")
	tokens := orchestrator.GetTokens(expressionMessage.Expression)
	for _, token := range tokens {
		err := aa.PublishMessage(expressionMessage.ExpressionID, token, expressionMessage.Expression)
		if err != nil {
			log.Fatalf("Agent Agregator Error: %v", err)
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
