package agent

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/internal/database"
	"Prrromanssss/DAEE/pkg/orchestrator"
	"context"
	"encoding/json"
	"log"
	"strconv"
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
	Expression   string    `json:"expression"`
}

type ExpressionMessage struct {
	ExpressionID uuid.UUID `json:"expression_id"`
	Token        string    `json:"token"`
	Expression   string    `json:"expression"`
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

func (aa *AgentAgregator) PublishMessage(expressionID uuid.UUID, token, expresssion string) {
	msg := ExpressionMessage{
		ExpressionID: expressionID,
		Token:        token,
		Expression:   expresssion,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("Failed to encode message to JSON: %v", err)
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
	log.Println("Publishing message to Queue")
	if err != nil {
		log.Fatalf("Can't publish message to %s queue: %v", aa.amqpProducer.Queue.Name, err)
		return
	}
}

func AgregateAgents(agentAg *AgentAgregator) {
	defer agentAg.amqpConfig.Conn.Close()
	defer agentAg.amqpConfig.Ch.Close()
	for {
		select {
		case expressionMessage := <-agentAg.tasks:
			log.Println("Consume message from orchestrator")
			tokens := orchestrator.GetTokens(expressionMessage.Expression)
			for _, token := range tokens {
				log.Println("Consume token to queue: ", agentAg.amqpProducer.Queue.Name)
				agentAg.PublishMessage(expressionMessage.ExpressionID, token, expressionMessage.Expression)
			}
		case msgFromAgents := <-agentAg.amqpConsumer.Messages:
			err := msgFromAgents.Ack(false)
			if err != nil {
				log.Fatalf("Error acknowledging message: %v", err)
			}
			var exprMsg ExpressionMessage
			if err := json.Unmarshal(msgFromAgents.Body, &exprMsg); err != nil {
				log.Fatalf("Failed to parse JSON: %v", err)
				continue
			}
			if exprMsg.IsPing {
				err := agentAg.dbConfig.DB.UpdateAgentLastPing(context.Background(), database.UpdateAgentLastPingParams{
					ID:       exprMsg.AgentID,
					LastPing: time.Now().UTC(),
				})
				if err != nil {
					log.Fatalf("Can't update last ping: %v", err)
				}
			} else {
				newExpr, newToken, err := orchestrator.InsertResultToToken(exprMsg.Expression, exprMsg.Token, exprMsg.Result)
				if err != nil {
					log.Fatalf("Can't get expression by id: %v", err)
				}
				result, err := strconv.Atoi(newExpr)
				if err != nil {
					log.Fatalf("Can't convert result to int: %v", err)
				}

				if orchestrator.IsNumber(newExpr) || (newExpr[0] == '-' && orchestrator.IsNumber(newExpr[1:])) {
					err := agentAg.dbConfig.DB.MakeExpressionReady(context.Background(), database.MakeExpressionReadyParams{
						ParseData: "",
						Result:    int32(result),
						UpdatedAt: time.Now().UTC(),
						ID:        exprMsg.ExpressionID,
					})
					if err != nil {
						log.Fatalf("Can't make expression ready: %v", err)
					}
					continue
				}

				if newToken != "" {
					agentAg.PublishMessage(exprMsg.ExpressionID, newToken, newExpr)
				}

				_, err = agentAg.dbConfig.DB.UpdateExpressionData(context.Background(), database.UpdateExpressionDataParams{
					ID:   exprMsg.ExpressionID,
					Data: newExpr,
				})
				if err != nil {
					log.Fatalf("Can't update expression data: %v", err)
				}
			}
		}
	}

}
