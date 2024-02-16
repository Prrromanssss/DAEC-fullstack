package agent

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/internal/database"
	"Prrromanssss/DAEE/pkg/orchestrator"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type AgentAgregator struct {
	rabbitMQURL                string
	titleForExpressionQueue    string
	titleForResultAndPingQueue string
	amqpConfig                 *config.AMQPConfig
	dbConfig                   *config.DBConfig
	tasks                      chan MessageFromOrchestrator
	amqpProducer               *config.AMQPProducer
	amqpConsumer               *config.AMQPConsumer
	mu                         *sync.Mutex
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
	rabbitMQURL string,
	dbCfg *config.DBConfig,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
) *AgentAgregator {
	amqpCfg, err := config.NewAMQPConfig(rabbitMQURL)
	if err != nil {
		log.Fatalf("Can't create NewAMQPConfig: %v", err)
	}
	amqpProd, err := config.NewAMQPProducer(amqpCfg, titleForExpressionQueue)
	if err != nil {
		log.Fatalf("Can't create NewAMQPProducer: %v", err)
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, titleForResultAndPingQueue)
	if err != nil {
		log.Fatalf("Can't create NewAMQPConsumer: %v", err)
	}

	return &AgentAgregator{
		rabbitMQURL:                rabbitMQURL,
		titleForExpressionQueue:    titleForExpressionQueue,
		titleForResultAndPingQueue: titleForResultAndPingQueue,
		amqpConfig:                 amqpCfg,
		dbConfig:                   dbCfg,
		tasks:                      make(chan MessageFromOrchestrator),
		amqpProducer:               amqpProd,
		amqpConsumer:               amqpCons,
		mu:                         &sync.Mutex{},
	}
}

func (aa *AgentAgregator) Reconnect() {
	aa.amqpConfig.ChannelForConsume.Close()
	aa.amqpConfig.ChannelForProduce.Close()
	aa.amqpConfig.Conn.Close()

	amqpCfg, err := config.NewAMQPConfig(aa.rabbitMQURL)
	if err != nil {
		log.Fatalf("Can't create NewAMQPConfig: %v", err)
	}
	amqpProd, err := config.NewAMQPProducer(amqpCfg, aa.titleForExpressionQueue)
	if err != nil {
		log.Fatalf("Can't create NewAMQPProducer: %v", err)
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, aa.titleForResultAndPingQueue)
	if err != nil {
		log.Fatalf("Can't create NewAMQPConsumer: %v", err)
	}
	aa.amqpConfig = amqpCfg
	aa.amqpConsumer = amqpCons
	aa.amqpProducer = amqpProd
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
		log.Fatalf("Can't publish message to %s queue: %v", aa.amqpProducer.Queue.Name, err)
		return
	}
	log.Printf("Publishing message to Queue: %s", aa.amqpProducer.Queue.Name)
}

func AgregateAgents(agentAg *AgentAgregator) {
	defer agentAg.amqpConfig.Conn.Close()
	defer agentAg.amqpConfig.ChannelForConsume.Close()
	defer agentAg.amqpConfig.ChannelForProduce.Close()

	go func() {
		for msgFromAgents := range agentAg.amqpConsumer.Messages {
			go func(msgFromAgents amqp.Delivery) {
				log.Println("Agent agregator consume message from agent", msgFromAgents.Body)
				err := msgFromAgents.Ack(false)
				if err != nil {
					log.Fatalf("Error acknowledging message: %v", err)
				}
				var exprMsg ExpressionMessage
				if err := json.Unmarshal(msgFromAgents.Body, &exprMsg); err != nil {
					log.Fatalf("Failed to parse JSON: %v", err)
					return
				}
				if exprMsg.IsPing {
					err := agentAg.dbConfig.DB.UpdateAgentLastPing(
						context.Background(),
						database.UpdateAgentLastPingParams{
							ID:       exprMsg.AgentID,
							LastPing: time.Now().UTC(),
						})
					if err != nil {
						log.Fatalf("Can't update last ping: %v", err)
					}
				} else {
					agentAg.mu.Lock()
					expression, err := agentAg.dbConfig.DB.GetExpressionByID(context.Background(), exprMsg.ExpressionID)
					if err != nil {
						agentAg.mu.Unlock()
						log.Printf("Can't get expression by id: %v", err)
						return
					}

					newExpr, newToken, err := orchestrator.InsertResultToToken(
						expression.ParseData,
						exprMsg.Token,
						exprMsg.Result,
					)
					if err != nil {
						log.Fatalf("Can't insert tokens to expression: %v", err)
					}

					err = agentAg.dbConfig.DB.UpdateExpressionParseData(
						context.Background(),
						database.UpdateExpressionParseDataParams{
							ID:        exprMsg.ExpressionID,
							ParseData: newExpr,
						})

					if err != nil {
						log.Fatalf("Can't update expression data: %v", err)
					}
					agentAg.mu.Unlock()

					result, err := strconv.Atoi(newExpr)
					if err == nil &&
						orchestrator.IsNumber(newExpr) ||
						(newExpr[0] == '-' && orchestrator.IsNumber(newExpr[1:])) {
						err := agentAg.dbConfig.DB.MakeExpressionReady(
							context.Background(),
							database.MakeExpressionReadyParams{
								ParseData: "",
								Result:    int32(result),
								UpdatedAt: time.Now().UTC(),
								ID:        exprMsg.ExpressionID,
							})
						if err != nil {
							log.Fatalf("Can't make expression ready: %v", err)
						}
						return
					}

					if newToken != "" {
						agentAg.PublishMessage(exprMsg.ExpressionID, newToken, newExpr)
					}

				}
			}(msgFromAgents)
		}
	}()

	for {
		expressionMessage := <-agentAg.tasks
		go func(expressionMessage MessageFromOrchestrator) {
			log.Println("Agent agregator consume message from orchestrator")
			tokens := orchestrator.GetTokens(expressionMessage.Expression)
			for _, token := range tokens {
				agentAg.PublishMessage(expressionMessage.ExpressionID, token, expressionMessage.Expression)
			}
		}(expressionMessage)
	}
}
