package rabbitmq

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/streadway/amqp"
)

type AMQPProducer struct {
	log     *slog.Logger
	amqpCfg *AMQPConfig
	Queue   amqp.Queue
	channel *amqp.Channel
}

func NewAMQPProducer(log *slog.Logger, amqpCfg *AMQPConfig, queueName string) (*AMQPProducer, error) {
	chProd, err := amqpCfg.conn.Channel()
	if err != nil {
		log.Error("can't create a channel from RabbitMQ", sl.Err(err))
		return nil, err
	}

	queue, err := chProd.QueueDeclare(
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

	return &AMQPProducer{
		log:     log,
		amqpCfg: amqpCfg,
		Queue:   queue,
		channel: chProd,
	}, nil
}

func (ap *AMQPProducer) PublishMessage(msg *messages.ExpressionMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return errors.New("failed to encode message to JSON")
	}

	err = ap.channel.Publish(
		"",
		ap.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)

	if err != nil {
		ap.log.Error("can't publish message to queue", slog.String("queue", ap.Queue.Name), sl.Err(err))
		return errors.New("can't publish message to queue")
	}
	ap.log.Info("publishing message to queue", slog.String("queue", ap.Queue.Name))

	return nil
}

func (ap *AMQPProducer) Reconnect() (*AMQPProducer, error) {
	ap.channel.Close()

	return NewAMQPProducer(ap.log, ap.amqpCfg, ap.Queue.Name)
}
