package rabbitmq

import (
	"log/slog"

	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/sl"
	"github.com/streadway/amqp"
)

type AMQPConsumer struct {
	log      *slog.Logger
	amqpCfg  *AMQPConfig
	Queue    amqp.Queue
	Channel  *amqp.Channel
	Messages <-chan amqp.Delivery
}

// NewAMQPConsumer creates new Consumer for AMQP protocol.
func NewAMQPConsumer(
	log *slog.Logger,
	amqpCfg *AMQPConfig,
	queueName string,
) (*AMQPConsumer, error) {
	chCons, err := amqpCfg.conn.Channel()
	if err != nil {
		log.Error("can't create a channel from RabbitMQ", sl.Err(err))
		return nil, err
	}
	queue, err := chCons.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("can't create a RabbitMQ queue", sl.Err(err))
		return nil, err
	}
	msgs, err := chCons.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("can't create a channel to consume messages from RabbitMQ", sl.Err(err))
		return nil, err
	}
	return &AMQPConsumer{
		log:      log,
		amqpCfg:  amqpCfg,
		Queue:    queue,
		Channel:  chCons,
		Messages: msgs,
	}, nil
}

// GetMessages returns messages from the Consumer channel.
func (ac *AMQPConsumer) GetMessages() <-chan amqp.Delivery {
	return ac.Messages
}

// Close closes Consumer channel.
func (ac *AMQPConsumer) Close() {
	ac.Channel.Close()
}
