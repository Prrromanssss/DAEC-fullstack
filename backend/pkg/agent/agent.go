package agent

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/internal/database"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type Agent struct {
	agentID                         uuid.UUID
	number_of_parallel_calculations int
	number_of_active_calculations   int
	last_ping                       time.Time
	status                          string
	amqpConfig                      *config.AMQPConfig
	dbConfig                        *config.DBConfig
	simpleComputers                 chan *ExpressionMessage
	amqpProducer                    *config.AMQPProducer
	amqpConsumer                    *config.AMQPConsumer
}

func NewAgent(
	amqpCfg *config.AMQPConfig,
	dbCfg *config.DBConfig,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
	numberOfParallelCalculations int32,
) (*Agent, error) {
	amqpProd, err := config.NewAMQProducer(amqpCfg, titleForExpressionQueue)
	if err != nil {
		log.Printf("Cant't create NewAMQPProducer: %v", err)
		return nil, err
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, titleForResultAndPingQueue)
	if err != nil {
		log.Printf("Cant't create NewAMQPConsumer: %v", err)
		return nil, err
	}
	agent, err := dbCfg.DB.CreateAgent(context.Background(), database.CreateAgentParams{
		ID:                           uuid.New(),
		CreatedAt:                    time.Now().UTC(),
		NumberOfParallelCalculations: numberOfParallelCalculations,
		LastPing:                     time.Now().UTC(),
		Status:                       "waiting",
	})
	if err != nil {
		log.Printf("Cant't create Agent: %v", err)
		return nil, err
	}
	return &Agent{
		agentID:                         agent.ID,
		number_of_parallel_calculations: int(agent.NumberOfParallelCalculations),
		number_of_active_calculations:   0,
		last_ping:                       agent.LastPing,
		status:                          string(agent.Status),
		amqpConfig:                      amqpCfg,
		dbConfig:                        dbCfg,
		simpleComputers:                 make(chan *ExpressionMessage),
		amqpProducer:                    amqpProd,
		amqpConsumer:                    amqpCons,
	}, nil

}

func (a *Agent) PublishMessage(msg *ExpressionMessage) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to encode message to JSON: :%v", err)
		return
	}
	err = a.amqpProducer.Ch.Publish(
		"",
		a.amqpProducer.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)
	log.Println("Publishing message to Queue")
	if err != nil {
		log.Printf("Can't publish message to %s queue: %v", a.amqpProducer.Queue.Name, err)
	}
}

func AgentService(agent *Agent) {
	for {
		select {
		case result := <-agent.simpleComputers:
			agent.PublishMessage(result)
			agent.number_of_active_calculations--
			if agent.number_of_active_calculations == 0 {
				agent.status = "waiting"
				err := agent.dbConfig.DB.UpdateAgentStatus(context.Background(), database.UpdateAgentStatusParams{
					Status: "waiting",
					ID:     agent.agentID,
				})
				if err != nil {
					log.Printf("Can't update agent status: %v", err)
				}
			}
		case msgFromAgentAgregator := <-agent.amqpConsumer.Messages:
			var exprMsg ExpressionMessage
			if err := json.Unmarshal(msgFromAgentAgregator.Body, &exprMsg); err != nil {
				log.Printf("Failed to parse JSON: %v", err)
				continue
			}
			if agent.number_of_active_calculations >= agent.number_of_parallel_calculations {
				continue
			}

			err := msgFromAgentAgregator.Ack(false)
			if err != nil {
				log.Printf("Error acknowledging message: %v", err)
			}

			tokenSplit := strings.Split(exprMsg.Token, " ")
			if len(tokenSplit) != 3 {
				log.Println("Invalid token")
			}
			oper := tokenSplit[2]
			if !(oper == "+" || oper == "-" || oper == "/" || oper == "*") {
				log.Println("Operation in token doesn't match any of these +, -, /, *")
			}

			digit1, err := strconv.Atoi(tokenSplit[0])
			if err != nil {
				log.Printf("Can't convert int to str: %v", err)
			}
			digit2, err := strconv.Atoi(tokenSplit[1])
			if err != nil {
				log.Printf("Can't convert int to str: %v", err)
			}

			time_for_oper, err := agent.dbConfig.DB.GetOperationTimeByType(context.Background(), oper)
			if err != nil {
				log.Printf("Can't get execution time by operation type: %v", err)
			}

			timer := time.NewTimer(time.Duration(time_for_oper) * time.Second)

			go simpleComputer(&exprMsg, digit1, digit2, oper, timer, agent.simpleComputers)

			agent.number_of_active_calculations++
			if agent.number_of_active_calculations == agent.number_of_parallel_calculations {
				agent.status = "sleeping"
				err := agent.dbConfig.DB.UpdateAgentStatus(context.Background(), database.UpdateAgentStatusParams{
					Status: "sleeping",
					ID:     agent.agentID,
				})
				if err != nil {
					log.Printf("Can't update agent status: %v", err)
				}
			} else if agent.status != "running" {
				agent.status = "running"
				err := agent.dbConfig.DB.UpdateAgentStatus(context.Background(), database.UpdateAgentStatusParams{
					Status: "running",
					ID:     agent.agentID,
				})
				if err != nil {
					log.Printf("Can't update agent status: %v", err)
				}
			}

		}
	}
}
