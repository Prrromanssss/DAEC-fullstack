package agent

import (
	"Prrromanssss/DAEE/config"
	"Prrromanssss/DAEE/internal/database"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type Agent struct {
	rabbitMQURL                     string
	titleForExpressionQueue         string
	titleForResultAndPingQueue      string
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
	rabbitMQURL string,
	dbCfg *config.DBConfig,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
	numberOfParallelCalculations int32,
) (*Agent, error) {
	amqpCfg, err := config.NewAMQPConfig(rabbitMQURL)
	if err != nil {
		log.Fatalf("Can't create NewAMQPConfig: %v", err)
	}
	amqpProd, err := config.NewAMQPProducer(amqpCfg, titleForResultAndPingQueue)
	if err != nil {
		log.Printf("Can't create NewAMQPProducer: %v", err)
		return nil, err
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, titleForExpressionQueue)
	if err != nil {
		log.Printf("Can't create NewAMQPConsumer: %v", err)
		return nil, err
	}
	agent, err := dbCfg.DB.CreateAgent(
		context.Background(),
		database.CreateAgentParams{
			ID:                           uuid.New(),
			CreatedAt:                    time.Now().UTC(),
			NumberOfParallelCalculations: numberOfParallelCalculations,
			LastPing:                     time.Now().UTC(),
			Status:                       "waiting",
		})
	if err != nil {
		log.Printf("Can't create Agent: %v", err)
		return nil, err
	}
	return &Agent{
		rabbitMQURL:                     rabbitMQURL,
		titleForExpressionQueue:         titleForExpressionQueue,
		titleForResultAndPingQueue:      titleForResultAndPingQueue,
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

func (a *Agent) Reconnect() {
	a.amqpConfig.ChannelForConsume.Close()
	a.amqpConfig.ChannelForProduce.Close()
	a.amqpConfig.Conn.Close()

	amqpCfg, err := config.NewAMQPConfig(a.rabbitMQURL)
	if err != nil {
		log.Fatalf("Can't create NewAMQPConfig: %v", err)
	}
	amqpProd, err := config.NewAMQPProducer(amqpCfg, a.titleForResultAndPingQueue)
	if err != nil {
		log.Fatalf("Can't create NewAMQPProducer: %v", err)
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, a.titleForExpressionQueue)
	if err != nil {
		log.Fatalf("Can't create NewAMQPConsumer: %v", err)
	}
	a.amqpConfig = amqpCfg
	a.amqpConsumer = amqpCons
	a.amqpProducer = amqpProd
}

func (a *Agent) PublishMessage(msg *ExpressionMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to encode message to JSON: :%v", err)
		return errors.New("failed to encode message to JSON")
	}

	// a.Reconnect()

	err = a.amqpProducer.ChannelForProduce.Publish(
		"",
		a.amqpProducer.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)

	if err != nil {
		log.Printf("Can't publish message to %s queue: %v", a.amqpProducer.Queue.Name, err)
		return errors.New("can't publish message to queue")
	}
	log.Printf("Publishing message to Queue: %s", a.amqpProducer.Queue.Name)
	return nil
}

func AgentService(agent *Agent) {
	defer agent.amqpConfig.Conn.Close()
	defer agent.amqpConfig.ChannelForConsume.Close()
	defer agent.amqpConfig.ChannelForProduce.Close()

	go func() {
		for msgFromAgentAgregator := range agent.amqpConsumer.Messages {
			go func(msgFromAgentAgregator amqp.Delivery) {
				log.Println("Agent consume msg from agent agregator", msgFromAgentAgregator.Body)
				var exprMsg ExpressionMessage
				if err := json.Unmarshal(msgFromAgentAgregator.Body, &exprMsg); err != nil {
					log.Printf("Failed to parse JSON: %v", err)
					return
				}

				if agent.number_of_active_calculations >= agent.number_of_parallel_calculations {
					return
				}

				err := msgFromAgentAgregator.Ack(false)

				if err != nil {
					log.Printf("Error acknowledging message: %v", err)
					return
				}

				err = agent.dbConfig.DB.UpdateExpressionStatus(
					context.Background(),
					database.UpdateExpressionStatusParams{
						ID:     exprMsg.ExpressionID,
						Status: "computing",
					})

				if err != nil {
					log.Printf("Can't update expression status: %v", err)
					return
				}

				tokenSplit := strings.Split(exprMsg.Token, " ")
				if len(tokenSplit) != 3 {
					log.Println("Invalid token")
					return
				}
				oper := tokenSplit[2]
				if !(oper == "+" || oper == "-" || oper == "/" || oper == "*") {
					log.Println("Operation in token doesn't match any of these +, -, /, *")
					return
				}

				digit1, err := strconv.Atoi(tokenSplit[0])
				if err != nil {
					log.Printf("Can't convert int to str: %v", err)
					return
				}
				digit2, err := strconv.Atoi(tokenSplit[1])
				if err != nil {
					log.Printf("Can't convert int to str: %v", err)
					return
				}

				time_for_oper, err := agent.dbConfig.DB.GetOperationTimeByType(context.Background(), oper)
				if err != nil {
					log.Printf("Can't get execution time by operation type: %v", err)
					return
				}

				timer := time.NewTimer(time.Duration(time_for_oper) * time.Second)

				go simpleComputer(&exprMsg, digit1, digit2, oper, timer, agent.simpleComputers)

				agent.number_of_active_calculations++
				if agent.number_of_active_calculations == agent.number_of_parallel_calculations {
					agent.status = "sleeping"
					err := agent.dbConfig.DB.UpdateAgentStatus(
						context.Background(),
						database.UpdateAgentStatusParams{
							Status: "sleeping",
							ID:     agent.agentID,
						})
					if err != nil {
						log.Printf("Can't update agent status: %v", err)
						return
					}
				} else if agent.status != "running" {
					agent.status = "running"
					err := agent.dbConfig.DB.UpdateAgentStatus(
						context.Background(),
						database.UpdateAgentStatusParams{
							Status: "running",
							ID:     agent.agentID,
						})
					if err != nil {
						log.Printf("Can't update agent status: %v", err)
						return
					}
				}
			}(msgFromAgentAgregator)
		}
	}()

	for {
		result := <-agent.simpleComputers
		go func(result *ExpressionMessage) {
			log.Println("Agent consume message from computers", result)
			err := agent.PublishMessage(result)
			if err != nil {
				agent.Reconnect()
				err := agent.PublishMessage(result)
				if err != nil {
					return
				}
			}
			agent.number_of_active_calculations--
			if agent.number_of_active_calculations == 0 {
				agent.status = "waiting"
				err := agent.dbConfig.DB.UpdateAgentStatus(
					context.Background(),
					database.UpdateAgentStatusParams{
						Status: "waiting",
						ID:     agent.agentID,
					})
				if err != nil {
					log.Printf("Can't update agent status: %v", err)
					return
				}
			}
		}(result)
	}

}
