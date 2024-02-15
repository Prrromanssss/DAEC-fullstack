package agent

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/internal/database"
	"Prrromanssss/DAEE/pkg/orchestrator"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type AgentAgregator struct {
	amqpConfig   *config.AMQPConfig
	dbConfig     *config.DBConfig
	tasks        chan MessageFromOrchestrator
	amqpProducer *config.AMQPProducer
	amqpConsumer *config.AMQPConsumer
}

type MessageFromOrchestrator struct {
	ExpressionID uuid.UUID `json:"expression_id"`
	Expression   string    `json:"data"`
}

type ExpressionMessage struct {
	ExpressionID uuid.UUID `json:"expression_id"`
	Token        string    `json:"data"`
	Result       int       `json:"result"`
	IsComputing  bool      `json:"is_computing"`
	IsPing       bool      `json:"is_ping"`
	AgentID      uuid.UUID `json:"agent_id"`
}

func NewAgentAgregator(
	amqpCfg *config.AMQPConfig,
	dbCfg *config.DBConfig,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
) *AgentAgregator {
	amqpProd, err := config.NewAMQProducer(amqpCfg, titleForExpressionQueue)
	if err != nil {
		log.Fatalf("Cant't create NewAMQPProducer: %v", err)
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, titleForResultAndPingQueue)
	if err != nil {
		log.Fatalf("Cant't create NewAMQPConsumer: %v", err)
	}

	return &AgentAgregator{
		amqpConfig:   amqpCfg,
		dbConfig:     dbCfg,
		tasks:        make(chan MessageFromOrchestrator),
		amqpProducer: amqpProd,
		amqpConsumer: amqpCons,
	}
}

func (aa *AgentAgregator) AddTask(expressionMessage MessageFromOrchestrator) {
	aa.tasks <- expressionMessage
}

func (aa *AgentAgregator) PublishMessage(expressionID uuid.UUID, token string) {
	msg := ExpressionMessage{
		ExpressionID: expressionID,
		Token:        token,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to encode message to JSON: :%v", err)
		return
	}
	err = aa.amqpProducer.Ch.Publish(
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
		log.Printf("Can't publish message to %s queue: %v", aa.amqpProducer.Queue.Name, err)
	}
}

func AgregateAgents(agentAg *AgentAgregator) {
	defer agentAg.amqpConfig.Conn.Close()
	defer agentAg.amqpConfig.Ch.Close()

	for {
		select {
		case expressionMessage := <-agentAg.tasks:
			tokens := orchestrator.GetTokens(expressionMessage.Expression)
			for _, token := range tokens {
				agentAg.PublishMessage(expressionMessage.ExpressionID, token)
			}
		case msgResOrPing := <-agentAg.amqpConsumer.Messages:
			var exprMsg ExpressionMessage
			if err := json.Unmarshal(msgResOrPing.Body, &exprMsg); err != nil {
				log.Printf("Failed to parse JSON: %v", err)
				continue
			}
			if exprMsg.IsPing {
				_, err := agentAg.dbConfig.DB.UpdateAgentLastPing(context.Background(), database.UpdateAgentLastPingParams{
					ID:       exprMsg.AgentID,
					LastPing: time.Now().UTC(),
				})
				if err != nil {
					log.Printf("Can't update last ping: %v", err)
				}
			} else {
				parseExpr, err := agentAg.dbConfig.DB.GetExpressionByID(context.Background(), exprMsg.ExpressionID)
				if err != nil {
					log.Printf("Can't get expression by id: %v", err)
				}
				newExpr, newToken, err := orchestrator.InsertResultToToken(parseExpr.Data, exprMsg.Token, exprMsg.Result)
				if err != nil {
					log.Printf("Can't get expression by id: %v", err)
				}

				if newToken != "" {
					agentAg.PublishMessage(exprMsg.ExpressionID, newToken)
				}
				_, err = agentAg.dbConfig.DB.UpdateExpressionData(context.Background(), database.UpdateExpressionDataParams{
					ID:   exprMsg.ExpressionID,
					Data: newExpr,
				})
				if err != nil {
					log.Printf("Can't update expression data: %v", err)
				}
			}
		}
	}

}
