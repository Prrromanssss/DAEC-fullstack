package config

import (
	"log"
	"sync"

	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	Conn   *amqp.Connection
	Ch     *amqp.Channel
	Queues map[string]amqp.Queue
	mu     *sync.Mutex
}

type AMQPConsumer struct {
	Queue    amqp.Queue
	Messages <-chan amqp.Delivery
}

type AMQPProducer struct {
	Queue amqp.Queue
	Ch    *amqp.Channel
}

func NewAMQPConfig(amqpUrl string) (*AMQPConfig, error) {
	conn, err := amqp.Dial(amqpUrl)
	if err != nil {
		log.Printf("Can't connect to RabbitMQ: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to RabbitMQ instance")

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Can't create a channel from RabbitMQ: %v", err)
		return nil, err
	}

	return &AMQPConfig{
		Conn:   conn,
		Ch:     ch,
		Queues: make(map[string]amqp.Queue),
	}, nil
}

func NewAMQProducer(config AMQPConfig, queueName string) (*AMQPProducer, error) {
	queue, err := config.Ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Can't create a RabbitMQ queue: %v", err)
		return nil, err
	}
	config.mu.Lock()
	if _, ok := config.Queues[queueName]; !ok {
		config.Queues[queueName] = queue
	}
	config.mu.Unlock()
	return &AMQPProducer{
		Queue: queue,
		Ch:    config.Ch,
	}, nil
}

func NewAMQPConsumer(config AMQPConfig, queueName string) (*AMQPConsumer, error) {
	queue, err := config.Ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Can't create a RabbitMQ queue: %v", err)
		return nil, err
	}
	config.mu.Lock()
	if _, ok := config.Queues[queueName]; !ok {
		config.Queues[queueName] = queue
	}
	config.mu.Unlock()
	msgs, err := config.Ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Can't create a channel to consume messages from RabbitMQ: %v", err)
		return nil, err
	}
	return &AMQPConsumer{
		Queue:    queue,
		Messages: msgs,
	}, nil
}
