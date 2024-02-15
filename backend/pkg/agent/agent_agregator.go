package agent

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/pkg/orchestrator"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type AgentAgregator struct {
	amqpConfig   *config.AMQPConfig
	tasks        chan MessageFromOrchestrator
	amqpProducer *config.AMQPProducer
	amqpConsumer *config.AMQPConsumer
}

type MessageFromOrchestrator struct {
	ID         uuid.UUID `json:"id"`
	Expression string    `json:"data"`
}

type ExpressionMessage struct {
	ID          uuid.UUID `json:"id"`
	Token       string    `json:"data"`
	Result      int       `json:"result"`
	IsComputing bool      `json:"is_computing"`
}

func NewAgentAgregator(
	amqpCfg *config.AMQPConfig,
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
		tasks:        make(chan MessageFromOrchestrator),
		amqpProducer: amqpProd,
		amqpConsumer: amqpCons,
	}
}

func (aa *AgentAgregator) AddTask(expressionMessage MessageFromOrchestrator) {
	aa.tasks <- expressionMessage
}

func AgregateAgents(agentAg *AgentAgregator) {
	defer agentAg.amqpConfig.Conn.Close()
	defer agentAg.amqpConfig.Ch.Close()

	for {
		select {
		case expressionMessage := <-agentAg.tasks:
			readyExpressions := orchestrator.HandleParseExpression(expressionMessage.Expression)
			for _, token := range readyExpressions {
				msg := ExpressionMessage{
					ID:          expressionMessage.ID,
					Token:       token,
					IsComputing: false,
				}
				jsonData, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Failed to encode message to JSON: :%v", err)
					return
				}
				err = agentAg.amqpProducer.Ch.Publish(
					"",
					agentAg.amqpProducer.Queue.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        jsonData,
					},
				)
				if err != nil {
					log.Printf("Can't publish message to %s queue: %v", agentAg.amqpProducer.Queue.Name, err)
				}
			}
		case resOrPing := <-agentAg.amqpConsumer.Messages:
			fmt.Println(resOrPing)
		}
	}

}
