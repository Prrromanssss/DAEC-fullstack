package config

import (
	"log"

	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

func NewAMQPConfig(amqpUrl string) *AMQPConfig {
	conn, err := amqp.Dial(amqpUrl)
	if err != nil {
		log.Fatalf("Can't connect to RabbitMQ: %v", err)
	}

	log.Println("Successfully connected to RabbitMQ instance")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Can't create a channel from RabbitMQ: %v", err)
	}
	return &AMQPConfig{
		Conn: conn,
		Ch:   ch,
	}
}
