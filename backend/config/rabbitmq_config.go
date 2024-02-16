package config

import (
	"log"

	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	Conn              *amqp.Connection
	ChannelForProduce *amqp.Channel
	ChannelForConsume *amqp.Channel
}

type AMQPConsumer struct {
	Queue             amqp.Queue
	Messages          <-chan amqp.Delivery
	ChannelForConsume *amqp.Channel
}

type AMQPProducer struct {
	Queue             amqp.Queue
	ChannelForProduce *amqp.Channel
}

func NewAMQPConfig(amqpUrl string) (*AMQPConfig, error) {
	conn, err := amqp.Dial(amqpUrl)
	if err != nil {
		log.Fatalf("Can't connect to RabbitMQ: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to RabbitMQ instance")

	chProd, err := conn.Channel()
	if err != nil {
		log.Fatalf("Can't create a channel from RabbitMQ: %v", err)
		return nil, err
	}
	chCons, err := conn.Channel()
	if err != nil {
		log.Fatalf("Can't create a channel from RabbitMQ: %v", err)
		return nil, err
	}

	return &AMQPConfig{
		Conn:              conn,
		ChannelForProduce: chProd,
		ChannelForConsume: chCons,
	}, nil
}

func NewAMQPProducer(amqpCfg *AMQPConfig, queueName string) (*AMQPProducer, error) {
	queue, err := amqpCfg.ChannelForProduce.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Can't create a RabbitMQ queue: %v", err)
		return nil, err
	}
	return &AMQPProducer{
		Queue:             queue,
		ChannelForProduce: amqpCfg.ChannelForProduce,
	}, nil
}

func NewAMQPConsumer(amqpCfg *AMQPConfig, queueName string) (*AMQPConsumer, error) {
	queue, err := amqpCfg.ChannelForConsume.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Can't create a RabbitMQ queue: %v", err)
		return nil, err
	}
	msgs, err := amqpCfg.ChannelForConsume.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Can't create a channel to consume messages from RabbitMQ: %v", err)
		return nil, err
	}
	return &AMQPConsumer{
		Queue:             queue,
		Messages:          msgs,
		ChannelForConsume: amqpCfg.ChannelForConsume,
	}, nil
}
