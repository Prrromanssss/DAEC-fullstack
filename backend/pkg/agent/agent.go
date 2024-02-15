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
	kill                            chan struct{}
}

func NewAgent(
	amqpCfg *config.AMQPConfig,
	dbCfg *config.DBConfig,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
	numberOfParallelCalculations int32,
) (*Agent, error) {
	amqpProd, err := config.NewAMQProducer(amqpCfg, titleForResultAndPingQueue)
	if err != nil {
		log.Printf("Cant't create NewAMQPProducer: %v", err)
		return nil, err
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, titleForExpressionQueue)
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
		kill:                            make(chan struct{}),
	}, nil

}

func (a *Agent) PublishMessage(msg *ExpressionMessage) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to encode message to JSON: :%v", err)
		a.kill <- struct{}{}
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
		a.kill <- struct{}{}
		return
	}
}

func AgentService(agent *Agent) {
	for {
		select {
		case <-agent.kill:
			log.Printf("Agent %s is down", agent.agentID)
			agent.status = "terminated"
			err := agent.dbConfig.DB.UpdateAgentStatus(context.Background(), database.UpdateAgentStatusParams{
				ID:     agent.agentID,
				Status: "terminated",
			})
			if err != nil {
				log.Printf("Can't update status of agent to terminate: %v", err)
			}
			return
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
					agent.kill <- struct{}{}
					continue
				}
			}
		case msgFromAgentAgregator := <-agent.amqpConsumer.Messages:
			var exprMsg ExpressionMessage
			if err := json.Unmarshal(msgFromAgentAgregator.Body, &exprMsg); err != nil {
				log.Printf("Failed to parse JSON: %v", err)
				agent.kill <- struct{}{}
				continue
			}
			if agent.number_of_active_calculations >= agent.number_of_parallel_calculations {
				continue
			}

			err := msgFromAgentAgregator.Ack(false)
			if err != nil {
				log.Printf("Error acknowledging message: %v", err)
				agent.kill <- struct{}{}
				continue
			}

			tokenSplit := strings.Split(exprMsg.Token, " ")
			if len(tokenSplit) != 3 {
				log.Println("Invalid token")
				agent.kill <- struct{}{}
				continue
			}
			oper := tokenSplit[2]
			if !(oper == "+" || oper == "-" || oper == "/" || oper == "*") {
				log.Println("Operation in token doesn't match any of these +, -, /, *")
				agent.kill <- struct{}{}
				continue
			}

			digit1, err := strconv.Atoi(tokenSplit[0])
			if err != nil {
				log.Printf("Can't convert int to str: %v", err)
				agent.kill <- struct{}{}
				continue
			}
			digit2, err := strconv.Atoi(tokenSplit[1])
			if err != nil {
				log.Printf("Can't convert int to str: %v", err)
				agent.kill <- struct{}{}
				continue
			}

			time_for_oper, err := agent.dbConfig.DB.GetOperationTimeByType(context.Background(), oper)
			if err != nil {
				log.Printf("Can't get execution time by operation type: %v", err)
				agent.kill <- struct{}{}
				continue
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
					agent.kill <- struct{}{}
					continue
				}
			} else if agent.status != "running" {
				agent.status = "running"
				err := agent.dbConfig.DB.UpdateAgentStatus(context.Background(), database.UpdateAgentStatusParams{
					Status: "running",
					ID:     agent.agentID,
				})
				if err != nil {
					log.Printf("Can't update agent status: %v", err)
					agent.kill <- struct{}{}
					continue
				}
			}

		}
	}
}
