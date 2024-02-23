package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Prrromanssss/DAEE-fullstack/config"
	"github.com/Prrromanssss/DAEE-fullstack/internal/database"

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
	inactiveTime                    int32
	status                          string
	amqpConfig                      *config.AMQPConfig
	dbConfig                        *config.DBConfig
	simpleComputers                 chan *ExpressionMessage
	amqpProducer                    *config.AMQPProducer
	amqpConsumer                    *config.AMQPConsumer
	mu                              *sync.Mutex
	kill                            chan struct{}
	allExpressions                  map[uuid.UUID]struct{}
}

func NewAgent(
	rabbitMQURL string,
	dbCfg *config.DBConfig,
	titleForExpressionQueue,
	titleForResultAndPingQueue string,
	numberOfParallelCalculations int32,
	inactiveTime int32,
) (*Agent, error) {
	amqpCfg, err := config.NewAMQPConfig(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPConfig for Agent: %v", err)
	}
	amqpProd, err := config.NewAMQPProducer(amqpCfg, titleForResultAndPingQueue)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPProducer for Agent: %v", err)
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, titleForExpressionQueue)
	if err != nil {
		return nil, fmt.Errorf("can't create NewAMQPConsumer for Agent: %v", err)
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
		inactiveTime:                    inactiveTime,
		status:                          string(agent.Status),
		amqpConfig:                      amqpCfg,
		dbConfig:                        dbCfg,
		simpleComputers:                 make(chan *ExpressionMessage),
		amqpProducer:                    amqpProd,
		amqpConsumer:                    amqpCons,
		mu:                              &sync.Mutex{},
		kill:                            make(chan struct{}),
		allExpressions:                  make(map[uuid.UUID]struct{}),
	}, nil
}

func (a *Agent) Reconnect() error {
	a.amqpConfig.ChannelForConsume.Close()
	a.amqpConfig.ChannelForProduce.Close()
	a.amqpConfig.Conn.Close()
	var err error
	a.amqpConfig, a.amqpProducer, a.amqpConsumer, err = a.ConnectToRabbitMQ()
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) ConnectToRabbitMQ() (
	*config.AMQPConfig,
	*config.AMQPProducer,
	*config.AMQPConsumer,
	error,
) {
	amqpCfg, err := config.NewAMQPConfig(a.rabbitMQURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't create NewAMQPConfig: %v", err)
	}
	amqpProd, err := config.NewAMQPProducer(amqpCfg, a.titleForResultAndPingQueue)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't create NewAMQPProducer: %v", err)
	}
	amqpCons, err := config.NewAMQPConsumer(amqpCfg, a.titleForExpressionQueue)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't create NewAMQPConsumer: %v", err)
	}
	return amqpCfg, amqpProd, amqpCons, nil
}

func (a *Agent) PublishMessage(msg *ExpressionMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return errors.New("failed to encode message to JSON")
	}

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

func (a *Agent) Terminate() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := a.dbConfig.DB.UpdateAgentStatus(context.Background(), database.UpdateAgentStatusParams{
		ID:     a.agentID,
		Status: "terminated",
	})
	if err != nil {
		return fmt.Errorf("can't terminate agent(id = %s): %v", a.agentID, err)
	}
	return nil
}

func (a *Agent) DecrementActiveComputers() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.number_of_active_calculations--
	if a.number_of_active_calculations == 0 {
		a.status = "waiting"
		err := a.dbConfig.DB.UpdateAgentStatus(
			context.Background(),
			database.UpdateAgentStatusParams{
				Status: "waiting",
				ID:     a.agentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v", err)
		}
	} else {
		a.status = "running"
		err := a.dbConfig.DB.UpdateAgentStatus(
			context.Background(),
			database.UpdateAgentStatusParams{
				Status: "running",
				ID:     a.agentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v", err)
		}
	}
	return nil
}

func (a *Agent) ChangeAgentStatusToRunningOrSleeping() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.number_of_active_calculations == a.number_of_parallel_calculations {
		a.status = "sleeping"
		err := a.dbConfig.DB.UpdateAgentStatus(
			context.Background(),
			database.UpdateAgentStatusParams{
				Status: "sleeping",
				ID:     a.agentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v", err)
		}
	} else if a.status != "running" {
		a.status = "running"
		err := a.dbConfig.DB.UpdateAgentStatus(
			context.Background(),
			database.UpdateAgentStatusParams{
				Status: "running",
				ID:     a.agentID,
			})
		if err != nil {
			return fmt.Errorf("can't update agent status: %v", err)
		}
	}
	return nil
}

func (a *Agent) RunSimpleComputer(exprMsg *ExpressionMessage) error {
	tokenSplit := strings.Split(exprMsg.Token, " ")
	if len(tokenSplit) != 3 {
		return fmt.Errorf("invalid token")
	}
	oper := tokenSplit[2]
	if !(oper == "+" || oper == "-" || oper == "/" || oper == "*") {
		return fmt.Errorf("operation in token doesn't match any of these +, -, /, *")
	}

	digit1, err := strconv.Atoi(tokenSplit[0])
	if err != nil {
		return fmt.Errorf("can't convert int to str: %v", err)
	}
	digit2, err := strconv.Atoi(tokenSplit[1])
	if err != nil {
		return fmt.Errorf("can't convert int to str: %v", err)
	}

	time_for_oper, err := a.dbConfig.DB.GetOperationTimeByType(context.Background(), oper)
	if err != nil {
		return fmt.Errorf("can't get execution time by operation type: %v", err)
	}

	timer := time.NewTimer(time.Duration(time_for_oper) * time.Second)

	go simpleComputer(exprMsg, digit1, digit2, oper, timer, a.simpleComputers)

	a.mu.Lock()
	a.number_of_active_calculations++
	a.mu.Unlock()

	return nil
}

func (a *Agent) ChangeExpressionStatus(exprID uuid.UUID, newStatus string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := a.dbConfig.DB.UpdateExpressionStatus(
		context.Background(),
		database.UpdateExpressionStatusParams{
			ID:     exprID,
			Status: database.ExpressionStatus(newStatus),
		})
	if err != nil {
		return fmt.Errorf("can't update expression status: %v", err)
	}
	return nil
}

func (a *Agent) MakeExpressionsTerminated() {
	log.Println(a.allExpressions)

	for exprID := range a.allExpressions {
		err := a.ChangeExpressionStatus(exprID, "terminated")
		if err != nil {
			log.Printf("Can't make expressions terminated 'cause agent is down: %v", err)
		}
	}
}

func (a *Agent) ConsumeMessageFromComputers(result *ExpressionMessage) {
	log.Println("Agent consume message from computers", result)
	err := a.PublishMessage(result)
	if err != nil {
		err := a.Reconnect()
		if err != nil {
			log.Printf("Agent Error: %v", err)
			a.kill <- struct{}{}
			return
		}
		err = a.PublishMessage(result)
		if err != nil {
			log.Printf("Agent Error: %v", err)
			a.kill <- struct{}{}
			return
		}
	}

	a.mu.Lock()
	delete(a.allExpressions, result.ExpressionID)
	a.mu.Unlock()

	err = a.DecrementActiveComputers()
	if err != nil {
		log.Printf("Agent Error: %v", err)
		a.kill <- struct{}{}
		return
	}
}

func (a *Agent) ConsumeMessageFromAgentAgregator(msgFromAgentAgregator amqp.Delivery) {
	log.Println("Agent consume msg from agent agregator", msgFromAgentAgregator.Body)
	var exprMsg ExpressionMessage
	if err := json.Unmarshal(msgFromAgentAgregator.Body, &exprMsg); err != nil {
		log.Printf("Agent Error: Failed to parse JSON: %v", err)
		a.kill <- struct{}{}
		return
	}

	a.mu.Lock()
	if a.number_of_active_calculations >= a.number_of_parallel_calculations {
		err := msgFromAgentAgregator.Nack(false, true)
		if err != nil {
			log.Printf("Agent Error: %v", err)
			a.kill <- struct{}{}
			a.mu.Unlock()
			return
		}
		a.mu.Unlock()
		return
	}
	a.mu.Unlock()

	err := msgFromAgentAgregator.Ack(false)
	if err != nil {
		log.Printf("Agent Error: Error acknowledging message: %v", err)
		a.kill <- struct{}{}
		return
	}

	err = a.ChangeExpressionStatus(exprMsg.ExpressionID, "computing")
	if err != nil {
		log.Printf("Agent Error: %v", err)
		a.kill <- struct{}{}
		return
	}

	a.mu.Lock()
	a.allExpressions[exprMsg.ExpressionID] = struct{}{}
	a.mu.Unlock()

	err = a.RunSimpleComputer(&exprMsg)
	if err != nil {
		log.Printf("Agent Error: %v", err)
		a.kill <- struct{}{}
		return
	}

	err = a.ChangeAgentStatusToRunningOrSleeping()
	if err != nil {
		log.Printf("Agent Error: %v", err)
		a.kill <- struct{}{}
		return
	}
}

func AgentService(agent *Agent) {
	defer agent.amqpConfig.Conn.Close()
	defer agent.amqpConfig.ChannelForConsume.Close()
	defer agent.amqpConfig.ChannelForProduce.Close()
	defer agent.MakeExpressionsTerminated()

	go func() {
		for msgFromAgentAgregator := range agent.amqpConsumer.Messages {
			go agent.ConsumeMessageFromAgentAgregator(msgFromAgentAgregator)
		}
	}()

	ticker := time.NewTicker(time.Duration(agent.inactiveTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case result := <-agent.simpleComputers:
			go agent.ConsumeMessageFromComputers(result)
		case <-agent.kill:
			err := agent.Terminate()
			if err != nil {
				log.Printf("Agent Error: %v", err)
			}
			return
		case <-ticker.C:
			exprMsg := ExpressionMessage{
				IsPing:  true,
				AgentID: agent.agentID,
			}
			err := agent.PublishMessage(&exprMsg)
			if err != nil {
				log.Printf("Agent Error: Can't send ping: %v", err)
			}
		}
	}
}
